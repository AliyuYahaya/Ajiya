// Package view holds the JSON shapes ajiya prints (--json, MCP, data.js).
// They are part of ajiya's interface: add fields, never rename or remove them.
package view

import (
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// Status is a ticket's Status cell, parsed.
type Status struct {
	State   string   `json:"state"` // pending, in_progress, done, dropped
	Text    string   `json:"text"`  // the Status cell as written
	Note    string   `json:"note,omitempty"`
	Human   string   `json:"human,omitempty"`
	Blocked string   `json:"blocked,omitempty"`
	Commit  string   `json:"commit,omitempty"`
	By      string   `json:"by,omitempty"`
	Date    string   `json:"date,omitempty"`
	Tests   bool     `json:"tests_passed,omitempty"`
	Reason  string   `json:"reason,omitempty"`
	Issue   string   `json:"issue,omitempty"`
	Before  bool     `json:"done_before_ajiya,omitempty"`
	Link    string   `json:"link,omitempty"`
	Aliases []string `json:"aliases,omitempty"`

	// Superseded: the replacement ticket, the change, and the status it had.
	SupersededBy string  `json:"superseded_by,omitempty"`
	Change       string  `json:"change,omitempty"`
	Was          *Status `json:"was,omitempty"`
}

// Tags is a ticket's Tags cell. Each list is empty, never null.
type Tags struct {
	Tech []string `json:"tech"`
	Dec  []string `json:"dec"`
	Path []string `json:"path"`
}

// Ticket is one ticket row.
type Ticket struct {
	ID       string   `json:"id"`
	Phase    string   `json:"phase"`
	App      string   `json:"app"`
	Title    string   `json:"title"`
	DoneWhen string   `json:"done_when"`
	Depends  []string `json:"depends"`
	Tags     Tags     `json:"tags"`
	Status   Status   `json:"status"`
}

// StateNames are the JSON names of ticket states.
var StateNames = map[plan.State]string{plan.Pending: "pending", plan.InProgress: "in_progress", plan.Done: "done", plan.Dropped: "dropped", plan.Superseded: "superseded"}

// TicketOf returns the JSON form of a ticket.
func TicketOf(t *plan.Ticket) Ticket {
	s := t.Status
	deps := t.Depends
	if deps == nil {
		deps = []string{}
	}
	orEmpty := func(s []string) []string {
		if s == nil {
			return []string{}
		}
		return s
	}
	return Ticket{
		ID: t.ID, Phase: t.Phase, App: t.App, Title: t.Title, DoneWhen: t.DoneWhen, Depends: deps,
		Tags:   Tags{Tech: orEmpty(t.Tags.Tech), Dec: orEmpty(t.Tags.Dec), Path: orEmpty(t.Tags.Path)},
		Status: StatusOf(s),
	}
}

// DecisionStatus is a decision's Status line, parsed.
type DecisionStatus struct {
	State        string `json:"state"` // proposed, accepted, superseded
	Text         string `json:"text"`  // the Status line as written
	Date         string `json:"date,omitempty"`
	By           string `json:"by,omitempty"`
	SupersededBy string `json:"superseded_by,omitempty"`
	Change       string `json:"change,omitempty"`
	Was          string `json:"was,omitempty"` // for superseded: proposed or accepted
}

// Decision is one decision record. Tickets are the IDs of the tickets that link it.
type Decision struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Status       DecisionStatus `json:"status"`
	Tech         []string       `json:"tech"`
	Context      string         `json:"context"`
	Decision     string         `json:"decision"`
	Consequences string         `json:"consequences"`
	Tickets      []string       `json:"tickets"`
}

// DecisionStateName is the JSON name of a decision state.
func DecisionStateName(s plan.DecState) string { return strings.ToLower(s.String()) }

// DecisionOf returns the JSON form of a decision. tickets are the IDs linking it.
func DecisionOf(d *plan.Decision, tickets []string) Decision {
	s := d.Status
	st := DecisionStatus{State: DecisionStateName(s.State), Text: s.String(), Date: s.Date, By: s.By,
		SupersededBy: s.SupersededBy, Change: s.Change}
	if s.State == plan.DecSuperseded {
		st.Was = DecisionStateName(s.Was)
	}
	tech := d.Tags.Tech
	if tech == nil {
		tech = []string{}
	}
	if tickets == nil {
		tickets = []string{}
	}
	return Decision{ID: d.ID, Title: d.Title, Status: st, Tech: tech, Context: d.Context,
		Decision: d.Decision, Consequences: d.Consequences, Tickets: tickets}
}

// StatusOf returns the JSON form of a status.
func StatusOf(s plan.Status) Status {
	out := Status{
		State: StateNames[s.State], Text: s.String(), Note: s.Note, Human: s.Human, Blocked: s.Blocked,
		Commit: s.Commit, By: s.By, Date: s.Date, Tests: s.Tests, Reason: s.Reason,
		Issue: s.Issue, Before: s.Before, Link: s.Link, Aliases: s.Aliases,
		SupersededBy: s.SupersededBy, Change: s.Change,
	}
	if s.Prev != nil {
		prev := StatusOf(*s.Prev)
		out.Was = &prev
	}
	return out
}
