package plan

import (
	"fmt"
	"strings"
)

// ApplyOptions are the inputs of ApplyChange that are not in the change file.
type ApplyOptions struct {
	Prefix      string // ticket ID prefix of the project, such as AJ
	NewDecision string // decision that replaces the superseded ones; "" for none
	By          string // who approves
	Date        string // YYYY-MM-DD
	// Milestones are checked for targets that this change supersedes; the
	// ones that need it come back as Retargets for the caller to write.
	Milestones []Milestone
}

// Retarget is a milestone whose targets this change moves: Targets is the new
// list. The caller writes it to ajiya.toml in the same write as the plan.
type Retarget struct {
	Name    string
	Targets []string
}

// ApplyError lists everything that stops a change from being applied.
type ApplyError struct{ Problems []string }

func (e *ApplyError) Error() string { return strings.Join(e.Problems, "; ") }

// CandidateProblems lists why a change cannot be approved yet: candidates with
// no verdict or with investigate. It does not look at the plan.
func (c *Change) CandidateProblems() []string {
	var missing, investigate []string
	for _, k := range c.Candidates {
		switch k.Verdict {
		case "":
			missing = append(missing, k.ID)
		case VerdictInvestigate:
			investigate = append(investigate, k.ID)
		}
	}
	var out []string
	if len(missing) > 0 {
		out = append(out, "no verdict yet for "+strings.Join(missing, ", "))
	}
	if len(investigate) > 0 {
		out = append(out, "still marked investigate: "+strings.Join(investigate, ", "))
	}
	return out
}

// newTicketTags is the tag set of a ticket created from old by a change: its
// tags, less the replaced values, plus what the change brings in. A decision
// the change supersedes is swapped for newDec (or dropped without one).
func newTicketTags(old Tags, c *Change, newDec string) Tags {
	out := old
	for _, key := range TagKeys {
		var vals []string
		drop := map[string]bool{}
		for _, v := range c.Replaces.Values(key) {
			drop[v] = true
		}
		if key == TagDec {
			for _, d := range c.Decisions {
				drop[d] = true
			}
		}
		dropped := false
		for _, v := range old.Values(key) {
			if drop[v] {
				dropped = true
			} else {
				vals = append(vals, v)
			}
		}
		vals = append(vals, c.With.Values(key)...)
		if key == TagDec && dropped && newDec != "" {
			vals = append(vals, newDec)
		}
		out, _ = out.With(key, vals) // values came from valid tags
	}
	return out
}

