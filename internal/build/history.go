package build

import (
	"regexp"
	"sort"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// Event kinds in Data.History.
const (
	EvDecisionProposed   = "decision_proposed"
	EvDecisionAccepted   = "decision_accepted"
	EvDecisionSuperseded = "decision_superseded"
	EvChangeOpened       = "change_opened"
	EvChangeApproved     = "change_approved"
	EvChangeRejected     = "change_rejected"
)

// HistoryEvent is one step in how the plan changed. History is oldest first;
// events on the same day keep the order things happen in (a decision is
// proposed, then accepted, a change opens, a change closes, and a decision it
// replaces is superseded).
//
// A decision's proposed and accepted events carry the date of the state the
// decision is in now (or was in before it was superseded): the file keeps one
// date for it. A decision superseded is dated by the change that did it. A
// change keeps one date too: when it was opened while open, when it was
// approved or rejected after that.
type HistoryEvent struct {
	Date  string `json:"date"` // YYYY-MM-DD; "" when the record has none
	Kind  string `json:"kind"`
	ID    string `json:"id"` // the decision or change the event is about
	Title string `json:"title"`
	By    string `json:"by,omitempty"`
	// Text is the event in one line, such as "DEC-0003 accepted by Aliyu".
	Text string `json:"text"`
	// Change is, on a decision superseded, the change that did it.
	Change string `json:"change,omitempty"`
	// Replaces is, on a decision's own events, the decisions it took the place
	// of. SupersededBy is, on a decision superseded, its replacement.
	Replaces     []string `json:"replaces,omitempty"`
	SupersededBy string   `json:"superseded_by,omitempty"`
	// Reason is why a change was rejected.
	Reason string `json:"reason,omitempty"`
	// Outcome is what an approved change did to the plan.
	Outcome *Outcome `json:"outcome,omitempty"`
}

// Outcome is the Results of an approved change, as lists of IDs.
type Outcome struct {
	Created    []Created    `json:"created"`
	Superseded []Superseded `json:"superseded"`
	Edited     []Edited     `json:"edited"`
	Moved      []Moved      `json:"moved"`
	Decisions  []DecMove    `json:"decisions"`
	Milestones []string     `json:"milestones"` // retarget lines, as written
}

// Created is a ticket the change added.
type Created struct {
	ID       string `json:"id"`
	Replaces string `json:"replaces,omitempty"`  // the ticket it replaces
	FollowUp string `json:"follow_up,omitempty"` // the done ticket it follows
}

// Superseded is a ticket the change closed.
type Superseded struct {
	ID       string `json:"id"`
	By       string `json:"by,omitempty"`        // "" when invalidated
	Was      string `json:"was"`                 // the status it had, as written
	WasState string `json:"was_state,omitempty"` // done, in_progress or pending
}

// Edited is an open ticket the change edited in place.
type Edited struct {
	ID     string   `json:"id"`
	Fields []string `json:"fields"` // title, done when, decision tag
}

// Moved is a ticket whose dependencies the change moved.
type Moved struct {
	ID   string   `json:"id"`
	From []string `json:"from"`
	To   []string `json:"to"`
}

// DecMove is a decision the change superseded.
type DecMove struct {
	ID string `json:"id"`
	By string `json:"by,omitempty"` // "" when superseded with no replacement
}

// The line shapes ApplyChange writes into Results.
var (
	createdRE = regexp.MustCompile(`^- Created (\S+): .* \((replaces|follow-up to) (\S+)\)$`)
	editedRE  = regexp.MustCompile(`^- Edited (\S+): .* \(([^()]*)\)$`)
	supByRE   = regexp.MustCompile(`^- Superseded (\S+) by (\S+) \(was: (.*)\)$`)
	supNoRE   = regexp.MustCompile(`^- Superseded (\S+) with no replacement \(was: (.*)\)$`)
	movedRE   = regexp.MustCompile(`^- Moved dependencies of (\S+): (.*) -> (.*)$`)
	decByRE   = regexp.MustCompile(`^- Superseded decision (\S+) by (\S+)$`)
	decNoRE   = regexp.MustCompile(`^- Superseded decision (\S+) with no replacement$`)
)

func idList(s string) []string {
	out := []string{}
	if s == "-" || s == "" {
		return out
	}
	for _, p := range strings.Split(s, ",") {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

func wasState(s string) string {
	switch {
	case strings.Contains(s, "Done"):
		return "done"
	case strings.Contains(s, "In progress"):
		return "in_progress"
	case strings.Contains(s, "Pending"):
		return "pending"
	}
	return ""
}

// ParseResults reads the Results text that plan.ApplyChange writes. Lines in
// another shape are left out.
func ParseResults(results string) *Outcome {
	o := &Outcome{Created: []Created{}, Superseded: []Superseded{}, Edited: []Edited{}, Moved: []Moved{}, Decisions: []DecMove{}, Milestones: []string{}}
	for _, line := range strings.Split(results, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case createdRE.MatchString(line):
			m := createdRE.FindStringSubmatch(line)
			c := Created{ID: m[1]}
			if m[2] == "replaces" {
				c.Replaces = m[3]
			} else {
				c.FollowUp = m[3]
			}
			o.Created = append(o.Created, c)
		case editedRE.MatchString(line):
			m := editedRE.FindStringSubmatch(line)
			o.Edited = append(o.Edited, Edited{ID: m[1], Fields: strings.Split(m[2], ", ")})
		case supByRE.MatchString(line):
			m := supByRE.FindStringSubmatch(line)
			o.Superseded = append(o.Superseded, Superseded{ID: m[1], By: m[2], Was: m[3], WasState: wasState(m[3])})
		case supNoRE.MatchString(line):
			m := supNoRE.FindStringSubmatch(line)
			o.Superseded = append(o.Superseded, Superseded{ID: m[1], Was: m[2], WasState: wasState(m[2])})
		case movedRE.MatchString(line):
			m := movedRE.FindStringSubmatch(line)
			o.Moved = append(o.Moved, Moved{ID: m[1], From: idList(m[2]), To: idList(m[3])})
		case decByRE.MatchString(line):
			m := decByRE.FindStringSubmatch(line)
			o.Decisions = append(o.Decisions, DecMove{ID: m[1], By: m[2]})
		case decNoRE.MatchString(line):
			m := decNoRE.FindStringSubmatch(line)
			o.Decisions = append(o.Decisions, DecMove{ID: m[1]})
		case strings.HasPrefix(line, "- milestone "):
			o.Milestones = append(o.Milestones, strings.TrimPrefix(line, "- "))
		}
	}
	return o
}

// History lists the decision and change events of a plan, oldest first.
func History(p *plan.Plan) []HistoryEvent {
	type ranked struct {
		HistoryEvent
		rank int
	}
	var evs []ranked
	add := func(rank int, e HistoryEvent) { evs = append(evs, ranked{e, rank}) }

	changeDate := map[string]string{}
	for _, c := range p.Changes {
		changeDate[c.ID] = c.Status.Date
	}
	replaces := map[string][]string{} // decision -> decisions it replaced
	for _, d := range p.Decisions {
		if d.Status.State == plan.DecSuperseded && d.Status.SupersededBy != "" {
			replaces[d.Status.SupersededBy] = append(replaces[d.Status.SupersededBy], d.ID)
		}
	}

	for _, d := range p.Decisions {
		s := d.Status
		st := s.State
		if st == plan.DecSuperseded {
			st = s.Was
		}
		kind, verb, rank := EvDecisionProposed, "proposed", 0
		if st == plan.DecAccepted {
			kind, verb, rank = EvDecisionAccepted, "accepted", 1
		}
		e := HistoryEvent{Date: s.Date, Kind: kind, ID: d.ID, Title: d.Title, By: s.By, Replaces: replaces[d.ID], Text: d.ID + " " + verb}
		if s.By != "" {
			e.Text += " by " + s.By
		}
		add(rank, e)
		if s.State == plan.DecSuperseded {
			se := HistoryEvent{Date: changeDate[s.Change], Kind: EvDecisionSuperseded, ID: d.ID, Title: d.Title,
				Change: s.Change, SupersededBy: s.SupersededBy, Text: d.ID + " superseded"}
			if s.SupersededBy != "" {
				se.Text += " by " + s.SupersededBy
			}
			se.Text += " in " + s.Change
			add(4, se)
		}
	}

	for _, c := range p.Changes {
		s := c.Status
		switch s.State {
		case plan.ChgOpen:
			add(2, HistoryEvent{Date: s.Date, Kind: EvChangeOpened, ID: c.ID, Title: c.Title, Text: c.ID + " opened"})
		case plan.ChgApproved:
			add(3, HistoryEvent{Date: s.Date, Kind: EvChangeApproved, ID: c.ID, Title: c.Title, By: s.By,
				Text: c.ID + " approved by " + s.By, Outcome: ParseResults(c.Results)})
		case plan.ChgRejected:
			add(3, HistoryEvent{Date: s.Date, Kind: EvChangeRejected, ID: c.ID, Title: c.Title, By: s.By, Reason: s.Reason,
				Text: c.ID + " rejected by " + s.By})
		}
	}

	sort.SliceStable(evs, func(i, j int) bool {
		a, b := evs[i], evs[j]
		if a.Date != b.Date {
			return a.Date < b.Date
		}
		if a.rank != b.rank {
			return a.rank < b.rank
		}
		return plan.LessID(a.ID, b.ID)
	})
	out := make([]HistoryEvent, 0, len(evs))
	for _, e := range evs {
		out = append(out, e.HistoryEvent)
	}
	return out
}
