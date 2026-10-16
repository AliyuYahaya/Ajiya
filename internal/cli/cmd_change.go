package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/AliyuYahaya/Ajiya/internal/gitx"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
	"github.com/AliyuYahaya/Ajiya/internal/view"
)

func (pr *project) change(id string) (*plan.Change, error) {
	if c := pr.plan.Change(id); c != nil {
		return c, nil
	}
	if pe := pr.plan.BrokenChanges[id]; pe != nil {
		return nil, refused("%s:%d: %s; see 'ajiya check'", plan.ChangePath(id), pe.Line, pe.Msg)
	}
	return nil, refused("no change %s; see 'ajiya change list'", id)
}

// openChange returns a change that is still open, for the commands that edit it.
func (pr *project) openChange(id string) (*plan.Change, error) {
	c, err := pr.change(id)
	if err != nil {
		return nil, err
	}
	if !c.Status.Open() {
		return nil, refused("%s is %s, so it can no longer be changed", c.ID, strings.ToLower(c.Status.String()))
	}
	return c, nil
}

// tagFlag reads a comma-separated list of key:value tags.
func tagFlag(flag, v string) (plan.Tags, error) {
	t, err := plan.ParseTags(strings.Join(plan.ParseTagList(v), " "))
	if len(plan.ParseTagList(v)) == 0 {
		return plan.Tags{}, nil
	}
	if err != nil {
		return plan.Tags{}, usageErr("--%s: %v (tags are key:value, for example tech:aws-rds)", flag, err)
	}
	return t, nil
}

func runChangeOpen(e *env, args []string) error {
	fs := newFlags("change open")
	replaces := fs.String("replaces", "", "tags of the work being replaced, comma separated, such as tech:aws-rds")
	with := fs.String("with", "", "tags of what replaces it, comma separated, such as tech:supabase")
	decision := fs.String("decision", "", "decision IDs this change would supersede, comma separated")
	paths := fs.String("paths", "", "file globs the change touches, comma separated, such as infra/rds/**")
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	title := pos[0]
	if title == "" {
		return usageErr("the title is empty")
	}
	if err := checkTexts("title", title); err != nil {
		return err
	}
	repl, err := tagFlag("replaces", *replaces)
	if err != nil {
		return err
	}
	by, err := tagFlag("with", *with)
	if err != nil {
		return err
	}
	decs := plan.ParseIDList(*decision)
	for _, d := range decs {
		if !plan.ValidID(plan.DecPrefix, d) {
			return usageErr("--decision: %q is not a decision ID such as DEC-0003", d)
		}
	}
	var globs []string
	for _, g := range plan.ParseTagList(*paths) {
		if err := plan.CheckTag(plan.TagPath, g); err != nil {
			return usageErr("--paths: %v", err)
		}
		globs = append(globs, g)
	}
	if repl.Empty() && len(decs) == 0 && len(globs) == 0 {
		return usageErr("say what the change touches: give at least one of --replaces, --decision or --paths")
	}
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	for _, id := range decs {
		d, err := pr.decision(id)
		if err != nil {
			return err
		}
		if d.Status.State == plan.DecSuperseded {
			return refused("%s is already %s", id, d.Status)
		}
	}
	id := pr.plan.NextChangeID()
	// The snapshot comes first: it refuses while ajiya/ has uncommitted changes,
	// and it must name the plan as it was before this change.
	snap, err := gitx.CreateSnapshot(pr.cfg.Root, plan.Dir, plan.SnapshotName(id))
	if err != nil {
		return refused("%v", err)
	}
	c := &plan.Change{
		ID: id, Title: title, Status: plan.ChangeStatus{State: plan.ChgOpen, Date: e.today()},
		Replaces: repl, With: by, Decisions: decs, Paths: globs, Snapshot: snap.Name,
	}
	if err := pr.plan.SaveChange(c); err != nil {
		return err
	}
	if *asJSON {
		return writeJSON(e, view.ChangeOf(c))
	}
	fmt.Fprintf(e.stdout, "Opened %s: %s\n", c.ID, c.Title)
	fmt.Fprintf(e.stdout, "Snapshot %s (tag %s at %s).\n", snap.Name, snap.Tag, snap.Commit)
	fmt.Fprintf(e.stdout, "%s is new and uncommitted (the snapshot names the commit before it); commit it with the plan.\n", plan.ChangePath(c.ID))
	fmt.Fprintf(e.stdout, "Next: 'ajiya change impact %s' lists the tickets it might affect.\n", c.ID)
	return nil
}

// ticketFilePaths returns, for every ticket named by a commit, the files those
// commits changed as paths from the project root.
func (pr *project) ticketFilePaths() (map[string][]string, error) {
	all, err := pr.filesByTicket()
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for id, tf := range all {
		for _, f := range tf.Files {
			out[id] = append(out[id], pr.displayPath(f))
		}
	}
	return out, nil
}

