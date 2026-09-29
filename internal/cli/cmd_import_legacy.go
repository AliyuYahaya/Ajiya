package cli

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/importer"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

const (
	archivePhase = "archive"
	archiveGoal  = "Done before Ajiya, not in any rollout work package"
	maxSlugLen   = 40
	importedBy   = "ajiya import"
)

// runImportLegacy imports a project tracked in the collate.py format (see
// package importer for the config file and the parsing rules).
//
// Mapping, with a rollout WBS (the RO- work packages):
//   - Each RO package becomes a phase: the slug is the RO number and the
//     first words of its name (ro-021-production-host-and-deploy, at most 40
//     characters), the title is the name as written and the goal is the name
//     ending "(was RO-021)". The tickets its scope counts move into it, in
//     file order. A ticket named by several packages goes into the first in
//     file order, and the report lists the others.
//   - Each row becomes a ticket with a new ID, the title as written, Done when
//     "-", and its old ID kept as an alias ("Was:"). 🟩 is Done with "Done
//     before Ajiya" as evidence, 🟨 In progress, 🟥 or no mark Pending. Rows
//     sharing an old ID each get their own ticket and are listed in the
//     report. A ticket keeps its area as its app: the app key of its source,
//     or --app (default infra) for a source without one. Every app must be
//     registered.
//   - A cross-reference (a ticket "tracked as" another ticket in the same
//     package) is imported as Dropped, "tracked as <new ID> (was <old ID>)",
//     decided by "ajiya import", so the work counts once. The report lists
//     them so a person can reopen any of them.
//   - A package with no scope (user acceptance, cutover) becomes a phase with
//     one ticket: the RO name, with its typed status and the RO ID as an
//     alias. A pending one is marked Needs a human. A package whose scope is
//     empty after earlier packages took its tickets is treated the same way.
//   - When package A depends on package B, each root ticket of A (nothing
//     inside A depends on it) depends on each sink ticket of B (nothing
//     inside B depends on it), so every ticket in A waits for all of B
//     through the chain. A link that would make a loop is left out and
//     reported, as are scope entries and dependencies that name nothing.
//   - Open tickets in no package go to the phase inbox, done ones to the
//     phase archive.
//   - Each "##" heading of the rollout WBS that has packages becomes a
//     milestone, in file order, named from the heading ("Gate 0: Staging
//     proven" is gate-0) and targeting the phases of the packages under it.
//   - The display order of the phases is the order of the packages, then
//     inbox and archive.
//
// Without a rollout WBS, each source WBS becomes one phase (the slug of its
// area, or its phase key, titled with the area name): epics in these files
// are often small or missing, and agents reorganise the phases after the
// import. There are no milestones.
//
// There is no RO ticket, so 'ajiya ticket show RO-021' finds nothing; the
// report (and --json) maps each RO ID to its phase.
//
// The import refuses to run when an old ticket ID it would add is already an
// alias in the plan: the project was imported before, and doing it again
// would duplicate every ticket. RO IDs are not compared (unless there are no
// ticket rows), so two legacy projects with their own prefixes can both be
// imported.
func runImportLegacy(e *env, args []string) error {
	fs := newFlags("import legacy")
	app := fs.String("app", config.InfraApp, "app for the imported tickets")
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	path := pos[0]
	if !filepath.IsAbs(path) {
		path = filepath.Join(e.dir, path)
	}
	lc, err := importer.LoadLegacyConfig(path)
	if err != nil {
		return usageErr("%v", err)
	}
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	if err := pr.checkApp(*app); err != nil {
		return err
	}
	for _, s := range lc.Sources {
		if s.App != "" {
			if err := pr.checkApp(s.App); err != nil {
				return refused("source %s: %v", s.Area, err)
			}
		}
	}
	lp, err := importer.ReadLegacy(lc)
	if err != nil {
		return usageErr("%v", err)
	}

	imported := map[string]string{} // alias -> ticket that has it
	for _, t := range pr.plan.Tickets() {
		for _, a := range t.Status.Aliases {
			if _, ok := imported[a]; !ok {
				imported[a] = t.ID
			}
		}
	}
	// Ticket IDs identify the project; RO IDs only when it has no tickets,
	// since every rollout WBS starts at RO-1.
	var olds []string
	for _, t := range lp.Tickets {
		olds = append(olds, t.ID)
	}
	if len(olds) == 0 {
		for _, p := range lp.Packages {
			olds = append(olds, p.ID)
		}
	}
	for _, old := range olds {
		if id, ok := imported[old]; ok {
			return refused("%s was already imported as %s; importing this project again would duplicate its tickets", old, id)
		}
	}

	im := &legacyImport{pr: pr, lc: lc, lp: lp, app: *app, notes: slices.Clone(lp.Notes)}
	im.rows = make([]*plan.Ticket, len(lp.Tickets))
	if lc.Rollout != "" {
		im.byPackage()
	} else {
		im.byArea()
	}
	for _, ph := range im.touched {
		if err := pr.plan.Save(ph); err != nil {
			return err
		}
	}
	if err := im.writeConfig(); err != nil {
		return err
	}
	rep := im.report()
	if *asJSON {
		return writeJSON(e, rep)
	}
	rep.print(e)
	return nil
}

