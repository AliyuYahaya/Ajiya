package plan

import "fmt"

// Milestone is a named gate whose targets are phase slugs or ticket IDs.
type Milestone struct {
	Name    string
	Targets []string
}

// Required returns the tickets required for one target: the target's own
// tickets (a phase's, or the ticket itself) plus everything they depend on,
// directly or not. kind is "phase" or "ticket".
func (p *Plan) Required(target string) (ids map[string]bool, kind string, err error) {
	ids = map[string]bool{}
	kind, err = p.addRequired(ids, NewGraph(p.Tickets()), target)
	return ids, kind, err
}

// RequiredAll is Required for a list of targets.
func (p *Plan) RequiredAll(targets []string) (map[string]bool, error) {
	ids := map[string]bool{}
	g := NewGraph(p.Tickets())
	for _, t := range targets {
		if _, err := p.addRequired(ids, g, t); err != nil {
			return nil, err
		}
	}
	return ids, nil
}

func (p *Plan) addRequired(ids map[string]bool, g *Graph, target string) (kind string, err error) {
	var roots []string
	if ph := p.Phase(target); ph != nil {
		kind = "phase"
		for _, t := range ph.Tickets {
			roots = append(roots, t.ID)
		}
	} else if t := p.Ticket(target); t != nil {
		kind = "ticket"
		roots = []string{t.ID}
	} else {
		return "", fmt.Errorf("target %q is neither a phase nor a ticket", target)
	}
	for _, r := range roots {
		ids[r] = true
		for _, u := range g.Upstream(r) {
			ids[u] = true
		}
	}
	return kind, nil
}

// Schedule is the milestones worked out against the plan.
type Schedule struct {
	Milestones []Milestone
	Required   []map[string]bool // per milestone, in order
	// Of maps each scheduled ticket to the earliest milestone that requires
	// it. Tickets missing from it are unscheduled.
	Of map[string]string
}

// Schedule works out each milestone's required set and the earliest
// milestone of every ticket. It fails on a target that does not exist.
func (p *Plan) Schedule(ms []Milestone) (*Schedule, error) {
	s := &Schedule{Milestones: ms, Of: map[string]string{}}
	for _, m := range ms {
		req, err := p.RequiredAll(m.Targets)
		if err != nil {
			return nil, fmt.Errorf("milestone %s: %w", m.Name, err)
		}
		s.Required = append(s.Required, req)
		for id := range req {
			if _, earlier := s.Of[id]; !earlier {
				s.Of[id] = m.Name
			}
		}
	}
	return s, nil
}

// Index returns the position of a milestone by name, or -1.
func (s *Schedule) Index(name string) int {
	for i, m := range s.Milestones {
		if m.Name == name {
			return i
		}
	}
	return -1
}
