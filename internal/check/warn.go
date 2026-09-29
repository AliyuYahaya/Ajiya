package check

// Warnings: things that are probably wrong but do not make numbers wrong.
// 'ajiya check --strict' fails on them.
//
//	W001 a phase has no tickets
//	W002 a phase's tickets are all for one app, though its goal names more
//	W003 a ticket was done in a commit that changed nothing under its app
//	W004 a commit changed files outside every registered app
//	W005 a folder with a manifest is not a registered app
//	W006 an open ticket added after the plan was first built is linked to nothing
//	W007 the launch milestone's required set has no deployment or hosting ticket
//	W008 an open ticket has needed a human for more than 14 days
//	W009 an earlier milestone already requires everything a later one does
//	W010 open tickets are in no milestone, when a [[milestones]] list exists
//	W011 [phases] order lists a phase before a phase it waits on

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/detect"
	"github.com/AliyuYahaya/Ajiya/internal/gitx"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// HumanWait is how long a ticket may need a human before W008.
const HumanWait = 14 * 24 * time.Hour

var deployRE = regexp.MustCompile(`(?i)\b(deploy\w*|hosting|hosted|production|go-live|go live|launch|release)\b`)

func warning(code, loc, fix, format string, a ...any) Finding {
	return Finding{Code: code, Level: Warning, Location: loc, Message: fmt.Sprintf(format, a...), Fix: fix}
}

// planWarnings are the warnings that need only the plan and the files on disk.
func planWarnings(cfg *config.Config, p *plan.Plan) []Finding {
	var fs []Finding
	for _, ph := range p.Phases {
		loc := plan.Path(ph.Slug) + ":1"
		if len(ph.Tickets) == 0 {
			fs = append(fs, warning("W001", loc, fmt.Sprintf("ajiya ticket add --phase %s ...", ph.Slug),
				"phase %s has no tickets", ph.Slug))
			continue
		}
		named := appsNamedIn(cfg, ph.Goal)
		used := map[string]bool{}
		for _, t := range ph.Tickets {
			used[t.App] = true
		}
		if len(named) > 1 && len(used) == 1 {
			var only string
			for a := range used {
				only = a
			}
			fs = append(fs, warning("W002", loc, "add tickets for the other apps, or reword the goal",
				"the goal of %s names %s, but every ticket is for %s", ph.Slug, strings.Join(named, " and "), only))
		}
	}

	if cfg.Root != "" {
		dirs, _ := detect.AppManifestDirs(cfg.Root)
		for _, d := range dirs {
			if !underAnyApp(cfg, d) {
				fs = append(fs, warning("W005", d, fmt.Sprintf("ajiya app add %s --path %s (add --library for a shared package)", suggestName(d, cfg), d),
					"%s has its own manifest but is not a registered app", d))
			}
		}
	}

	return append(fs, milestoneWarnings(cfg, p)...)
}

// MaxListed is how many ticket IDs one finding lists before "and N more".
const MaxListed = 10

// milestoneWarnings are W007, W009 and W010. They are skipped while a target
// does not exist: that is E012, and the required sets are not known.
func milestoneWarnings(cfg *config.Config, p *plan.Plan) []Finding {
	var ms []plan.Milestone
	for _, m := range cfg.MilestoneList() {
		ms = append(ms, plan.Milestone{Name: m.Name, Targets: m.Targets})
	}
	if len(ms) == 0 {
		return nil
	}
	s, err := p.Schedule(ms)
	if err != nil {
		return nil
	}
	var fs []Finding
	tickets := p.Tickets()

	// W007: only the milestone named launch needs a deployment ticket.
	if i := s.Index(config.LaunchMilestone); i >= 0 {
		found := false
		for _, t := range tickets {
			if s.Required[i][t.ID] && deployRE.MatchString(t.Title+" "+t.DoneWhen) {
				found = true
				break
			}
		}
		if !found {
			fs = append(fs, warning("W007", config.FileName, "add a ticket for deployment or hosting that the launch target depends on",
				"nothing required for launch mentions deployment or hosting"))
		}
	}

	// W009: a required set is closed under dependencies, so when an earlier
	// milestone requires every ticket of a later one's targets, it requires
	// the later one's whole set, and the later gate opens no later. Each
	// later milestone is named once, against the first such earlier one.
	for j := range ms {
		if len(s.Required[j]) == 0 {
			continue
		}
		for i := range j {
			if subset(s.Required[j], s.Required[i]) {
				a, b := ms[i].Name, ms[j].Name
				fs = append(fs, warning("W009", config.FileName,
					fmt.Sprintf("put %s first with 'ajiya milestone move %s --before %s', or give %s targets that %s does not require", b, b, a, b, a),
					"milestone %s comes after %s, but %s already requires every ticket %s does, so %s is reached no later than %s",
					b, a, a, b, b, a))
				break
			}
		}
	}

	// W010: open work that no milestone requires. Only for a [[milestones]]
	// list: with the older [launch] target alone, work after launch is the
	// expected state ('launch show' counts it), not a problem.
	var loose []string
	for _, t := range tickets {
		if _, ok := s.Of[t.ID]; !ok && !t.Status.Closed() && len(cfg.Milestones) > 0 {
			loose = append(loose, t.ID)
		}
	}
	if n := len(loose); n > 0 {
		more := ""
		if n > MaxListed {
			more = fmt.Sprintf(" and %d more", n-MaxListed)
			loose = loose[:MaxListed]
		}
		noun := "tickets are"
		if n == 1 {
			noun = "ticket is"
		}
		fs = append(fs, warning("W010", config.FileName, "make a milestone's targets depend on them, or add a milestone for the later work",
			"%d open %s in no milestone: %s%s", n, noun, strings.Join(loose, ", "), more))
	}
	return fs
}