// legacyImport holds the state of one legacy import.
type legacyImport struct {
	pr  *project
	lc  *importer.LegacyConfig
	lp  *importer.LegacyProject
	app string // --app

	notes   []string
	touched []*plan.Phase
	created []*plan.Ticket // every ticket added, in ID order
	rows    []*plan.Ticket // by index into lp.Tickets
	pkgs    []legacyPkg    // one per package, in file order

	dropped    []legacyDrop
	multi      []legacyMulti // tickets named by more than one package
	milestones []config.Milestone
	order      []string // phase slugs for [phases] order
	inbox      int
	archive    int
}

// legacyPkg is what a work package became.
type legacyPkg struct {
	phase   *plan.Phase
	members []*plan.Ticket // its tickets; the one placeholder for a package without scope
}

type legacyDrop struct {
	Ticket       string `json:"ticket"`
	Was          string `json:"was"`
	TrackedAs    string `json:"tracked_as"`
	TrackedAsWas string `json:"tracked_as_was"`
	Package      string `json:"package"`
}

type legacyMulti struct {
	Old    string   `json:"old_id"`
	Ticket string   `json:"ticket"`
	Kept   string   `json:"kept_in"`
	Others []string `json:"also_named_by"`
}

func (im *legacyImport) phaseFor(slug, title, goal string) *plan.Phase {
	if im.pr.plan.Phase(slug) == nil {
		im.pr.ensurePhase(slug, goal).Title = title
	}
	ph := im.pr.plan.Phase(slug)
	if !slices.Contains(im.touched, ph) {
		im.touched = append(im.touched, ph)
	}
	return ph
}

func statusOf(s importer.LegacyStatus, old string) plan.Status {
	st := plan.Status{State: plan.Pending, Aliases: []string{old}}
	switch s {
	case importer.LegacyDone:
		st.State, st.Before = plan.Done, true
	case importer.LegacyPartial:
		st.State = plan.InProgress
	}
	return st
}

// newTicket adds a ticket to a phase and remembers it for the report.
func (im *legacyImport) newTicket(ph *plan.Phase, app, title string, st plan.Status) *plan.Ticket {
	t := im.pr.newTicket(ph, app, title, "-", nil, st)
	im.created = append(im.created, t)
	return t
}

// addRow adds the ticket for a row of a source WBS.
func (im *legacyImport) addRow(ph *plan.Phase, row int) *plan.Ticket {
	lt := im.lp.Tickets[row]
	title := lt.Title
	if title == "" {
		title = lt.ID
		im.notes = append(im.notes, fmt.Sprintf("%s: no title, so the old ID is the title", lt.ID))
	}
	app := im.app
	if a := im.lc.Sources[lt.Source].App; a != "" {
		app = a
	}
	t := im.newTicket(ph, app, title, statusOf(lt.Status, lt.ID))
	im.rows[row] = t
	return t
}

// byArea is the import without a rollout WBS: one phase per source.
func (im *legacyImport) byArea() {
	for si, s := range im.lc.Sources {
		ph := im.phaseFor(s.Phase, s.Area, fmt.Sprintf("Every ticket imported from %s is done", s.Path))
		for i, t := range im.lp.Tickets {
			if t.Source == si {
				im.addRow(ph, i)
			}
		}
	}
}

