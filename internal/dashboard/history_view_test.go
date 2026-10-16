package dashboard

import (
	"strings"
	"testing"
)

// History is an optional view drawn from data.js's history list: events in
// date order, newest first, each linking to its card and to the tickets an
// approved change touched.
func TestHistoryView(t *testing.T) {
	s := pageText(t)
	for _, want := range []string{
		// optional: hidden until the plan has an event; reads d.history
		"optional: () => timeline.length", "timeline = arr(d.history);", `data-tab="history"`, "history: renderHistory",
		"timeline.filter(e => recMatch(text(e))).reverse()",
		// event kinds the Go side writes (internal/build/history.go)
		"decision_proposed:", "decision_accepted:", "decision_superseded:", "change_opened:", "change_approved:", "change_rejected:",
		// event fields
		"e.replaces", "e.superseded_by", "e.change", "e.reason", "e.outcome", "e.by", "e.date",
		// what an approved change did
		"o.created", "o.superseded", "x.was_state", "x.was", "o.edited", "o.moved", "o.decisions", "o.milestones",
		"'Created'", "'Superseded'", "'Edited'", "'Dependencies moved'",
		// links: ticket chips open the dialog, decision and change chips jump to their cards
		"chip(c.id)", "decChip(", "chgChip(",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("page lacks %q", want)
		}
	}
	// window.history must stay reachable: the data is not called history.
	if strings.Contains(s, "let history") || strings.Contains(s, ", history,") {
		t.Error("a local named history would hide window.history")
	}
}
