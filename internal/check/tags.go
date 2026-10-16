package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// oldFormWarning is W012. Any ajiya command that saves the phase upgrades it,
// so the fix is a no-op edit of its first ticket.
func oldFormWarning(ph *plan.Phase) Finding {
	fix := fmt.Sprintf("ajiya ticket add --phase %s ... (any command that changes this phase rewrites it with the Tags column)", ph.Slug)
	if len(ph.Tickets) > 0 {
		t := ph.Tickets[0]
		fix = fmt.Sprintf("run 'ajiya ticket edit %s --app %s': any command that changes this phase rewrites it with the Tags column", t.ID, t.App)
	}
	return warning("W012", plan.Path(ph.Slug)+":1", fix,
		"%s is in the old six-column form, without the Tags column", plan.Path(ph.Slug))
}

// techKey reduces a tech tag to what postgres and postgresql, node and nodejs,
// aws-rds and awsrds share: no punctuation, and a trailing ql, then js, then s
// cut (postgresql loses ql and then s, which postgres loses too).
func techKey(v string) string {
	k := strings.ReplaceAll(v, "-", "")
	for _, suffix := range []string{"ql", "js", "s"} {
		if len(k) > len(suffix) && strings.HasSuffix(k, suffix) {
			k = strings.TrimSuffix(k, suffix)
		}
	}
	return k
}

// techWarnings is W013: one finding for each group of tech values that look
// like the same technology spelled two ways. It names the most used spelling.
func techWarnings(p *plan.Plan) []Finding {
	type use struct {
		n     int
		first *plan.Ticket
	}
	uses := map[string]*use{}
	for _, t := range p.Tickets() {
		for _, v := range t.Tags.Tech {
			if uses[v] == nil {
				uses[v] = &use{first: t}
			}
			uses[v].n++
		}
	}
	groups := map[string][]string{}
	for v := range uses {
		k := techKey(v)
		groups[k] = append(groups[k], v)
	}
	var fs []Finding
	for _, vs := range groups {
		if len(vs) < 2 {
			continue
		}
		sort.Slice(vs, func(i, j int) bool {
			if uses[vs[i]].n != uses[vs[j]].n {
				return uses[vs[i]].n > uses[vs[j]].n
			}
			return vs[i] < vs[j]
		})
		var first *plan.Ticket // the first ticket, by ID, using any of them
		for _, v := range vs {
			if t := uses[v].first; first == nil || plan.LessID(t.ID, first.ID) {
				first = t
			}
		}
		fs = append(fs, warning("W013", fmt.Sprintf("%s:%d", plan.Path(first.Phase), first.Line),
			fmt.Sprintf("use one spelling, such as tech:%s, with 'ajiya ticket edit <ID> --tech <value>'", vs[0]),
			"tech tags %s look like the same technology", strings.Join(vs, " and ")))
	}
	return fs
}
