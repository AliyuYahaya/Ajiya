package plan

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// orderPlan builds a plan from "slug: ID<DEP,DEP ID ..." specs, with the
// phases sorted by slug as Load leaves them.
func orderPlan(specs ...string) *Plan {
	p := &Plan{}
	for _, s := range specs {
		slug, rest, _ := strings.Cut(s, ":")
		ph := &Phase{Slug: slug}
		for _, f := range strings.Fields(rest) {
			id, deps, _ := strings.Cut(f, "<")
			t := tk(id)
			if deps != "" {
				t.Depends = strings.Split(deps, ",")
			}
			t.Phase = slug
			ph.Tickets = append(ph.Tickets, t)
		}
		p.Phases = append(p.Phases, ph)
	}
	sort.Slice(p.Phases, func(i, j int) bool { return p.Phases[i].Slug < p.Phases[j].Slug })
	return p
}

func TestSuggestOrder(t *testing.T) {
	tests := []struct {
		name       string
		phases     []string
		milestones []Milestone
		want       string
		cycles     string
	}{
		{"dependency beats creation",
			[]string{"a: CB-0002", "b: CB-0001<CB-0002"}, nil,
			"a b", ""},
		{"ties by creation, empty phases last by slug",
			[]string{"x: CB-0003", "y: CB-0001", "zz:", "za:", "w: CB-0002<CB-0003"}, nil,
			"y x w za zz", ""},
		{"milestone tie-break before creation",
			[]string{"p: CB-0001", "q: CB-0002", "r: CB-0003"},
			[]Milestone{{"first", []string{"r"}}, {"second", []string{"q"}}},
			"r q p", ""},
		{"a milestone never breaks a dependency",
			[]string{"base: CB-0005", "top: CB-0001<CB-0005", "other: CB-0002"},
			[]Milestone{{"first", []string{"other"}}, {"second", []string{"top"}}},
			"other base top", ""},
		{"phases that wait on each other stay together in creation order",
			[]string{"a: CB-0003<CB-0005", "b: CB-0005 CB-0006<CB-0003", "c: CB-0007<CB-0006", "d: CB-0001", "e: CB-0008"}, nil,
			"d a b c e", "a b"},
		{"a cycle group is placed by its earliest milestone",
			[]string{"a: CB-0001", "m: CB-0002<CB-0003", "n: CB-0003<CB-0002"},
			[]Milestone{{"first", []string{"CB-0003"}}},
			"m n a", "m n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := orderPlan(tt.phases...)
			var sched *Schedule
			if tt.milestones != nil {
				var err error
				if sched, err = p.Schedule(tt.milestones); err != nil {
					t.Fatal(err)
				}
			}
			order, cycles := p.SuggestOrder(sched)
			if got := strings.Join(order, " "); got != tt.want {
				t.Errorf("order = %s, want %s", got, tt.want)
			}
			var cs []string
			for _, c := range cycles {
				cs = append(cs, strings.Join(c, " "))
			}
			if got := strings.Join(cs, "; "); got != tt.cycles {
				t.Errorf("cycles = %q, want %q", got, tt.cycles)
			}
		})
	}
}

func TestPhaseWaits(t *testing.T) {
	p := orderPlan("a: CB-0001", "b: CB-0002<CB-0001 CB-0003<CB-0001,CB-0099 CB-0004<CB-0003")
	got := fmt.Sprint(p.PhaseWaits())
	if want := "map[b:map[a:{CB-0002 CB-0001}]]"; got != want {
		t.Errorf("PhaseWaits = %s, want %s", got, want)
	}
}

func TestPhaseMilestone(t *testing.T) {
	p := orderPlan("a: CB-0001", "b: CB-0002<CB-0001 CB-0003")
	s, err := p.Schedule([]Milestone{{"one", []string{"CB-0002"}}, {"two", []string{"b"}}})
	if err != nil {
		t.Fatal(err)
	}
	for slug, want := range map[string]string{"a": "one 0", "b": "one 0"} {
		if name, i := PhaseMilestone(p.Phase(slug), s); fmt.Sprint(name, " ", i) != want {
			t.Errorf("%s: %s %d, want %s", slug, name, i, want)
		}
	}
	s, _ = p.Schedule([]Milestone{{"one", []string{"CB-0003"}}})
	if name, i := PhaseMilestone(p.Phase("a"), s); name != "" || i != -1 {
		t.Errorf("unscheduled: %q %d", name, i)
	}
	if name, i := PhaseMilestone(p.Phase("a"), nil); name != "" || i != -1 {
		t.Errorf("nil schedule: %q %d", name, i)
	}
}

func TestDisplayOrder(t *testing.T) {
	p := orderPlan("a: CB-0001", "b: CB-0002<CB-0001", "c: CB-0003", "d:")
	tests := []struct {
		listed []string
		want   string
	}{
		{nil, "a b c d"},
		{[]string{"c"}, "c a b d"},
		{[]string{"b", "gone", "a"}, "b a c d"},
		{[]string{"d", "c", "b", "a"}, "d c b a"},
		{[]string{"c", "c"}, "c a b d"},
	}
	for _, tt := range tests {
		var got []string
		for _, ph := range p.DisplayOrder(tt.listed, nil) {
			got = append(got, ph.Slug)
		}
		if strings.Join(got, " ") != tt.want {
			t.Errorf("DisplayOrder(%v) = %v, want %s", tt.listed, got, tt.want)
		}
	}
	// Plan.Phases itself stays sorted by slug.
	if p.Phases[0].Slug != "a" || p.Phases[3].Slug != "d" {
		t.Error("Phases reordered")
	}
}
