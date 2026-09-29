package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// Milestones are named gates kept in ajiya.toml, in order. The older [launch]
// target counts as one milestone named launch; the commands that change the
// list convert it to a [[milestones]] table first.

// schedule works out every milestone against the plan.
func (pr *project) schedule() (*plan.Schedule, error) {
	ms := pr.cfg.MilestoneList()
	pms := make([]plan.Milestone, len(ms))
	for i, m := range ms {
		pms[i] = plan.Milestone{Name: m.Name, Targets: m.Targets}
	}
	s, err := pr.plan.Schedule(pms)
	if err != nil {
		return nil, refused("%v; fix it in %s or with 'ajiya milestone' commands", err, config.FileName)
	}
	return s, nil
}

// milestone returns the milestone called name, or nil.
func (pr *project) milestone(name string) *config.Milestone {
	for _, m := range pr.cfg.MilestoneList() {
		if m.Name == name {
			return &m
		}
	}
	return nil
}

// checkMilestone refuses a name that is not a milestone.
func (pr *project) checkMilestone(name string) error {
	if pr.milestone(name) != nil {
		return nil
	}
	var names []string
	for _, m := range pr.cfg.MilestoneList() {
		names = append(names, m.Name)
	}
	if len(names) == 0 {
		return refused("no milestone %q; there are none yet: add one with 'ajiya milestone add <name> --targets <phases|tickets>'", name)
	}
	return refused("no milestone %q; milestones are %s", name, strings.Join(names, ", "))
}

// canStart reports whether a ticket is open, not blocked and waits on nothing.
func canStart(pr *project, t *plan.Ticket) bool {
	return !t.Status.Closed() && t.Status.Blocked == "" && len(waitingOn(pr, t)) == 0
}

// editConfig applies an edit to ajiya.toml, validates it and writes it.
func editConfig(pr *project, edit func(string) (string, error)) error {
	var editErr error
	err := writeConfig(pr, func(text string) (string, bool) {
		out, err := edit(text)
		if err != nil {
			editErr = err
			return text, false
		}
		return out, true
	})
	if editErr != nil {
		return refused("%v", editErr)
	}
	return err
}

// splitTargets splits a comma separated list, dropping empty items.
func splitTargets(s string) []string {
	var out []string
	for _, f := range strings.Split(s, ",") {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}

// placeFlags adds --before and --after and checks that at most one is set.
type placeFlags struct{ before, after *string }

func (p placeFlags) check(pr *project, need bool) error {
	switch {
	case *p.before != "" && *p.after != "":
		return usageErr("give --before or --after, not both")
	case need && *p.before == "" && *p.after == "":
		return usageErr("give --before <milestone> or --after <milestone>")
	}
	for _, n := range []string{*p.before, *p.after} {
		if n != "" {
			if err := pr.checkMilestone(n); err != nil {
				return err
			}
		}
	}
	return nil
}

func milestoneOrder(pr *project) string {
	var names []string
	for _, m := range pr.cfg.MilestoneList() {
		names = append(names, m.Name)
	}
	return strings.Join(names, ", ")
}

func runMilestoneAdd(e *env, args []string) error {
	fs := newFlags("milestone add")
	targets := fs.String("targets", "", "phase slugs or ticket IDs, comma separated")
	place := placeFlags{fs.String("before", "", "put it before this milestone"), fs.String("after", "", "put it after this milestone")}
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	name, list := pos[0], splitTargets(*targets)
	if err := config.CheckMilestoneName(name); err != nil {
		return usageErr("%v", err)
	}
	if len(list) == 0 {
		return usageErr("--targets is required: phase slugs or ticket IDs, comma separated")
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	if pr.milestone(name) != nil {
		return refused("milestone %q already exists; see 'ajiya milestone list'", name)
	}
	if err := place.check(pr, false); err != nil {
		return err
	}
	for _, t := range list {
		if _, _, err := pr.plan.Required(t); err != nil {
			return refused("%v; give phase slugs or ticket IDs", err)
		}
	}
	converted := pr.cfg.Launch.Target != ""
	m := config.Milestone{Name: name, Targets: list}
	if err := editConfig(pr, func(text string) (string, error) {
		return config.AddMilestoneText(text, m, *place.before, *place.after)
	}); err != nil {
		return err
	}
	if converted {
		fmt.Fprintf(e.stdout, "Moved the [launch] target into [[milestones]] as %q.\n", config.LaunchMilestone)
	}
	fmt.Fprintf(e.stdout, "Added milestone %s: %s.\nOrder: %s.\n", name, strings.Join(list, ", "), milestoneOrder(pr))
	return nil
}

func runMilestoneMove(e *env, args []string) error {
	fs := newFlags("milestone move")
	place := placeFlags{fs.String("before", "", "put it before this milestone"), fs.String("after", "", "put it after this milestone")}
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	name := pos[0]
	if err := pr.checkMilestone(name); err != nil {
		return err
	}
	if err := place.check(pr, true); err != nil {
		return err
	}
	if *place.before == name || *place.after == name {
		return usageErr("cannot move milestone %s relative to itself", name)
	}
	if err := editConfig(pr, func(text string) (string, error) {
		return config.MoveMilestoneText(text, name, *place.before, *place.after)
	}); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "Moved milestone %s.\nOrder: %s.\n", name, milestoneOrder(pr))
	return nil
}

func runMilestoneRemove(e *env, args []string) error {
	pos, err := parse(newFlags("milestone remove"), args, 1)
	if err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	name := pos[0]
	if err := pr.checkMilestone(name); err != nil {
		return err
	}
	if err := editConfig(pr, func(text string) (string, error) { return config.RemoveMilestoneText(text, name) }); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "Removed milestone %s; its tickets are unchanged.\n", name)
	return nil
}