// packageSlug is the RO number and the first words of the name, at most
// maxSlugLen characters and not in used.
func packageSlug(id, name string, used map[string]bool) string {
	base := importer.LegacySlug(id + " " + name)
	if len(base) > maxSlugLen {
		cut := base[:maxSlugLen]
		if base[maxSlugLen] != '-' {
			if i := strings.LastIndex(cut, "-"); i > 0 {
				cut = cut[:i]
			}
		}
		base = strings.Trim(cut, "-")
	}
	slug := base
	for n := 2; used[slug]; n++ {
		suf := fmt.Sprintf("-%d", n)
		b := base
		if len(b)+len(suf) > maxSlugLen {
			b = strings.TrimRight(b[:maxSlugLen-len(suf)], "-")
		}
		slug = b + suf
	}
	used[slug] = true
	return slug
}

// byPackage is the import with a rollout WBS: one phase per package.
func (im *legacyImport) byPackage() {
	lp := im.lp
	// Which package gets each row: the first in file order to name it.
	owner := make([]int, len(lp.Tickets))
	crossRef := make([]bool, len(lp.Tickets)) // owned as a cross-reference
	for i := range owner {
		owner[i] = -1
	}
	named := map[int][]int{} // row -> packages that named it, in order
	for pi, p := range lp.Packages {
		mark := func(r int, cross bool) {
			named[r] = append(named[r], pi)
			if owner[r] < 0 {
				owner[r], crossRef[r] = pi, cross
			}
		}
		for _, r := range p.Counted {
			mark(r, false)
		}
		for _, r := range p.Aliased {
			mark(r, true)
		}
	}

	used := map[string]bool{inboxPhase: true, archivePhase: true}
	for _, ph := range im.pr.plan.Phases {
		used[ph.Slug] = true
	}
	im.pkgs = make([]legacyPkg, len(lp.Packages))
	for pi, p := range lp.Packages {
		name := p.Name
		if name == "" {
			name = p.ID
		}
		slug := packageSlug(p.ID, name, used)
		ph := im.phaseFor(slug, name, fmt.Sprintf("%s (was %s)", name, p.ID))
		im.order = append(im.order, slug)
		pk := &im.pkgs[pi]
		pk.phase = ph
		for i := range lp.Tickets {
			if owner[i] == pi {
				pk.members = append(pk.members, im.addRow(ph, i))
			}
		}
		if len(pk.members) == 0 {
			if p.Scope != "" {
				im.notes = append(im.notes, fmt.Sprintf("%s: every ticket in its scope belongs to an earlier package, so the phase has one ticket for the package itself", p.ID))
			}
			st := statusOf(p.Status, p.ID)
			if st.State == plan.Pending {
				st.Human = fmt.Sprintf("rollout task carried from %s; check who does it", p.ID)
			}
			pk.members = []*plan.Ticket{im.newTicket(ph, im.app, name, st)}
		}
	}

	// Cross-references become dropped tickets that point at the ticket they
	// duplicate.
	for pi, p := range lp.Packages {
		for _, r := range p.Aliased {
			if owner[r] != pi || !crossRef[r] {
				continue
			}
			t := lp.Tickets[r]
			target := -1
			for _, c := range p.Counted {
				if lp.Tickets[c].ID == t.AliasOf && im.rows[c] != nil {
					target = c
					break
				}
			}
			if target < 0 {
				for c, o := range lp.Tickets {
					if o.ID == t.AliasOf && im.rows[c] != nil {
						target = c
						break
					}
				}
			}
			if target < 0 {
				continue
			}
			im.rows[r].Status = plan.Status{
				State:   plan.Dropped,
				Reason:  fmt.Sprintf("tracked as %s (was %s)", im.rows[target].ID, t.ID),
				By:      importedBy,
				Aliases: []string{t.ID},
			}
			gone := im.rows[r]
			im.pkgs[pi].members = slices.DeleteFunc(im.pkgs[pi].members, func(m *plan.Ticket) bool { return m == gone })
			im.dropped = append(im.dropped, legacyDrop{
				Ticket: gone.ID, Was: t.ID, TrackedAs: im.rows[target].ID, TrackedAsWas: lp.Tickets[target].ID, Package: p.ID,
			})
		}
	}

	// A row in more than one package.
	for r, pis := range named {
		m := legacyMulti{Old: lp.Tickets[r].ID, Ticket: im.rows[r].ID, Kept: lp.Packages[pis[0]].ID}
		for _, pi := range pis[1:] {
			if id := lp.Packages[pi].ID; id != m.Kept && !slices.Contains(m.Others, id) {
				m.Others = append(m.Others, id)
			}
		}
		if len(m.Others) > 0 {
			im.multi = append(im.multi, m)
		}
	}
	slices.SortFunc(im.multi, func(a, b legacyMulti) int { return strings.Compare(a.Ticket, b.Ticket) })

	// Rows in no package: open ones to the inbox, done ones to the archive.
	for pass := 0; pass < 2; pass++ {
		for i, t := range lp.Tickets {
			if owner[i] >= 0 || (t.Status == importer.LegacyDone) != (pass == 1) {
				continue
			}
			if pass == 0 {
				im.addRow(im.phaseFor(inboxPhase, plan.TitleFromSlug(inboxPhase), inboxGoal), i)
				im.inbox++
			} else {
				im.addRow(im.phaseFor(archivePhase, plan.TitleFromSlug(archivePhase), archiveGoal), i)
				im.archive++
			}
		}
	}
	if im.inbox > 0 {
		im.order = append(im.order, inboxPhase)
	}
	if im.archive > 0 {
		im.order = append(im.order, archivePhase)
	}

	im.linkPackages()
	im.headingsToMilestones()
}

