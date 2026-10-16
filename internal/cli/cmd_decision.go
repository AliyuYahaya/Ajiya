package cli

import (
	"fmt"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
	"github.com/AliyuYahaya/Ajiya/internal/view"
)

// today is the date for new records: the clock of the env when it has one.
func (e *env) today() string {
	if e.now != nil {
		return e.now().Format("2006-01-02")
	}
	return today()
}

func ticketIDs(ts []*plan.Ticket) []string {
	ids := []string{}
	for _, t := range ts {
		ids = append(ids, t.ID)
	}
	return ids
}

func (pr *project) decisionJSON(d *plan.Decision) view.Decision {
	return view.DecisionOf(d, ticketIDs(pr.plan.DecisionTickets(d.ID)))
}

func (pr *project) decision(id string) (*plan.Decision, error) {
	if d := pr.plan.Decision(id); d != nil {
		return d, nil
	}
	if pe := pr.plan.BrokenDecisions[id]; pe != nil {
		return nil, refused("%s:%d: %s; see 'ajiya check'", plan.DecisionPath(id), pe.Line, pe.Msg)
	}
	return nil, refused("no decision %s; see 'ajiya decision list'", id)
}

func runDecisionAdd(e *env, args []string) error {
	fs := newFlags("decision add")
	decision := fs.String("decision", "", "what was chosen")
	context := fs.String("context", "", "why the choice was needed")
	consequences := fs.String("consequences", "", "what follows from it, including what it rules out")
	tech := fs.String("tech", "", "technology or service tags, comma separated (lower-case slugs such as aws-rds)")
	by := fs.String("by", "", "who proposed it")
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	title := pos[0]
	if title == "" {
		return usageErr("the title is empty")
	}
	if *decision == "" {
		return usageErr("--decision is required: say what was chosen")
	}
	if err := checkTexts("title", title); err != nil {
		return err
	}
	if err := plan.CheckStatusText("--by", *by); err != nil {
		return usageErr("%v", err)
	}
	for _, f := range []struct{ name, v string }{{"--decision", *decision}, {"--context", *context}, {"--consequences", *consequences}} {
		if err := plan.CheckDecisionText(f.name, f.v); err != nil {
			return usageErr("%v", err)
		}
	}
	tags, err := plan.Tags{}.With(plan.TagTech, plan.ParseTagList(*tech))
	if err != nil {
		return usageErr("--tech: %v", err)
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	d := &plan.Decision{
		ID: pr.plan.NextDecisionID(), Title: title, Tags: tags,
		Context: *context, Decision: *decision, Consequences: *consequences,
		Status: plan.DecStatus{State: plan.DecProposed, Date: e.today(), By: *by},
	}
	if err := pr.plan.SaveDecision(d); err != nil {
		return err
	}
	if *asJSON {
		return writeJSON(e, pr.decisionJSON(d))
	}
	fmt.Fprintf(e.stdout, "Added %s (%s): %s\n", d.ID, d.Status.State, d.Title)
	fmt.Fprintf(e.stdout, "Record who decided with 'ajiya decision accept %s --by <name>'.\n", d.ID)
	return nil
}

func runDecisionShow(e *env, args []string) error {
	fs := newFlags("decision show")
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	d, err := pr.decision(pos[0])
	if err != nil {
		return err
	}
	if *asJSON {
		return writeJSON(e, pr.decisionJSON(d))
	}
	w := e.stdout
	fmt.Fprintf(w, "%s  %s\n", d.ID, d.Title)
	fmt.Fprintf(w, "Status:     %s\n", d.Status)
	fmt.Fprintf(w, "Tags:       %s\n", d.Tags)
	linked := pr.plan.DecisionTickets(d.ID)
	if len(linked) == 0 {
		fmt.Fprintf(w, "%-11s -\n", "Tickets:")
	}
	for i, t := range linked {
		label := "Tickets:"
		if i > 0 {
			label = ""
		}
		fmt.Fprintf(w, "%-11s %s %s %s\n", label, t.ID, t.Status.Mark(), t.Title)
	}
	for _, s := range []struct{ name, text string }{{"Context", d.Context}, {"Decision", d.Decision}, {"Consequences", d.Consequences}} {
		text := s.text
		if text == "" {
			text = "-"
		}
		fmt.Fprintf(w, "\n%s\n%s\n", s.name, text)
	}
	return nil
}

func runDecisionList(e *env, args []string) error {
	fs := newFlags("decision list")
	tag := fs.String("tag", "", "only decisions with this tag, such as tech:postgres")
	asJSON := fs.Bool("json", false, "print JSON")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	value := ""
	if isSet(fs, "tag") {
		key, v, ok := strings.Cut(*tag, ":")
		if !ok {
			return usageErr("--tag %q is not key:value, for example tech:postgres", *tag)
		}
		if key != plan.TagTech {
			return usageErr("--tag: decisions carry only tech tags, for example tech:postgres")
		}
		if err := plan.CheckTag(key, v); err != nil {
			return usageErr("--tag: %v", err)
		}
		value = v
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	var ds []*plan.Decision
	for _, d := range pr.plan.Decisions {
		if value == "" || slices.Contains(d.Tags.Tech, value) {
			ds = append(ds, d)
		}
	}
	if *asJSON {
		items := []view.Decision{}
		for _, d := range ds {
			items = append(items, pr.decisionJSON(d))
		}
		return writeJSON(e, items)
	}
	if len(ds) == 0 {
		fmt.Fprintln(e.stdout, "No decisions.")
		return nil
	}
	fmt.Fprintf(e.stdout, "%d decision(s)\n", len(ds))
	w := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
	for _, d := range ds {
		fmt.Fprintf(w, "%s\t%s\t%s\n", d.ID, d.Status.State, d.Title)
	}
	return w.Flush()
}

func runDecisionAccept(e *env, args []string) error {
	fs := newFlags("decision accept")
	by := fs.String("by", "", "who decided")
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	if *by == "" {
		return usageErr("--by is required: say who decided")
	}
	if err := plan.CheckStatusText("--by", *by); err != nil {
		return usageErr("%v", err)
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	d, err := pr.decision(pos[0])
	if err != nil {
		return err
	}
	if err := d.Accept(*by, e.today()); err != nil {
		return refused("%v", err)
	}
	if err := pr.plan.SaveDecision(d); err != nil {
		return err
	}
	if *asJSON {
		return writeJSON(e, pr.decisionJSON(d))
	}
	fmt.Fprintf(e.stdout, "%s %s\n", d.ID, d.Status)
	return nil
}
