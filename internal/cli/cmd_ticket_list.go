package cli

import (
	"fmt"
	"slices"
	"sort"
	"text/tabwriter"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// runTicketList lists tickets in plan order: phases in display order, then by
// ID within a phase. --state also takes the dashboard's ready and human filters.
func runTicketList(e *env, args []string) error {
	fs := newFlags("ticket list")
	phase := fs.String("phase", "", "only tickets in this phase")
	app := fs.String("app", "", "only tickets for this app")
	state := fs.String("state", "", "only tickets in this state: pending, in_progress, done, dropped, ready or human")
	milestone := fs.String("milestone", "", "only tickets required for this milestone")
	asJSON := fs.Bool("json", false, "print JSON")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	if isSet(fs, "state") && !slices.Contains([]string{"pending", "in_progress", "done", "dropped", "ready", "human"}, *state) {
		return usageErr("--state %q is not one of pending, in_progress, done, dropped, ready, human", *state)
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	if isSet(fs, "phase") {
		if _, err := pr.phase(*phase); err != nil {
			return err
		}
	}
	if isSet(fs, "app") {
		if err := pr.checkApp(*app); err != nil {
			return err
		}
	}
	var required map[string]bool
	if isSet(fs, "milestone") {
		if err := pr.checkMilestone(*milestone); err != nil {
			return err
		}
		s, err := pr.schedule()
		if err != nil {
			return err
		}
		required = s.Required[s.Index(*milestone)]
	}

	match := func(t *plan.Ticket) bool {
		open := !t.Status.Closed()
		switch {
		case isSet(fs, "app") && t.App != *app:
			return false
		case required != nil && !required[t.ID]:
			return false
		case *state == "ready":
			return canStart(pr, t)
		case *state == "human":
			return open && t.Status.Human != ""
		case *state != "":
			return stateNames[t.Status.State] == *state
		}
		return true
	}
	var list []*plan.Ticket
	for _, ph := range pr.displayPhases() {
		if isSet(fs, "phase") && ph.Slug != *phase {
			continue
		}
		ts := slices.Clone(ph.Tickets)
		sort.SliceStable(ts, func(i, j int) bool { return plan.LessID(ts[i].ID, ts[j].ID) })
		for _, t := range ts {
			if match(t) {
				list = append(list, t)
			}
		}
	}

	if *asJSON {
		items := []jsonTicket{}
		for _, t := range list {
			items = append(items, toJSON(t))
		}
		return writeJSON(e, items)
	}
	if len(list) == 0 {
		fmt.Fprintln(e.stdout, "No tickets match.")
		return nil
	}
	w := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
	for _, t := range list {
		fmt.Fprintf(w, "%s %s\t%s\t%s\t%s\n", t.Status.Mark(), t.ID, t.App, stateNames[t.Status.State], t.Title)
	}
	return w.Flush()
}