func runChangeImpact(e *env, args []string) error {
	fs := newFlags("change impact")
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	c, err := pr.openChange(pos[0])
	if err != nil {
		return err
	}
	var files map[string][]string
	if len(c.Paths) > 0 { // commits are read only when a path can match them
		if files, err = pr.ticketFilePaths(); err != nil {
			return err
		}
	}
	res := c.MergeCandidates(pr.plan, pr.plan.FindImpact(c, files))
	if err := pr.plan.SaveChange(c); err != nil {
		return err
	}
	if *asJSON {
		return writeJSON(e, struct {
			view.Change
			Impact struct {
				Added   []string `json:"added"`
				Kept    []string `json:"kept"`
				Stale   []string `json:"stale"`
				Dropped []string `json:"dropped"`
			} `json:"impact"`
		}{Change: view.ChangeOf(c), Impact: struct {
			Added   []string `json:"added"`
			Kept    []string `json:"kept"`
			Stale   []string `json:"stale"`
			Dropped []string `json:"dropped"`
		}{nonNil(res.Added), nonNil(res.Kept), nonNil(res.Stale), nonNil(res.Dropped)}})
	}
	w := e.stdout
	fmt.Fprintf(w, "%s  %s\n", c.ID, c.Title)
	fmt.Fprintf(w, "%d candidate(s): %d new, %d kept, %d stale, %d dropped.\n",
		len(c.Candidates), len(res.Added), len(res.Kept), len(res.Stale), len(res.Dropped))
	if len(res.Stale) > 0 {
		fmt.Fprintf(w, "Stale rows (%s) no longer match but keep their verdicts.\n", strings.Join(res.Stale, ", "))
	}
	if len(res.Dropped) > 0 {
		fmt.Fprintf(w, "Dropped (no longer match, no verdict): %s.\n", strings.Join(res.Dropped, ", "))
	}
	pr.printCandidates(e, c)
	fmt.Fprintf(w, "%s changed; commit it with the plan.\n", plan.ChangePath(c.ID))
	return nil
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func (pr *project) printCandidates(e *env, c *plan.Change) {
	if len(c.Candidates) == 0 {
		fmt.Fprintln(e.stdout, "No candidates.")
		return
	}
	w := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATE\tVERDICT\tREASONS\tTICKET")
	for _, k := range c.Candidates {
		state := "-"
		if t := pr.plan.Ticket(k.ID); t != nil {
			state = stateNames[t.Status.State]
		}
		verdict := string(k.Verdict)
		if verdict == "" {
			verdict = "-"
		}
		reasons := strings.Join(k.Reasons, " ")
		if k.Stale {
			reasons += " (stale)"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", k.ID, state, verdict, reasons, k.Title)
	}
	w.Flush()
	for _, k := range c.Candidates {
		if k.Why == "" {
			continue
		}
		fmt.Fprintf(e.stdout, "  %s %s: %s\n", k.ID, k.Verdict, k.Why)
		if k.NewTitle != "" || k.NewDoneWhen != "" {
			fmt.Fprintf(e.stdout, "    new ticket: %s | done when: %s\n", k.NewTitle, k.NewDoneWhen)
		}
	}
}

func runChangeList(e *env, args []string) error {
	fs := newFlags("change list")
	asJSON := fs.Bool("json", false, "print JSON")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	if *asJSON {
		items := []view.Change{}
		for _, c := range pr.plan.Changes {
			items = append(items, view.ChangeOf(c))
		}
		return writeJSON(e, items)
	}
	if len(pr.plan.Changes) == 0 {
		fmt.Fprintln(e.stdout, "No changes.")
		return nil
	}
	fmt.Fprintf(e.stdout, "%d change(s)\n", len(pr.plan.Changes))
	w := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
	for _, c := range pr.plan.Changes {
		judged := 0
		for _, k := range c.Candidates {
			if k.Verdict != "" {
				judged++
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%d candidate(s), %d with a verdict\t%s\n", c.ID, c.Status.State, len(c.Candidates), judged, c.Title)
	}
	return w.Flush()
}

func runChangeShow(e *env, args []string) error {
	fs := newFlags("change show")
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	c, err := pr.change(pos[0])
	if err != nil {
		return err
	}
	if *asJSON {
		return writeJSON(e, view.ChangeOf(c))
	}
	w := e.stdout
	fmt.Fprintf(w, "%s  %s\n", c.ID, c.Title)
	fmt.Fprintf(w, "Status:     %s\n", c.Status)
	fmt.Fprintf(w, "Replaces:   %s\n", c.Replaces)
	fmt.Fprintf(w, "With:       %s\n", c.With)
	if len(c.Decisions) == 0 {
		fmt.Fprintf(w, "%-11s -\n", "Decisions:")
	}
	for i, id := range c.Decisions {
		label := "Decisions:"
		if i > 0 {
			label = ""
		}
		text := "(does not exist)"
		if l := pr.decLink(id); l.State != "missing" {
			text = l.Status + " " + l.Title
		}
		fmt.Fprintf(w, "%-11s %s %s\n", label, id, text)
	}
	paths := "-"
	if len(c.Paths) > 0 {
		paths = strings.Join(c.Paths, " ")
	}
	fmt.Fprintf(w, "Paths:      %s\n", paths)
	fmt.Fprintf(w, "Snapshot:   %s\n", c.Snapshot)
	fmt.Fprintf(w, "\nCandidates (%d)\n", len(c.Candidates))
	pr.printCandidates(e, c)
	if c.Results != "" {
		fmt.Fprintf(w, "\nResults\n%s\n", c.Results)
	}
	return nil
}

// changeLink is a change that listed a ticket as a candidate.
type changeLink struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Status  string   `json:"status"` // open, approved, rejected
	Verdict string   `json:"verdict"`
	Reasons []string `json:"reasons"`
}

// ticketChanges lists the changes that touched a ticket, by change ID.
func (pr *project) ticketChanges(id string) []changeLink {
	out := []changeLink{}
	for _, c := range pr.plan.TicketChanges(id) {
		k := c.Candidate(id)
		reasons := k.Reasons
		if reasons == nil {
			reasons = []string{}
		}
		out = append(out, changeLink{c.ID, c.Title, view.ChangeStateName(c.Status.State), string(k.Verdict), reasons})
	}
	return out
}