// linkPackages makes the roots of a package depend on the sinks of each
// package it depends on.
func (im *legacyImport) linkPackages() {
	byID := map[string]*plan.Ticket{}
	for _, t := range im.pr.plan.Tickets() {
		byID[t.ID] = t
	}
	link := func(from, to *plan.Ticket) bool {
		if from == to || reaches(byID, to.ID, from.ID) {
			return false
		}
		if !slices.Contains(from.Depends, to.ID) {
			from.Depends = append(from.Depends, to.ID)
			plan.SortIDs(from.Depends)
		}
		return true
	}
	// Roots and sinks are worked out before any link is added.
	roots := make([][]*plan.Ticket, len(im.pkgs))
	sinks := make([][]*plan.Ticket, len(im.pkgs))
	for pi, pk := range im.pkgs {
		for _, t := range pk.members {
			isRoot, isSink := true, true
			for _, o := range pk.members {
				if slices.Contains(t.Depends, o.ID) {
					isRoot = false
				}
				if slices.Contains(o.Depends, t.ID) {
					isSink = false
				}
			}
			if isRoot {
				roots[pi] = append(roots[pi], t)
			}
			if isSink {
				sinks[pi] = append(sinks[pi], t)
			}
		}
	}
	byRO := map[string][]int{}
	for pi, p := range im.lp.Packages {
		byRO[p.ID] = append(byRO[p.ID], pi)
	}
	for pi, p := range im.lp.Packages {
		for _, d := range p.Depends {
			if d == p.ID {
				continue // reported by the parser
			}
			for _, dj := range byRO[d] {
				ok := true
				for _, r := range roots[pi] {
					for _, s := range sinks[dj] {
						if !link(r, s) {
							ok = false
						}
					}
				}
				if !ok {
					im.notes = append(im.notes, fmt.Sprintf("%s: depends on %s, which would make a dependency loop; not linked", p.ID, d))
				}
			}
		}
	}
}

// headingsToMilestones makes one milestone per "##" heading that has packages.
func (im *legacyImport) headingsToMilestones() {
	taken := map[string]bool{}
	for _, m := range im.pr.cfg.MilestoneList() {
		taken[m.Name] = true
	}
	index := map[string]int{} // heading -> position in im.milestones
	for pi, p := range im.lp.Packages {
		if p.Heading == "" {
			continue
		}
		i, ok := index[p.Heading]
		if !ok {
			name := p.Heading
			if c := strings.Index(name, ":"); c > 0 {
				name = name[:c]
			}
			base := importer.LegacySlug(name)
			if base == "" {
				base = fmt.Sprintf("milestone-%d", len(im.milestones)+1)
			}
			slug := base
			for n := 2; taken[slug]; n++ {
				slug = fmt.Sprintf("%s-%d", base, n)
			}
			taken[slug] = true
			i = len(im.milestones)
			index[p.Heading] = i
			im.milestones = append(im.milestones, config.Milestone{Name: slug})
		}
		im.milestones[i].Targets = append(im.milestones[i].Targets, im.pkgs[pi].phase.Slug)
	}
}

