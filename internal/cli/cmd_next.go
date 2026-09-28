package cli

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/AliyuYahaya/Ajiya/internal/check"
	"github.com/AliyuYahaya/Ajiya/internal/gitx"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

func runNext(e *env, args []string) error {
	fs := newFlags("next")
	app := fs.String("app", "", "only tickets for this app")
	if _, err := parse(fs, args, 0); err != nil {
		return err
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
	tickets := pr.plan.Tickets()
	g := plan.NewGraph(tickets)

	// unblocks counts the open tickets waiting, directly or not, on each ticket.
	unblocks := map[string]int{}
	var ready []*plan.Ticket
	for _, t := range tickets {
		if t.Status.Closed() || t.Status.Blocked != "" || (*app != "" && t.App != *app) || len(waitingOn(pr, t)) > 0 {
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
		if unblocks[a.ID] != unblocks[b.ID] {
			return unblocks[a.ID] > unblocks[b.ID]
		}
		return plan.LessID(a.ID, b.ID)
	})

	if len(ready) == 0 {
		fmt.Fprintln(e.stdout, "Nothing can start now.")
		return nil
	}
	w := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
	for _, t := range ready {
		extra := ""
		if n := unblocks[t.ID]; n > 0 {
			extra = fmt.Sprintf("unblocks %d", n)
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
		findings = append(findings, check.Commits(pr.plan, log)...)
	}
	for _, f := range findings {
		fmt.Fprintln(e.stdout, f)
	}
	errs, warns := check.Count(findings)
	if len(findings) == 0 {
		fmt.Fprintln(e.stdout, "No problems found.")
	} else {
		fmt.Fprintf(e.stdout, "%d error(s), %d warning(s).\n", errs, warns)
	}
	if errs > 0 || (*strict && warns > 0) {
		return silent(ExitRefused)
	}
	return nil
}
