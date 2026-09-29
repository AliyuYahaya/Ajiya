package cli

import (
	"fmt"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// orderSchedule works out the milestones for ordering phases. Display must
// not fail, so a milestone with a missing target gives no schedule; 'ajiya
// check' reports that target.
func (pr *project) orderSchedule() *plan.Schedule {
	var ms []plan.Milestone
	for _, m := range pr.cfg.MilestoneList() {
		ms = append(ms, plan.Milestone{Name: m.Name, Targets: m.Targets})
	}
	s, err := pr.plan.Schedule(ms)
	if err != nil {
		return nil
	}
	return s
}

// displayPhases returns the phases in display order: [phases] order first,
// then the rest in suggested order.
func (pr *project) displayPhases() []*plan.Phase {
	return pr.plan.DisplayOrder(pr.cfg.Phases.Order, pr.orderSchedule())
}

func runPhaseMove(e *env, args []string) error {
	fs := newFlags("phase move")
	before := fs.String("before", "", "put it just before this phase")
	after := fs.String("after", "", "put it just after this phase")
	first := fs.Bool("first", false, "put it first")
	last := fs.Bool("last", false, "put it last")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	n := 0
	for _, set := range []bool{isSet(fs, "before"), isSet(fs, "after"), *first, *last} {
		if set {
			n++
		}
	}
	if n != 1 {
		return usageErr("give exactly one of --before <slug>, --after <slug>, --first or --last")
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	slug := pos[0]
	if _, err := pr.phase(slug); err != nil {
		return err
	}
	anchor := *before + *after
	if isSet(fs, "before") || isSet(fs, "after") {
		if anchor == slug {
			return usageErr("cannot move %s relative to itself", slug)
		}
		if _, err := pr.phase(anchor); err != nil {
			return err
		}
	}

	// Start from what is shown now: the list, then every phase left out of
	// it. Listed slugs that are not phases stay where they are for 'ajiya
	// check' to report.
	order := slices.Clone(pr.cfg.Phases.Order)
	for _, ph := range pr.displayPhases() {
		if !slices.Contains(order, ph.Slug) {
			order = append(order, ph.Slug)
		}
	}
	order = slices.DeleteFunc(order, func(s string) bool { return s == slug })
	var at int
	switch {
	case *last:
		at = len(order)
	case *first:
		at = 0
	case isSet(fs, "before"):
		at = slices.Index(order, anchor)
	case isSet(fs, "after"):
		at = slices.Index(order, anchor) + 1
	}
	order = slices.Insert(order, at, slug)
	if _, err := config.SetPhaseOrder(pr.cfg.Root, order); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "Phase order: %s\n", strings.Join(order, ", "))
	return nil
}

func runPhaseList(e *env, args []string) error {
	fs := newFlags("phase list")
	asJSON := fs.Bool("json", false, "print JSON")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	type item struct {
		Slug      string `json:"slug"`
		Title     string `json:"title"`
		Done      int    `json:"done"` // closed: done or dropped
		Total     int    `json:"total"`
		Milestone string `json:"milestone"` // earliest milestone of its tickets; empty if none
	}
	sched := pr.orderSchedule()
	items := []item{}
	for _, ph := range pr.plan.DisplayOrder(pr.cfg.Phases.Order, sched) {
		it := item{Slug: ph.Slug, Title: ph.Title, Total: len(ph.Tickets)}
		for _, t := range ph.Tickets {
			if t.Status.Closed() {
				it.Done++
			}
		}
		it.Milestone, _ = plan.PhaseMilestone(ph, sched)
		items = append(items, it)
	}
	if *asJSON {
		return writeJSON(e, items)
	}
	if len(items) == 0 {
		fmt.Fprintln(e.stdout, "No phases yet. Add one with 'ajiya phase add <slug> \"<goal>\"'.")
		return nil
	}
	w := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PHASE\tTITLE\tDONE\tMILESTONE")
	for _, it := range items {
		m := it.Milestone
		if m == "" {
			m = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%d/%d\t%s\n", it.Slug, it.Title, it.Done, it.Total, m)
	}
	return w.Flush()
}

func runPhaseOrder(e *env, args []string) error {
	fs := newFlags("phase order")
	suggest := fs.Bool("suggest", false, "print the order worked out from the dependencies")
	apply := fs.Bool("apply", false, "write that order to [phases] order in ajiya.toml")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	if *suggest == *apply {
		return usageErr("give --suggest or --apply")
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	order, cycles := pr.plan.SuggestOrder(pr.orderSchedule())
	if len(order) == 0 {
		return refused("no phases to order; add one with 'ajiya phase add <slug> \"<goal>\"'")
	}
	if *apply {
		if _, err := config.SetPhaseOrder(pr.cfg.Root, order); err != nil {
			return err
		}
		fmt.Fprintf(e.stdout, "Phase order: %s\n", strings.Join(order, ", "))
	} else {
		for i, slug := range order {
			fmt.Fprintf(e.stdout, "%d. %s\n", i+1, slug)
		}
	}
	for _, c := range cycles {
		fmt.Fprintf(e.stdout, "Info: phases %s wait on each other; they are kept together, in the order they were created.\n", strings.Join(c, ", "))
	}
	return nil
}