// writeConfig adds the milestones and the phase order to ajiya.toml. The
// phases are saved first, so the config can name them.
func (im *legacyImport) writeConfig() error {
	pr := im.pr
	for _, m := range im.milestones {
		if err := editConfig(pr, func(text string) (string, error) {
			return config.AddMilestoneText(text, m, "", "")
		}); err != nil {
			return err
		}
	}
	if len(im.order) > 0 {
		order := slices.Clone(pr.cfg.Phases.Order)
		for _, s := range im.order {
			if !slices.Contains(order, s) {
				order = append(order, s)
			}
		}
		if _, err := config.SetPhaseOrder(pr.cfg.Root, order); err != nil {
			return err
		}
	}
	return nil
}

// reaches reports whether to can be reached from from by following dependencies.
func reaches(byID map[string]*plan.Ticket, from, to string) bool {
	seen := map[string]bool{}
	stack := []string{from}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if cur == to {
			return true
		}
		if seen[cur] {
			continue
		}
		seen[cur] = true
		if t := byID[cur]; t != nil {
			stack = append(stack, t.Depends...)
		}
	}
	return false
}

// legacyReport is what the import prints, as text or as --json.
type legacyReport struct {
	Tickets    int                 `json:"tickets"`
	Sources    int                 `json:"sources"`
	Packages   []legacyReportPkg   `json:"packages"`
	Phases     []legacyReportPhase `json:"phases"`
	Milestones []legacyReportMile  `json:"milestones"`
	Dropped    []legacyDrop        `json:"dropped"`
	MultiPkg   []legacyMulti       `json:"named_by_several_packages"`
	Duplicates []legacyReportDup   `json:"duplicate_old_ids"`
	Inbox      int                 `json:"inbox"`
	Archive    int                 `json:"archive"`
	Notes      []string            `json:"notes"`

	total legacyReportPhase
}

type legacyReportPkg struct {
	ID    string `json:"id"`
	Phase string `json:"phase"`
}

type legacyReportPhase struct {
	Slug       string `json:"slug"`
	Tickets    int    `json:"tickets"`
	Done       int    `json:"done"`
	InProgress int    `json:"in_progress"`
	Pending    int    `json:"pending"`
	Dropped    int    `json:"dropped"`
}

type legacyReportMile struct {
	Name    string   `json:"name"`
	Targets []string `json:"targets"`
}

type legacyReportDup struct {
	Old     string   `json:"old_id"`
	Tickets []string `json:"tickets"`
	Phases  []string `json:"phases"`
}

func (p *legacyReportPhase) count(t *plan.Ticket) {
	p.Tickets++
	switch t.Status.State {
	case plan.Done:
		p.Done++
	case plan.InProgress:
		p.InProgress++
	case plan.Pending:
		p.Pending++
	case plan.Dropped:
		p.Dropped++
	}
}

func (p legacyReportPhase) String() string {
	s := fmt.Sprintf("%d ticket(s): %d done, %d in progress, %d pending", p.Tickets, p.Done, p.InProgress, p.Pending)
	if p.Dropped > 0 {
		s += fmt.Sprintf(", %d dropped", p.Dropped)
	}
	return s
}