// milestoneStats counts one milestone's required tickets.
type milestoneStats struct {
	Name     string   `json:"name"`
	Targets  []string `json:"targets"`
	Required int      `json:"required"`
	Closed   int      `json:"closed"`
	Percent  int      `json:"percent"`
	CanStart int      `json:"can_start"` // open required tickets that can start now
}

func statsOf(pr *project, m plan.Milestone, req map[string]bool) milestoneStats {
	st := milestoneStats{Name: m.Name, Targets: m.Targets, Required: len(req)}
	if st.Targets == nil {
		st.Targets = []string{}
	}
	for _, t := range pr.plan.Tickets() {
		switch {
		case !req[t.ID]:
		case t.Status.Closed():
			st.Closed++
		case canStart(pr, t):
			st.CanStart++
		}
	}
	st.Percent = percent(st.Closed, st.Required)
	return st
}

func runMilestoneList(e *env, args []string) error {
	fs := newFlags("milestone list")
	asJSON := fs.Bool("json", false, "print JSON")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	s, err := pr.schedule()
	if err != nil {
		return err
	}
	list := []milestoneStats{}
	for i, m := range s.Milestones {
		list = append(list, statsOf(pr, m, s.Required[i]))
	}
	if *asJSON {
		return writeJSON(e, list)
	}
	if len(list) == 0 {
		fmt.Fprintln(e.stdout, "No milestones. Add one with 'ajiya milestone add <name> --targets <phases|tickets>'.")
		return nil
	}
	w := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
	for i, st := range list {
		fmt.Fprintf(w, "%d. %s\t%s\t%d of %d closed (%d%%)\t%d can start\n",
			i+1, st.Name, strings.Join(st.Targets, ", "), st.Closed, st.Required, st.Percent, st.CanStart)
	}
	return w.Flush()
}

func runMilestoneShow(e *env, args []string) error {
	fs := newFlags("milestone show")
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	name := pos[0]
	if err := pr.checkMilestone(name); err != nil {
		return err
	}
	s, err := pr.schedule()
	if err != nil {
		return err
	}
	i := s.Index(name)
	m, req := s.Milestones[i], s.Required[i]
	st := statsOf(pr, m, req)

	type target struct {
		Target string `json:"target"`
		Kind   string `json:"kind"` // phase or ticket
	}
	type openTicket struct {
		jsonTicket
		WaitingOn []string `json:"waiting_on"` // open direct dependencies
		CanStart  bool     `json:"can_start"`
	}
	out := struct {
		milestoneStats
		Position   int          `json:"position"` // 1 for the first milestone
		Kinds      []target     `json:"target_kinds"`
		Open       []openTicket `json:"open"`
		NeedsHuman []string     `json:"needs_human"` // IDs of open required tickets that need a person
	}{milestoneStats: st, Position: i + 1, Kinds: []target{}, Open: []openTicket{}, NeedsHuman: []string{}}
	for _, t := range m.Targets {
		_, kind, _ := pr.plan.Required(t)
		out.Kinds = append(out.Kinds, target{t, kind})
	}
	for _, t := range pr.plan.Tickets() {
		if !req[t.ID] || t.Status.Closed() {
			continue
		}
		o := openTicket{jsonTicket: toJSON(t), WaitingOn: []string{}, CanStart: canStart(pr, t)}
		for _, d := range waitingOn(pr, t) {
			o.WaitingOn = append(o.WaitingOn, d.ID)
		}
		out.Open = append(out.Open, o)
		if t.Status.Human != "" {
			out.NeedsHuman = append(out.NeedsHuman, t.ID)
		}
	}
	if *asJSON {
		return writeJSON(e, out)
	}

	var kinds []string
	for _, k := range out.Kinds {
		kinds = append(kinds, fmt.Sprintf("%s (%s)", k.Target, k.Kind))
	}
	fmt.Fprintf(e.stdout, "Milestone: %s (%d of %d)\n", name, i+1, len(s.Milestones))
	fmt.Fprintf(e.stdout, "Targets: %s\n", strings.Join(kinds, ", "))
	fmt.Fprintf(e.stdout, "Required: %d of %d closed (%d%%)\n", st.Closed, st.Required, st.Percent)
	if len(out.Open) == 0 {
		fmt.Fprintln(e.stdout, "Nothing left: every required ticket is closed.")
		return nil
	}
	fmt.Fprintf(e.stdout, "Open and required: %d, %d can start now\n", len(out.Open), st.CanStart)
	w := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
	for _, o := range out.Open {
		t := pr.plan.Ticket(o.ID)
		extra := ""
		switch {
		case t.Status.Blocked != "":
			extra = "blocked: " + t.Status.Blocked
		case len(o.WaitingOn) > 0:
			extra = "waits on " + strings.Join(o.WaitingOn, ", ")
		default:
			extra = "can start"
		}
		if t.Status.Human != "" {
			extra = join(extra, "needs a human: "+t.Status.Human)
		}
		fmt.Fprintf(w, "  %s %s\t%s\t%s\n", t.Status.Mark(), t.ID, t.Title, extra)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if len(out.NeedsHuman) > 0 {
		fmt.Fprintf(e.stdout, "Needs a human: %s\n", strings.Join(out.NeedsHuman, ", "))
	}
	return nil
}
