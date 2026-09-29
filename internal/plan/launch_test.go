package plan

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

// schedulePlan: gate-0 has 1 and 2 (2 depends on 1); go-live has 4, which
// depends on 2 and 3; reporting has 5 (depends on 4); 6 stands alone.
func schedulePlan() *Plan {
	ph := func(slug string, ts ...*Ticket) *Phase {
		for _, t := range ts {
			t.Phase = slug
		}
		return &Phase{Slug: slug, Tickets: ts}
	}
	return &Plan{Phases: []*Phase{
		ph("gate-0", tk("CB-0001"), tk("CB-0002", "CB-0001")),
		ph("go-live", tk("CB-0003"), tk("CB-0004", "CB-0002", "CB-0003")),
		ph("reporting", tk("CB-0005", "CB-0004"), tk("CB-0006")),
	}}
}

func keys(m map[string]bool) string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return strings.Join(ks, " ")
}

func TestSchedule(t *testing.T) {
	p := schedulePlan()
	s, err := p.Schedule([]Milestone{
		{"staging-proven", []string{"gate-0"}},
		{"launch", []string{"go-live"}},
		{"after-launch", []string{"CB-0005"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantReq := []string{"CB-0001 CB-0002", "CB-0001 CB-0002 CB-0003 CB-0004", "CB-0001 CB-0002 CB-0003 CB-0004 CB-0005"}
	for i, req := range s.Required {
		if keys(req) != wantReq[i] {
			t.Errorf("milestone %d required = %s, want %s", i, keys(req), wantReq[i])
		}
	}
	// Tickets required by two milestones belong to the earlier one; 6 is unscheduled.
	want := map[string]string{"CB-0001": "staging-proven", "CB-0002": "staging-proven", "CB-0003": "launch", "CB-0004": "launch", "CB-0005": "after-launch"}
	if !reflect.DeepEqual(s.Of, want) {
		t.Errorf("Of = %v, want %v", s.Of, want)
	}
	if s.Index("launch") != 1 || s.Index("nope") != -1 {
		t.Error("Index wrong")
	}
}

func TestScheduleMissingTarget(t *testing.T) {
	_, err := schedulePlan().Schedule([]Milestone{{"launch", []string{"go-live", "nope"}}})
	if err == nil || err.Error() != `milestone launch: target "nope" is neither a phase nor a ticket` {
		t.Errorf("err = %v", err)
	}
}

func TestRequiredAll(t *testing.T) {
	req, err := schedulePlan().RequiredAll([]string{"CB-0003", "CB-0006"})
	if err != nil || keys(req) != "CB-0003 CB-0006" {
		t.Errorf("RequiredAll = %s, %v", keys(req), err)
	}
}