func (im *legacyImport) report() *legacyReport {
	r := &legacyReport{
		Tickets: len(im.created), Sources: len(im.lc.Sources), Dropped: im.dropped, MultiPkg: im.multi,
		Inbox: im.inbox, Archive: im.archive,
		Packages: []legacyReportPkg{}, Phases: []legacyReportPhase{}, Milestones: []legacyReportMile{},
		Duplicates: []legacyReportDup{},
	}
	if r.Dropped == nil {
		r.Dropped = []legacyDrop{}
	}
	if r.MultiPkg == nil {
		r.MultiPkg = []legacyMulti{}
	}
	for pi, p := range im.lp.Packages {
		if im.pkgs != nil {
			r.Packages = append(r.Packages, legacyReportPkg{ID: p.ID, Phase: im.pkgs[pi].phase.Slug})
		}
	}
	for _, m := range im.milestones {
		r.Milestones = append(r.Milestones, legacyReportMile{Name: m.Name, Targets: m.Targets})
	}
	for _, ph := range im.touched {
		rp := legacyReportPhase{Slug: ph.Slug}
		for _, t := range im.created {
			if t.Phase == ph.Slug {
				rp.count(t)
			}
		}
		r.Phases = append(r.Phases, rp)
	}
	for _, t := range im.created {
		r.total.count(t)
	}

	// Old IDs on more than one ticket, in the order first seen.
	var order []string
	uses := map[string][]*plan.Ticket{}
	for i, t := range im.rows {
		if t == nil {
			continue
		}
		old := im.lp.Tickets[i].ID
		if _, ok := uses[old]; !ok {
			order = append(order, old)
		}
		uses[old] = append(uses[old], t)
	}
	for _, old := range order {
		if len(uses[old]) < 2 {
			continue
		}
		d := legacyReportDup{Old: old}
		for _, t := range uses[old] {
			d.Tickets = append(d.Tickets, t.ID)
			d.Phases = append(d.Phases, t.Phase)
		}
		r.Duplicates = append(r.Duplicates, d)
	}

	// An old ID shaped like a new ID would be shadowed by the real ticket.
	r.Notes = slices.Clone(im.notes)
	for _, old := range order {
		if plan.ValidID(im.pr.cfg.Project.Prefix, old) {
			r.Notes = append(r.Notes, fmt.Sprintf("%s: looks like a new ID, so 'ajiya ticket show %s' finds that ticket, not the imported one", old, old))
		}
	}
	if r.Notes == nil {
		r.Notes = []string{}
	}
	return r
}

func (r *legacyReport) print(e *env) {
	w := e.stdout
	fmt.Fprintf(w, "Imported %d ticket(s) from %d WBS file(s) and %d work package(s).\n", r.Tickets, r.Sources, len(r.Packages))
	for _, p := range r.Phases {
		fmt.Fprintf(w, "  %s: %s\n", p.Slug, p)
	}
	fmt.Fprintf(w, "  total: %s\n", r.total)

	section := func(head string, lines []string) {
		if len(lines) == 0 {
			return
		}
		fmt.Fprintln(w, head)
		for _, l := range lines {
			fmt.Fprintf(w, "  %s\n", l)
		}
	}
	var lines []string
	for _, p := range r.Packages {
		lines = append(lines, fmt.Sprintf("%s: %s", p.ID, p.Phase))
	}
	section("Work packages, now phases:", lines)

	lines = nil
	for _, m := range r.Milestones {
		lines = append(lines, fmt.Sprintf("%s: %s", m.Name, strings.Join(m.Targets, ", ")))
	}
	section("Milestones, from the headings:", lines)

	lines = nil
	for _, m := range r.MultiPkg {
		lines = append(lines, fmt.Sprintf("%s (%s): kept in %s, also named by %s", m.Old, m.Ticket, m.Kept, strings.Join(m.Others, ", ")))
	}
	section("Tickets named by more than one work package, each kept in the first:", lines)

	lines = nil
	for _, d := range r.Duplicates {
		var where []string
		for i, id := range d.Tickets {
			where = append(where, fmt.Sprintf("%s (%s)", id, d.Phases[i]))
		}
		lines = append(lines, fmt.Sprintf("%s: %s", d.Old, strings.Join(where, ", ")))
	}
	section("Old IDs used by more than one ticket, each now its own ticket ('ajiya ticket show <old ID>' finds the first):", lines)

	lines = nil
	for _, d := range r.Dropped {
		lines = append(lines, fmt.Sprintf("%s: %s (%s) dropped, tracked as %s (%s); reopen it if it is separate work", d.Package, d.Was, d.Ticket, d.TrackedAsWas, d.TrackedAs))
	}
	section("Cross-references, dropped so the work counts once:", lines)

	if r.Inbox > 0 || r.Archive > 0 {
		fmt.Fprintf(w, "In no work package: %d open ticket(s) in inbox, %d done in archive.\n", r.Inbox, r.Archive)
	}
	section("Not mapped, or read with a guess:", r.Notes)
}
