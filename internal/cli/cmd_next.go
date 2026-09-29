package cli

import (
	"fmt"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/AliyuYahaya/Ajiya/internal/check"
	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/gitx"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// runNext lists the tickets that can start now: earliest milestone first
// (unscheduled last), then those that unblock the most open tickets, then by ID.
func runNext(e *env, args []string) error {
	fs := newFlags("next")
	app := fs.String("app", "", "only tickets for this app")
	launchOnly := fs.Bool("launch", false, "only tickets required for the launch milestone")
	only := fs.String("milestone", "", "only tickets required for this milestone")
	asJSON := fs.Bool("json", false, "print JSON")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	if *launchOnly && *only != "" {
		return usageErr("give --launch or --milestone, not both")
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	if *app != "" {
		if err := pr.checkApp(*app); err != nil {
			return err
		}
	}
	if *launchOnly {
		if pr.milestone(config.LaunchMilestone) == nil {
			return refused("no launch target; set one with 'ajiya launch set <phase|ticket>'")
		}
		*only = config.LaunchMilestone
	} else if *only != "" {
		if err := pr.checkMilestone(*only); err != nil {
			return err
		}
	}
	s, err := pr.schedule()
	if err != nil {
		return err
	}
	tickets := pr.plan.Tickets()
	g := plan.NewGraph(tickets)

	// launch is the launch milestone's required set; filter is --milestone's.
	launch, filter := map[string]bool{}, map[string]bool{}
	if i := s.Index(config.LaunchMilestone); i >= 0 {
		launch = s.Required[i]
	}
	if *only != "" {
		filter = s.Required[s.Index(*only)]
	}
	// rank is the position of a ticket's earliest milestone; unscheduled last.
	rank := func(id string) int {
		if m, ok := s.Of[id]; ok {
			return s.Index(m)
		}
		return len(s.Milestones)
	}

	// unblocks counts the open tickets waiting, directly or not, on each ticket.
	unblocks := map[string]int{}
	var ready []*plan.Ticket
	for _, t := range tickets {
		if !canStart(pr, t) || (*app != "" && t.App != *app) || (*only != "" && !filter[t.ID]) {
			continue
		}
		for _, id := range g.Downstream(t.ID) {
			if d := pr.plan.Ticket(id); d != nil && !d.Status.Closed() {
				unblocks[t.ID]++
			}
		}
		ready = append(ready, t)
	}
	sort.SliceStable(ready, func(i, j int) bool {
		a, b := ready[i], ready[j]
		if ra, rb := rank(a.ID), rank(b.ID); ra != rb {
			return ra < rb
		}
		if unblocks[a.ID] != unblocks[b.ID] {
			return unblocks[a.ID] > unblocks[b.ID]
		}
		return plan.LessID(a.ID, b.ID)
	})

	if *asJSON {
		type item struct {
			jsonTicket
			Launch    bool   `json:"launch"`    // required for the launch milestone
			Milestone string `json:"milestone"` // earliest milestone; "" when unscheduled
			Unblocks  int    `json:"unblocks"`
		}
		items := []item{}
		for _, t := range ready {
			items = append(items, item{toJSON(t), launch[t.ID], s.Of[t.ID], unblocks[t.ID]})
		}
		return writeJSON(e, items)
	}
	if len(ready) == 0 {
		fmt.Fprintln(e.stdout, "Nothing can start now.")
		return nil
	}
	w := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
	for _, t := range ready {
		extra := s.Of[t.ID]
		if n := unblocks[t.ID]; n > 0 {
			extra = join(extra, fmt.Sprintf("unblocks %d", n))
		}
		if t.Status.State == plan.InProgress {
			extra = join(extra, "in progress")
		}
		if t.Status.Human != "" {
			extra = join(extra, "needs a human: "+t.Status.Human)
		}
		line := fmt.Sprintf("%s %s\t%s\t%s", t.Status.Mark(), t.ID, t.App, t.Title)
		if extra != "" { // no trailing tab, so the title is not padded with spaces
			line += "\t" + extra
		}
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func join(a, b string) string {
	if a == "" {
		return b
	}
	return a + " · " + b
}

func runCheck(e *env, args []string) error {
	fs := newFlags("check")
	strict := fs.Bool("strict", false, "fail on warnings too")
	asJSON := fs.Bool("json", false, "print the findings as a JSON list")
	commits := fs.String("commits", "", "also check the commit rule on a revision range, such as origin/main..HEAD")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	findings := check.Run(pr.cfg, pr.plan)
	if isSet(fs, "commits") {
		if *commits == "" {
			return usageErr("--commits needs a revision range, such as origin/main..HEAD")
		}
		log, err := gitx.LogRange(pr.cfg.Root, *commits)
		if err != nil {
			return usageErr("--commits %s: %v", *commits, err)
		}
		findings = append(findings, check.Commits(pr.cfg, pr.plan, log)...)
	}
	history, err := check.History(pr.cfg, pr.plan, time.Now())
	if err != nil {
		return err
	}
	findings = append(findings, history...)
	check.Sort(findings)
	errs, warns := check.Count(findings)
	if *asJSON {
		if findings == nil {
			findings = []check.Finding{}
		}
		if err := writeJSON(e, findings); err != nil {
			return err
		}
	} else {
		for _, f := range findings {
			fmt.Fprintln(e.stdout, f)
		}
		if len(findings) == 0 {
			fmt.Fprintln(e.stdout, "No problems found.")
		} else {
			fmt.Fprintf(e.stdout, "%d error(s), %d warning(s).\n", errs, warns)
		}
	}
	if errs > 0 || (*strict && warns > 0) {
		return silent(ExitRefused)
	}
	return nil
}
