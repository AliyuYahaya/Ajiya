package plan

import "fmt"

// Required returns the tickets required for launch: the target's own tickets
// (a phase's tickets, or the ticket itself) plus everything they depend on,
// directly or not. kind is "phase" or "ticket".
func (p *Plan) Required(target string) (ids map[string]bool, kind string, err error) {
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
		return nil, "", fmt.Errorf("launch target %q is neither a phase nor a ticket", target)
	}
	g := NewGraph(p.Tickets())
	ids = map[string]bool{}
	for _, r := range roots {
		ids[r] = true
		for _, u := range g.Upstream(r) {
			ids[u] = true
		}
	}
	return ids, kind, nil
}