// subset reports whether every key of a is in b.
func subset(a, b map[string]bool) bool {
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

// appsNamedIn returns the registered apps whose names appear as words in text.
func appsNamedIn(cfg *config.Config, text string) []string {
	var names []string
	for _, a := range cfg.Apps {
		if regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(a.Name) + `\b`).MatchString(text) {
			names = append(names, a.Name)
		}
	}
	sort.Strings(names)
	return names
}

func underAnyApp(cfg *config.Config, p string) bool {
	for _, a := range cfg.Apps {
		if detect.Under(p, a.Path) {
			return true
		}
	}
	return false
}

// suggestName turns a folder into an app name: its last part, in lower case.
func suggestName(dir string, cfg *config.Config) string {
	name := strings.ToLower(path.Base(dir))
	name = regexp.MustCompile(`[^a-z0-9._-]+`).ReplaceAllString(name, "-")
	if name == "." || name == "" || name == config.InfraApp || cfg.HasApp(name) {
		return "<name>"
	}
	return name
}

// exemptFromApps reports whether a changed file may be outside every app:
// files at the root, dot folders such as .github, and ajiya/ itself.
func exemptFromApps(f string) bool {
	return !strings.Contains(f, "/") || strings.HasPrefix(f, ".") || strings.HasPrefix(f, plan.Dir+"/")
}

// History returns the warnings that need git history: W003, W006 and W008.
// now is passed in so tests can fix it.
func History(cfg *config.Config, p *plan.Plan, now time.Time) ([]Finding, error) {
	root := cfg.Root
	var fs []Finding
	tickets := p.Tickets()
	loc := func(t *plan.Ticket) string { return fmt.Sprintf("%s:%d", plan.Path(t.Phase), t.Line) }

	// W003: the evidence commit should touch the ticket's app.
	files := map[string][]string{}
	for _, t := range tickets {
		app := appByName(cfg, t.App)
		if t.Status.State != plan.Done || t.Status.Commit == "" || app == nil || app.Path == "." {
			continue
		}
		changed, ok := files[t.Status.Commit]
		if !ok {
			var err error
			if changed, err = gitx.CommitFiles(root, t.Status.Commit); err != nil {
				changed = nil // a commit that is gone is not this check's business
			}
			files[t.Status.Commit] = changed
		}
		touched := false
		for _, f := range changed {
			touched = touched || detect.Under(f, app.Path)
		}
		if !touched && changed != nil {
			fs = append(fs, warning("W003", loc(t), "check that the trailer names the right ticket",
				"%s was done in commit %s, which changed nothing under %s", t.ID, t.Status.Commit, app.Path))
		}
	}

	// W006: a ticket added later that nothing links to, in either direction,
	// is easy to leave out of the launch set by mistake.
	g := plan.NewGraph(tickets)
	if first, err := gitx.FirstCommit(root, plan.Dir); err != nil {
		if !errors.Is(err, gitx.ErrNotRepo) {
			return nil, err
		}
	} else if first != "" {
		high := highestIDAt(root, first, cfg.Project.Prefix)
		for _, t := range tickets {
			_, n, ok := plan.SplitID(t.ID)
			if ok && n > high && len(t.Depends) == 0 && len(g.Dependants(t.ID)) == 0 && !t.Status.Closed() {
				fs = append(fs, warning("W006", loc(t), fmt.Sprintf("link it: ajiya ticket edit %s --depends <IDs>, or add it to the --depends of the work that needs it", t.ID),
					"%s was added after the plan was first built and is linked to no other ticket, so it may be missing from the launch set", t.ID))
			}
		}
	}

	// W008: a person has been needed for too long.
	for _, t := range tickets {
		if t.Status.Human == "" || t.Status.Closed() || t.Line == 0 {
			continue
		}
		since, err := gitx.LineTime(root, plan.Path(t.Phase), t.Line)
		if err != nil || since == 0 {
			continue
		}
		if waited := now.Sub(time.Unix(since, 0)); waited > HumanWait {
			fs = append(fs, warning("W008", loc(t), "ask the person named, or close it with 'ajiya ticket done --by'",
				"%s has needed a human (%s) for %d days", t.ID, t.Status.Human, int(waited.Hours()/24)))
		}
	}
	return fs, nil
}

func appByName(cfg *config.Config, name string) *config.App {
	for i := range cfg.Apps {
		if cfg.Apps[i].Name == name {
			return &cfg.Apps[i]
		}
	}
	return nil
}

// highestIDAt returns the highest ticket number in ajiya/ at a commit.
func highestIDAt(root, rev, prefix string) int {
	files, err := gitx.TreeFiles(root, rev, plan.Dir)
	if err != nil {
		return 0
	}
	high := 0
	for _, f := range files {
		slug, ok := strings.CutSuffix(strings.TrimPrefix(f, plan.Dir+"/"), ".md")
		if !ok || !plan.SlugRE.MatchString(slug) {
			continue
		}
		data, ok, err := gitx.Show(root, rev+":"+f)
		if err != nil || !ok {
			continue
		}
		ph, err := plan.ParsePhase(slug, data)
		if err != nil {
			continue
		}
		for _, t := range ph.Tickets {
			if pfx, n, ok := plan.SplitID(t.ID); ok && pfx == prefix && n > high {
				high = n
			}
		}
	}
	return high
}
