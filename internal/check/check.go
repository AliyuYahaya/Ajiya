// Package check finds problems in a plan. Each finding has a stable code, a
// location and a one-line fix.
//
// Codes are permanent: a code is never reused for a different problem.
//
//	E001 a phase file does not parse
//	E002 a phase file is not in canonical form (edited by hand)
//	E003 an ID does not match the project prefix and format
//	E004 an ID is used by more than one ticket
//	E005 a dependency on an ID that does not exist
//	E006 a ticket depends on itself
//	E007 a dependency loop
//
// Codes for later checks (apps, evidence, launch, commits and warnings) are
// added as those checks are built.
package check

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

type Level string

const (
	Error   Level = "error"
	Warning Level = "warning"
)

// Finding is one problem.
type Finding struct {
	Code     string `json:"code"`
	Level    Level  `json:"level"`
	Location string `json:"location"` // ajiya/<slug>.md:<line>
	Message  string `json:"message"`
	Fix      string `json:"fix"`
}

func (f Finding) String() string {
	return fmt.Sprintf("%s %s: %s. Fix: %s", f.Code, f.Location, f.Message, f.Fix)
}

// Run runs every check and returns the findings sorted by location and code.
func Run(cfg *config.Config, p *plan.Plan) []Finding {
	var fs []Finding
	add := func(code, loc, fix, format string, a ...any) {
		level := Error
		if strings.HasPrefix(code, "W") {
			level = Warning
		}
		fs = append(fs, Finding{Code: code, Level: level, Location: loc, Message: fmt.Sprintf(format, a...), Fix: fix})
	}
	restore := func(slug string) string {
		return fmt.Sprintf("restore it with 'git checkout -- %s' and make the change with ajiya commands", plan.Path(slug))
	}

	for slug, pe := range p.Broken {
		add("E001", fmt.Sprintf("%s:%d", plan.Path(slug), pe.Line), restore(slug), "%s", pe.Msg)
	}
	for _, ph := range p.Phases {
		if line := firstDiff(p.Raw[ph.Slug], ph.Format()); line > 0 {
			add("E002", fmt.Sprintf("%s:%d", plan.Path(ph.Slug), line), restore(ph.Slug),
				"file is not in the form ajiya writes (edited by hand?)")
		}
	}

	tickets := p.Tickets()
	loc := func(t *plan.Ticket) string { return fmt.Sprintf("%s:%d", plan.Path(t.Phase), t.Line) }
	byID := map[string]*plan.Ticket{}
	for _, t := range tickets {
		if !plan.ValidID(cfg.Project.Prefix, t.ID) {
			add("E003", loc(t), "restore the row from git; IDs are only assigned by 'ajiya ticket add'",
				"ID %q does not match the form %s", t.ID, plan.FormatID(cfg.Project.Prefix, 1))
		}
		if first, dup := byID[t.ID]; dup {
			add("E004", loc(t), "restore the phase files from git; IDs are only assigned by 'ajiya ticket add'",
				"ID %s is also used at %s", t.ID, loc(first))
			continue
		}
		byID[t.ID] = t
	}
	for _, t := range tickets {
		for _, d := range t.Depends {
			switch {
			case d == t.ID:
				add("E006", loc(t), fmt.Sprintf("ajiya ticket edit %s --depends <IDs without %s>", t.ID, t.ID),
					"%s depends on itself", t.ID)
			case byID[d] == nil:
				add("E005", loc(t), fmt.Sprintf("ajiya ticket edit %s --depends <existing IDs>", t.ID),
					"%s depends on %s, which does not exist", t.ID, d)
			}
		}
	}
	for _, c := range plan.NewGraph(tickets).Cycles() {
		t := byID[c[0]]
		add("E007", loc(t), fmt.Sprintf("remove one link, for example 'ajiya ticket edit %s --depends ...'", c[len(c)-2]),
			"dependency loop: %s", strings.Join(c, " -> "))
	}

	sort.SliceStable(fs, func(i, j int) bool {
		if fs[i].Location != fs[j].Location {
			return lessLocation(fs[i].Location, fs[j].Location)
		}
		return fs[i].Code < fs[j].Code
	})
	return fs
}

// Count returns the number of errors and warnings.
func Count(fs []Finding) (errors, warnings int) {
	for _, f := range fs {
		if f.Level == Error {
			errors++
		} else {
			warnings++
		}
	}
	return
}

// firstDiff returns the first line (1-based) where a and b differ, or 0.
func firstDiff(a, b []byte) int {
	if bytes.Equal(a, b) {
		return 0
	}
	la, lb := strings.Split(string(a), "\n"), strings.Split(string(b), "\n")
	for i := range min(len(la), len(lb)) {
		if la[i] != lb[i] {
			return i + 1
		}
	}
	return min(len(la), len(lb))
}

// lessLocation orders file:line locations by file, then line number.
func lessLocation(a, b string) bool {
	fa, la := splitLoc(a)
	fb, lb := splitLoc(b)
	if fa != fb {
		return fa < fb
	}
	return la < lb
}

func splitLoc(s string) (string, int) {
	i := strings.LastIndex(s, ":")
	if i < 0 {
		return s, 0
	}
	n := 0
	fmt.Sscanf(s[i+1:], "%d", &n)
	return s[:i], n
}