// ApplyChange applies every verdict of an open change to the plan in memory:
// new tickets, edits, superseded tickets, moved dependencies and superseded
// decisions. It marks the change Approved and writes its Results. It changes
// nothing if the checks up front refuse it; the refusal of an emptied
// milestone comes later, so on any error the caller discards the plan. The caller validates the result and saves
// the phases, decisions and the change (Save, SaveDecision, SaveChange).
//
// Rules:
//   - unaffected: nothing.
//   - modify of an open ticket: its title and done-when are edited in place
//     (an empty new value leaves that field alone). Tags are not touched.
//   - modify of a done ticket: a follow-up ticket is added to the same phase
//     and app, depending on it.
//   - replace: a new ticket is added to the same phase and app, depending on
//     what the old one depended on, and the old one is superseded by it.
//   - invalidate: the ticket is superseded with no replacement.
//   - Every ticket that is not closed and depended on a replaced ticket now
//     depends on its replacement. For an invalidated ticket the dependency is
//     replaced by that ticket's own dependencies (what it was waiting on).
//     Closed tickets keep their dependencies: history is not rewritten.
//   - A new ticket (from replace, or the follow-up of modify) gets the old
//     ticket's tags, less the replaced values, plus the change's With tags.
func (p *Plan) ApplyChange(c *Change, o ApplyOptions) ([]Retarget, error) {
	if !c.Status.Open() {
		return nil, fmt.Errorf("%s is %s", c.ID, strings.ToLower(c.Status.State.String()))
	}
	if probs := c.CandidateProblems(); len(probs) > 0 {
		return nil, &ApplyError{probs}
	}
	if len(c.Candidates) == 0 && len(c.Decisions) == 0 {
		return nil, &ApplyError{[]string{"the change has no candidates and names no decision, so there is nothing to apply; run 'ajiya change impact " + c.ID + "' or reject it"}}
	}
	if err := CheckStatusText("--by", o.By); err != nil || o.By == "" {
		return nil, &ApplyError{[]string{"--by needs a name without '·'"}}
	}

	// Everything that can be refused is checked before anything changes.
	var probs []string
	for _, k := range c.Candidates {
		t := p.Ticket(k.ID)
		switch {
		case t == nil:
			if k.Verdict != VerdictUnaffected {
				probs = append(probs, fmt.Sprintf("%s (%s) no longer exists", k.ID, k.Verdict))
			}
		case t.Status.Excluded() && k.Verdict != VerdictUnaffected:
			probs = append(probs, fmt.Sprintf("%s is already %s, so %s cannot be applied; assess it as unaffected", k.ID, strings.ToLower(stateWord(t.Status.State)), k.Verdict))
		case k.Verdict == VerdictReplace && (k.NewTitle == "" || k.NewDoneWhen == ""):
			probs = append(probs, k.ID+" is replace but has no new title and done-when")
		case k.Verdict == VerdictModify && t.Status.State == Done && (k.NewTitle == "" || k.NewDoneWhen == ""):
			probs = append(probs, k.ID+" is modify of a done ticket but has no new title and done-when")
		case k.Verdict == VerdictModify && !t.Status.Closed() && k.NewTitle == "" && k.NewDoneWhen == "":
			probs = append(probs, k.ID+" is modify but has no new title or done-when")
		}
	}
	for _, id := range c.Decisions {
		d := p.Decision(id)
		switch {
		case d == nil:
			probs = append(probs, "decision "+id+" does not exist")
		case d.Status.State == DecSuperseded:
			probs = append(probs, fmt.Sprintf("decision %s is already %s", id, d.Status))
		}
	}
	if o.NewDecision != "" {
		d := p.Decision(o.NewDecision)
		switch {
		case !ValidID(DecPrefix, o.NewDecision) || d == nil:
			probs = append(probs, "--new-decision "+o.NewDecision+" does not exist; add it with 'ajiya decision add'")
		case d.Status.State == DecSuperseded:
			probs = append(probs, "--new-decision "+o.NewDecision+" is itself superseded")
		case slicesContain(c.Decisions, o.NewDecision):
			probs = append(probs, "--new-decision "+o.NewDecision+" is one of the decisions being superseded")
		}
	}
	if len(probs) > 0 {
		return nil, &ApplyError{probs}
	}

	var res []string
	repl := map[string]string{} // old ticket -> its replacement
	inval := map[string]bool{}  // invalidated tickets
	var created []*Ticket

	for _, k := range c.Candidates {
		t := p.Ticket(k.ID)
		if t == nil {
			continue
		}
		switch k.Verdict {
		case VerdictModify:
			if t.Status.Closed() { // done: a follow-up
				n := p.addChangeTicket(t, k, o, c, append([]string(nil), t.ID))
				// the follow-up depends on the done ticket, not on what it depended on
				created = append(created, n)
				res = append(res, fmt.Sprintf("- Created %s: %s (follow-up to %s)", n.ID, n.Title, t.ID))
				continue
			}
			var edits []string
			if k.NewTitle != "" && k.NewTitle != t.Title {
				t.Title = k.NewTitle
				edits = append(edits, "title")
			}
			if k.NewDoneWhen != "" && k.NewDoneWhen != t.DoneWhen {
				t.DoneWhen = k.NewDoneWhen
				edits = append(edits, "done when")
			}
			// The one tag an edit changes: a link to a decision this change
			// supersedes moves to the new decision (or goes), so the open ticket
			// does not keep pointing at a superseded decision (W014).
			if swapped := swapDecisionTags(t.Tags, c.Decisions, o.NewDecision); swapped.String() != t.Tags.String() {
				t.Tags = swapped
				edits = append(edits, "decision tag")
			}
			if len(edits) > 0 {
				res = append(res, fmt.Sprintf("- Edited %s: %s (%s)", t.ID, t.Title, strings.Join(edits, ", ")))
			}
		case VerdictReplace:
			n := p.addChangeTicket(t, k, o, c, append([]string(nil), t.Depends...))
			created = append(created, n)
			res = append(res, fmt.Sprintf("- Created %s: %s (replaces %s)", n.ID, n.Title, t.ID))
			repl[t.ID] = n.ID
		case VerdictInvalidate:
			inval[t.ID] = true
		}
	}
	// Supersede the old tickets, in ID order.
	var ids []string
	for id := range repl {
		ids = append(ids, id)
	}
	for id := range inval {
		ids = append(ids, id)
	}
	SortIDs(ids)
	for _, id := range ids {
		t := p.Ticket(id)
		by := repl[id] // "" when invalidated
		st, err := t.Status.Supersede(by, c.ID)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", id, err) // unreachable: Excluded was checked above
		}
		was := t.Status
		t.Status = st
		if by != "" {
			res = append(res, fmt.Sprintf("- Superseded %s by %s (was: %s)", id, by, statusTail(was)))
		} else {
			res = append(res, fmt.Sprintf("- Superseded %s with no replacement (was: %s)", id, statusTail(was)))
		}
	}

	// resolve is what a superseded ticket stands for now: its replacement, or
	// for an invalidated one what it was waiting on.
	var resolve func(id string, seen map[string]bool) []string
	resolve = func(id string, seen map[string]bool) []string {
		if r, ok := repl[id]; ok {
			return []string{r}
		}
		if !inval[id] || seen[id] {
			return []string{id}
		}
		seen[id] = true
		var out []string
		if t := p.Ticket(id); t != nil {
			for _, d := range t.Depends {
				out = append(out, resolve(d, seen)...)
			}
		}
		return out
	}

	// Retarget the milestones that name a superseded ticket. Phase targets
	// and other tickets stay. The new list is de-duplicated, in order.
	var retargets []Retarget
	var empty []string
	for _, m := range o.Milestones {
		var next, lines []string
		seen := map[string]bool{}
		add := func(id string) {
			if !seen[id] {
				seen[id] = true
				next = append(next, id)
			}
		}
		for _, target := range m.Targets {
			switch {
			case repl[target] != "":
				add(repl[target])
				lines = append(lines, fmt.Sprintf("- milestone %s: target %s \u2192 %s", m.Name, target, repl[target]))
			case inval[target]:
				deps := resolve(target, map[string]bool{})
				for _, d := range deps {
					add(d)
				}
				if len(deps) == 0 {
					lines = append(lines, fmt.Sprintf("- milestone %s: target %s removed (invalidated, it depended on nothing)", m.Name, target))
				} else {
					lines = append(lines, fmt.Sprintf("- milestone %s: target %s \u2192 %s (invalidated; its dependencies)", m.Name, target, strings.Join(deps, ", ")))
				}
			default:
				add(target)
			}
		}
		if len(lines) == 0 {
			continue
		}
		if len(next) == 0 {
			empty = append(empty, m.Name)
		}
		retargets = append(retargets, Retarget{Name: m.Name, Targets: next})
		res = append(res, lines...)
	}
	if len(empty) > 0 {
		return nil, &ApplyError{[]string{"milestone " + strings.Join(empty, ", ") + " would be left with no targets, because every target is invalidated and depended on nothing; retarget or remove the milestone first ('ajiya milestone remove', 'ajiya launch set'), or assess the ticket as replace"}}
	}

	// Move dependencies of open tickets.
	if len(repl)+len(inval) > 0 {
		fresh := map[*Ticket]bool{}
		for _, n := range created {
			fresh[n] = true
		}
		for _, t := range p.Tickets() {
			if t.Status.Closed() {
				continue
			}
			var next []string
			seen := map[string]bool{}
			changed := false
			for _, d := range t.Depends {
				if repl[d] != "" || inval[d] {
					changed = true
					for _, r := range resolve(d, map[string]bool{}) {
						if r != t.ID && !seen[r] {
							seen[r] = true
							next = append(next, r)
						}
					}
					continue
				}
				if !seen[d] {
					seen[d] = true
					next = append(next, d)
				}
			}
			if !changed {
				continue
			}
			SortIDs(next)
			old := t.Depends
			t.Depends = next
			if !fresh[t] {
				res = append(res, fmt.Sprintf("- Moved dependencies of %s: %s -> %s", t.ID, depList(old), depList(next)))
			}
		}
	}

	for _, id := range c.Decisions {
		d := p.Decision(id)
		if err := d.SetSuperseded(o.NewDecision, c.ID); err != nil {
			return nil, err
		}
		if o.NewDecision != "" {
			res = append(res, fmt.Sprintf("- Superseded decision %s by %s", id, o.NewDecision))
		} else {
			res = append(res, fmt.Sprintf("- Superseded decision %s with no replacement", id))
		}
	}

	if len(res) == 0 {
		res = []string{"- No ticket or decision needed changing"}
	}
	c.Results = strings.Join(res, "\n")
	c.Status = ChangeStatus{State: ChgApproved, Date: o.Date, By: o.By}
	return retargets, nil
}

