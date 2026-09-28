package cli

import (
	"encoding/json"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// The JSON shapes below are part of ajiya's interface: add fields, never
// rename or remove them.

type jsonStatus struct {
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

type jsonTicket struct {
	ID       string     `json:"id"`
	Phase    string     `json:"phase"`
	App      string     `json:"app"`
	Title    string     `json:"title"`
	DoneWhen string     `json:"done_when"`
	Depends  []string   `json:"depends"`
	Status   jsonStatus `json:"status"`
}

var stateNames = map[plan.State]string{plan.Pending: "pending", plan.InProgress: "in_progress", plan.Done: "done", plan.Dropped: "dropped"}

func toJSON(t *plan.Ticket) jsonTicket {
	s := t.Status
	deps := t.Depends
	if deps == nil {
		deps = []string{}
	}
	return jsonTicket{
		ID: t.ID, Phase: t.Phase, App: t.App, Title: t.Title, DoneWhen: t.DoneWhen, Depends: deps,
		Status: jsonStatus{
			State: stateNames[s.State], Text: s.String(), Note: s.Note, Human: s.Human, Blocked: s.Blocked,
			Commit: s.Commit, By: s.By, Date: s.Date, Tests: s.Tests, Reason: s.Reason,
			Issue: s.Issue, Before: s.Before, Link: s.Link, Aliases: s.Aliases,
		},
	}
}

func writeJSON(e *env, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = e.stdout.Write(append(data, '\n'))
	return err
}
