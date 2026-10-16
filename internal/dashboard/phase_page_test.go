package dashboard

import (
	"regexp"
	"strings"
	"testing"
)

// A phase has a page at #phase/<slug>. Every phase click goes through
// openPhase, which opens it; the description shows only when the phase has
// one, and an unknown slug gets an empty state.
func TestPhasePage(t *testing.T) {
	s := pageText(t)
	for _, want := range []string{
		"const ROUTES = { phase:", `/^phase\/(.+)$/`, "'phase/' + encodeURIComponent(phaseSlug)",
		"phase: renderPhase", "function renderPhase()", "function phaseLinks(slug)",
		"function openPhase(slug) {\n    go('phase', slug);",
		// the page: goal, description, tickets by state, phase links, decisions
		"p.goal ?", "paragraphs(p.description)", `.split(/\n{2,}/)`,
		"sect('In progress'", "sect('Can start now'", "sect('Waiting'", "sect('Needs a human'", "sect('Done'", "Dropped and superseded",
		"'Depends on phases'", "'Needed by phases'", "decChip(d.id)",
		"Phase not found", "There is no phase",
		// links to the page: Tickets group headers and the phase cards
		`href="#phase/`, `data-phase-go="`, "data-phase-tickets",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("page lacks %q", want)
		}
	}
	// With a description the block is drawn; without one it is left out.
	if !strings.Contains(s, "${p.description ? `<div class=\"phase-desc\">${paragraphs(p.description)}</div>` : ''}") {
		t.Error("the description block is not conditional on p.description")
	}
	// Sidebar and Overview clicks open the page; they no longer filter Tickets.
	if strings.Contains(s, "openPhase(ph.dataset.phase, true)") {
		t.Error("the sidebar still passes a toggle to openPhase")
	}
	if strings.Contains(block(t, s, "function openPhase"), "setGrouped") {
		t.Error("openPhase still groups and filters Tickets")
	}
	// The phase route is not a tab: no nav button, not in VIEWS.
	if strings.Contains(s, `data-tab="phase"`) || regexp.MustCompile(`(?m)^\s*phase: \{ label:`).MatchString(block(t, s, "const VIEWS =")) {
		t.Error("phase must be a route, not a tab")
	}
}
