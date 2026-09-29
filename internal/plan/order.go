package plan

import (
	"slices"
	"sort"
)

// The display order of phases is for reading only: it never changes what can
// start, what next returns or what a milestone requires.

// PhaseLink is one ticket that makes a phase wait on another: Ticket depends
// on Dep, which is in the other phase.
type PhaseLink struct {
	Ticket, Dep string
}

// PhaseWaits returns, for each phase, the phases it waits on directly: those
// holding a ticket that one of its tickets depends on. Each comes with the
// lowest link that shows it. A phase never waits on itself.
func (p *Plan) PhaseWaits() map[string]map[string]PhaseLink {
	phaseOf := map[string]string{}
	for _, ph := range p.Phases {
		for _, t := range ph.Tickets {
			if _, dup := phaseOf[t.ID]; !dup {
				phaseOf[t.ID] = ph.Slug
			}
		}
	}
	waits := map[string]map[string]PhaseLink{}
	for _, t := range p.Tickets() {
		for _, d := range t.Depends {
			on, ok := phaseOf[d]
			if !ok || on == t.Phase {
				continue
			}
			if waits[t.Phase] == nil {
				waits[t.Phase] = map[string]PhaseLink{}
			}
			if _, seen := waits[t.Phase][on]; !seen { // tickets come sorted by ID
				waits[t.Phase][on] = PhaseLink{t.ID, d}
			}
		}
	}
	return waits
}

// PhaseGroups returns the groups of phases that wait on each other, directly
// or not, keyed by slug: phases in the same group share a number. A phase
// alone is a group of one.
func (p *Plan) PhaseGroups() map[string]int {
	waits := p.PhaseWaits()
	index, low := map[string]int{}, map[string]int{}
	onStack := map[string]bool{}
	var stack []string
	group := map[string]int{}
	n, g := 0, 0
	var visit func(v string)
	visit = func(v string) {
		index[v], low[v] = n, n
		n++
		stack = append(stack, v)
		onStack[v] = true
		for _, w := range sortedKeys(waits[v]) {
			if _, seen := index[w]; !seen {
				visit(w)
				low[v] = min(low[v], low[w])
			} else if onStack[w] {
				low[v] = min(low[v], index[w])
			}
		}
		if low[v] == index[v] {
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				group[w] = g
				if w == v {
					break
				}
			}
			g++
		}
	}
	for _, ph := range p.Phases {
		if _, seen := index[ph.Slug]; !seen {
			visit(ph.Slug)
		}
	}
	return group
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// PhaseMilestone returns the earliest milestone of any ticket in the phase and
// its index in the schedule, or "" and -1 when none of its tickets is
// scheduled (or sched is nil).
func PhaseMilestone(ph *Phase, sched *Schedule) (string, int) {
	if sched == nil {
		return "", -1
	}
	best := -1
	for _, t := range ph.Tickets {
		if m, ok := sched.Of[t.ID]; ok {
			if i := sched.Index(m); i >= 0 && (best < 0 || i < best) {
				best = i
			}
		}
	}
	if best < 0 {
		return "", -1
	}
	return sched.Milestones[best].Name, best
}

// lessCreated orders phases by when they were created: the one holding the
// lowest ticket ID first, phases with no tickets last, then by slug.
func lessCreated(a, b *Phase) bool {
	fa, fb := firstID(a), firstID(b)
	switch {
	case (fa == "") != (fb == ""):
		return fa != ""
	case fa != fb:
		return LessID(fa, fb)
	}
	return a.Slug < b.Slug
}

func firstID(ph *Phase) string {
	first := ""
	for _, t := range ph.Tickets {
		if first == "" || LessID(t.ID, first) {
			first = t.ID
		}
	}
	return first
}

// SuggestOrder works out an order of the phases from the dependencies: each
// phase after the phases it waits on. Ties go to the earliest milestone
// (phases with no scheduled ticket last), then to the phase created first.
// Phases that wait on each other are kept next to each other, in creation
// order, and returned in cycles, each in that order. sched may be nil.
func (p *Plan) SuggestOrder(sched *Schedule) (order []string, cycles [][]string) {
	waits := p.PhaseWaits()
	groupOf := p.PhaseGroups()

	// Each group: its phases in creation order and its sort key.
	type group struct {
		phases    []*Phase
		milestone int // earliest milestone index; unscheduled sorts last
		waitsOn   map[int]bool
		waiting   int // groups it waits on that are not yet placed
	}
	groups := map[int]*group{}
	for _, ph := range p.Phases {
		id := groupOf[ph.Slug]
		gr := groups[id]
		if gr == nil {
			gr = &group{milestone: -1, waitsOn: map[int]bool{}}
			groups[id] = gr
		}
		gr.phases = append(gr.phases, ph)
		if _, i := PhaseMilestone(ph, sched); i >= 0 && (gr.milestone < 0 || i < gr.milestone) {
			gr.milestone = i
		}
		for on := range waits[ph.Slug] {
			if og, ok := groupOf[on]; ok && og != id {
				gr.waitsOn[og] = true
			}
		}
	}
	for _, gr := range groups {
		sort.SliceStable(gr.phases, func(i, j int) bool { return lessCreated(gr.phases[i], gr.phases[j]) })
		gr.waiting = len(gr.waitsOn)
	}
	less := func(a, b *group) bool {
		ma, mb := a.milestone, b.milestone
		if ma < 0 {
			ma = int(^uint(0) >> 1)
		}
		if mb < 0 {
			mb = int(^uint(0) >> 1)
		}
		if ma != mb {
			return ma < mb
		}
		return lessCreated(a.phases[0], b.phases[0])
	}

	// Kahn's algorithm, always taking the smallest ready group.
	var ready []int
	for id, gr := range groups {
		if gr.waiting == 0 {
			ready = append(ready, id)
		}
	}
	for len(ready) > 0 {
		best := 0
		for i := range ready {
			if less(groups[ready[i]], groups[ready[best]]) {
				best = i
			}
		}
		id := ready[best]
		ready = slices.Delete(ready, best, best+1)
		gr := groups[id]
		var slugs []string
		for _, ph := range gr.phases {
			slugs = append(slugs, ph.Slug)
		}
		order = append(order, slugs...)
		if len(slugs) > 1 {
			cycles = append(cycles, slugs)
		}
		for oid, other := range groups {
			if other.waitsOn[id] {
				other.waiting--
				if other.waiting == 0 {
					ready = append(ready, oid)
				}
			}
		}
	}
	return order, cycles
}

// DisplayOrder returns the phases in the order to show them: the slugs in
// listed first, in that order, then the other phases in suggested order.
// Listed slugs that are not phases are skipped. sched may be nil.
func (p *Plan) DisplayOrder(listed []string, sched *Schedule) []*Phase {
	var out []*Phase
	placed := map[string]bool{}
	for _, slug := range listed {
		if ph := p.Phase(slug); ph != nil && !placed[slug] {
			placed[slug] = true
			out = append(out, ph)
		}
	}
	if len(out) == len(p.Phases) {
		return out
	}
	suggested, _ := p.SuggestOrder(sched)
	for _, slug := range suggested {
		if !placed[slug] {
			out = append(out, p.Phase(slug))
		}
	}
	return out
}
