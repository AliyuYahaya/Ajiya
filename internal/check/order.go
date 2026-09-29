package check

import (
	"fmt"
	"sort"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// phaseOrder checks [phases] order: E013 for a slug that is not a phase, and
// W011 for a phase listed before a phase it waits on. Listed phases come
// before every phase left out of the list, so waiting on one of those counts
// too. Phases that wait on each other may be in either order.
func phaseOrder(cfg *config.Config, p *plan.Plan) []Finding {
	var fs []Finding
	listed := map[string]int{}
	for i, slug := range cfg.Phases.Order {
		if _, ok := p.Raw[slug]; !ok {
			fs = append(fs, Finding{Code: "E013", Level: Error, Location: config.FileName,
				Message: fmt.Sprintf("[phases] order lists %q, which is not a phase", slug),
				Fix:     fmt.Sprintf("remove %q from [phases] order in %s, or add it with 'ajiya phase add %s \"<goal>\"'", slug, config.FileName, slug)})
			continue
		}
		listed[slug] = i
	}
	if len(listed) == 0 {
		return fs
	}
	waits := p.PhaseWaits()
	group := p.PhaseGroups()
	for _, a := range cfg.Phases.Order {
		i, ok := listed[a]
		if !ok {
			continue
		}
		for _, b := range sortedSlugs(waits[a]) {
			j, bListed := listed[b]
			if (bListed && j < i) || group[a] == group[b] {
				continue
			}
			link := waits[a][b]
			where := "listed before"
			if !bListed {
				where = "listed, so it comes before"
			}
			fs = append(fs, warning("W011", config.FileName, fmt.Sprintf("ajiya phase move %s --before %s, or 'ajiya phase order --apply'", b, a),
				"phase %s is %s phase %s, which it waits on: %s depends on %s", a, where, b, link.Ticket, link.Dep))
		}
	}
	return fs
}

func sortedSlugs(m map[string]plan.PhaseLink) []string {
	var out []string
	for slug := range m {
		out = append(out, slug)
	}
	sort.Strings(out)
	return out
}
