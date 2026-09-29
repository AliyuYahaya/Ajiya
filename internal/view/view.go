// Package view holds the JSON shapes ajiya prints (--json, MCP, data.js).
// They are part of ajiya's interface: add fields, never rename or remove them.
package view

import "github.com/AliyuYahaya/Ajiya/internal/plan"

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
}

// Ticket is one ticket row.
type Ticket struct {
	ID       string   `json:"id"`
	Phase    string   `json:"phase"`
	App      string   `json:"app"`
	Title    string   `json:"title"`
	DoneWhen string   `json:"done_when"`
	Depends  []string `json:"depends"`
	Status   Status   `json:"status"`
}

// StateNames are the JSON names of ticket states.
var StateNames = map[plan.State]string{plan.Pending: "pending", plan.InProgress: "in_progress", plan.Done: "done", plan.Dropped: "dropped"}

// TicketOf returns the JSON form of a ticket.
func TicketOf(t *plan.Ticket) Ticket {
	s := t.Status
	deps := t.Depends
	if deps == nil {
		deps = []string{}
	}
	return Ticket{
		ID: t.ID, Phase: t.Phase, App: t.App, Title: t.Title, DoneWhen: t.DoneWhen, Depends: deps,
		Status: Status{
			State: StateNames[s.State], Text: s.String(), Note: s.Note, Human: s.Human, Blocked: s.Blocked,
			Commit: s.Commit, By: s.By, Date: s.Date, Tests: s.Tests, Reason: s.Reason,
			Issue: s.Issue, Before: s.Before, Link: s.Link, Aliases: s.Aliases,
		},
	}
}
