package check

import (
	"fmt"
	"time"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// Change codes:
//
//	E018 a change file does not parse
//	E019 a change file is not in canonical form (edited by hand)
//	W015 a change has been open for more than 14 days (see History)

// ChangeWait is how long a change may stay open before W015.
const ChangeWait = 14 * 24 * time.Hour

// changeFindings checks the change files.
func changeFindings(p *plan.Plan) []Finding {
	var fs []Finding
	restore := func(id string) string {
		return fmt.Sprintf("restore it with 'git checkout -- %s' and make the change with ajiya commands", plan.ChangePath(id))
	}
	for id, pe := range p.BrokenChanges {
		fs = append(fs, Finding{Code: "E018", Level: Error, Location: fmt.Sprintf("%s:%d", plan.ChangePath(id), pe.Line),
			Message: pe.Msg, Fix: restore(id)})
	}
	for _, c := range p.Changes {
		if line := firstDiff(p.RawChanges[c.ID], c.Format()); line > 0 {
			fs = append(fs, Finding{Code: "E019", Level: Error, Location: fmt.Sprintf("%s:%d", plan.ChangePath(c.ID), line),
				Message: "file is not in the form ajiya writes (edited by hand?)", Fix: restore(c.ID)})
		}
	}
	return fs
}

// openChangeWarnings is W015: a change open for more than 14 days. Open
// changes hold up the plan, so one that nobody closes is probably forgotten.
func openChangeWarnings(p *plan.Plan, now time.Time) []Finding {
	var fs []Finding
	for _, c := range p.Changes {
		if !c.Status.Open() {
			continue
		}
		opened, err := time.Parse("2006-01-02", c.Status.Date)
		if err != nil {
			continue
		}
		// Whole days, so the 14th day is not yet late whatever the time of day.
		if days := int(now.Sub(opened) / (24 * time.Hour)); days > int(ChangeWait/(24*time.Hour)) {
			fs = append(fs, warning("W015", plan.ChangePath(c.ID)+":1",
				fmt.Sprintf("show the person 'ajiya change show %s', then a person runs 'ajiya change approve' or 'ajiya change reject'", c.ID),
				"%s has been open for %d days", c.ID, days))
		}
	}
	return fs
}