// addChangeTicket adds the ticket a replace, or the modify of a done ticket,
// creates, in the same phase and app as old.
func (p *Plan) addChangeTicket(old *Ticket, k Candidate, o ApplyOptions, c *Change, deps []string) *Ticket {
	ph := p.Phase(old.Phase)
	st := Status{State: Pending}
	if !old.Status.Closed() {
		st.Human = old.Status.Human
	}
	SortIDs(deps)
	n := &Ticket{
		ID: p.NextID(o.Prefix), App: old.App, Title: k.NewTitle, DoneWhen: k.NewDoneWhen,
		Depends: deps, Tags: newTicketTags(old.Tags, c, o.NewDecision), Status: st, Phase: ph.Slug,
	}
	ph.Tickets = append(ph.Tickets, n)
	return n
}

// swapDecisionTags replaces links to the decisions in old with newDec, or drops
// them when newDec is empty.
func swapDecisionTags(t Tags, old []string, newDec string) Tags {
	var vals []string
	hit := false
	for _, v := range t.Dec {
		if slicesContain(old, v) {
			hit = true
			continue
		}
		vals = append(vals, v)
	}
	if hit && newDec != "" {
		vals = append(vals, newDec)
	}
	out, err := t.With(TagDec, vals)
	if err != nil {
		return t
	}
	return out
}

func stateWord(s State) string {
	switch s {
	case Dropped:
		return "Dropped"
	case Superseded:
		return "Superseded"
	}
	return "closed"
}

// statusTail is a status without its coloured square, as shown after "was:".
func statusTail(s Status) string {
	_, tail, _ := strings.Cut(s.String(), " ")
	return tail
}

func depList(ids []string) string {
	if len(ids) == 0 {
		return "-"
	}
	return strings.Join(ids, ", ")
}

func slicesContain(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}
