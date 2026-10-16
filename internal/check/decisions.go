package check

import (
	"fmt"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// Decision codes:
//
//	E014 a ticket links a decision (dec tag) that does not exist
//	E015 a decision file does not parse
//	E016 a decision file is not in canonical form (edited by hand)
//	W014 an open ticket links a superseded decision

// decisionFindings checks the decision files and the tickets that link them.
func decisionFindings(p *plan.Plan) []Finding {
	var fs []Finding
	restore := func(id string) string {
		return fmt.Sprintf("restore it with 'git checkout -- %s' and make the change with ajiya commands", plan.DecisionPath(id))
	}
	for id, pe := range p.BrokenDecisions {
		fs = append(fs, Finding{Code: "E015", Level: Error, Location: fmt.Sprintf("%s:%d", plan.DecisionPath(id), pe.Line),
			Message: pe.Msg, Fix: restore(id)})
	}
	for _, d := range p.Decisions {
		if line := firstDiff(p.RawDecisions[d.ID], d.Format()); line > 0 {
			fs = append(fs, Finding{Code: "E016", Level: Error, Location: fmt.Sprintf("%s:%d", plan.DecisionPath(d.ID), line),
				Message: "file is not in the form ajiya writes (edited by hand?)", Fix: restore(d.ID)})
		}
	}
	for _, t := range p.Tickets() {
		loc := fmt.Sprintf("%s:%d", plan.Path(t.Phase), t.Line)
		for _, id := range t.Tags.Dec {
			d := p.Decision(id)
			switch {
			case d != nil:
				if !t.Status.Closed() && d.Status.State == plan.DecSuperseded {
					fix := fmt.Sprintf("read %s with 'ajiya decision show %s', then retag the ticket with 'ajiya ticket edit %s --decisions <IDs>'", id, id, t.ID)
					if d.Status.SupersededBy != "" {
						fix = fmt.Sprintf("read %s with 'ajiya decision show %s', then retag with 'ajiya ticket edit %s --decisions <IDs>'", d.Status.SupersededBy, d.Status.SupersededBy, t.ID)
					}
					fs = append(fs, warning("W014", loc, fix, "%s links %s, which is %s", t.ID, id, d.Status))
				}
			case p.BrokenDecisions[id] == nil: // a file that does not parse is E015, not missing
				fs = append(fs, Finding{Code: "E014", Level: Error, Location: loc,
					Message: fmt.Sprintf("%s links decision %s, which does not exist", t.ID, id),
					Fix:     fmt.Sprintf("add it with 'ajiya decision add', or correct the tag with 'ajiya ticket edit %s --decisions <existing IDs>|-'", t.ID)})
			}
		}
	}
	return fs
}
