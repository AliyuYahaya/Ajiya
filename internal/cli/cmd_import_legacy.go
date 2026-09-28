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

// runImportLegacy imports a project tracked in the collate.py format (see
// package importer for the config file and the parsing rules).
//
// Mapping:
//   - Each source WBS becomes one phase (the slug of its area, or its phase
//     key, titled with the area name), not one phase per epic: epics in these files are often small or
//     missing, and agents reorganise the phases after the import.
//   - Each row becomes a ticket with a new ID, in file order, with the title
//     as written, Done when "-", and its old ID kept as an alias ("Was:").
//     🟩 is Done with "Done before Ajiya" as evidence, 🟨 In progress, 🟥 or
//     no mark Pending. Rows sharing an old ID each get their own ticket and
//     are listed in the report.
//   - With a rollout WBS, each RO- work package becomes a ticket in the
//     rollout phase, with its RO ID as an alias and its name as the title.
//     It depends on the tickets its scope counts (a cross-reference whose
//     target is also in scope is left out, so the work counts once) and on
//     the packages in its Depends column. Each ticket the scope counts also
//     depends on the packages its package depends on, so it is not ready
//     before them. A scoped package's status comes from its tickets, as in
//     collate.py; one without a scope keeps its typed status.
//   - A link that would make a dependency loop (a ticket in two packages
//     that depend on each other, or packages depending on each other) is left
//     out and reported, as are scope entries and dependencies that name
//     nothing.
//
// The import refuses to run when an old ticket ID it would add is already an
// alias in the plan: the project was imported before, and doing it again
// would duplicate every ticket. RO IDs are not compared (unless there are no
// ticket rows), so two legacy projects with their own prefixes can both be
// imported; 'ajiya ticket show RO-1' then finds the first.
func runImportLegacy(e *env, args []string) error {
	fs := newFlags("import legacy")
	app := fs.String("app", config.InfraApp, "app for the imported tickets")
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

	notes := lp.Notes
	var touched []*plan.Phase
	phaseFor := func(slug, title, goal string) *plan.Phase {
		if pr.plan.Phase(slug) == nil {
			pr.ensurePhase(slug, goal).Title = title
		}
		ph := pr.plan.Phase(slug)
		if !slices.Contains(touched, ph) {
			touched = append(touched, ph)
		}
		return ph
	}
	statusOf := func(s importer.LegacyStatus, old string) plan.Status {
		st := plan.Status{State: plan.Pending, Aliases: []string{old}}
		switch s {
		case importer.LegacyDone:
			st.State, st.Before = plan.Done, true
		case importer.LegacyPartial:
			st.State = plan.InProgress
		}
		return st
	}
	add := func(ph *plan.Phase, title, old string, s importer.LegacyStatus) *plan.Ticket {
		if title == "" {
			title = old
			notes = append(notes, fmt.Sprintf("%s: no title, so the old ID is the title", old))
		}
		return pr.newTicket(ph, *app, title, "-", nil, statusOf(s, old))
	}

	rows := make([]*plan.Ticket, len(lp.Tickets))
	for si, s := range lc.Sources {
		ph := phaseFor(s.Phase, s.Area, fmt.Sprintf("Every ticket imported from %s is done", s.Path))
		for i, t := range lp.Tickets {
			if t.Source == si {
				rows[i] = add(ph, t.Title, t.ID, t.Status)
			}
		}
	}

	pkgs := make([]*plan.Ticket, len(lp.Packages))
	byRO := map[string][]*plan.Ticket{}
	if lc.Rollout != "" {
		ph := phaseFor(lc.RolloutPhase, plan.TitleFromSlug(lc.RolloutPhase), fmt.Sprintf("Every work package in %s is done", lc.Rollout))
		for i, p := range lp.Packages {
			pkgs[i] = add(ph, p.Name, p.ID, p.Status)
			byRO[p.ID] = append(byRO[p.ID], pkgs[i])
		}
	}

	byID := map[string]*plan.Ticket{}
	for _, t := range pr.plan.Tickets() {
		byID[t.ID] = t
	}
	// link adds from -> to unless it is there already or makes a loop.
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
	for i, p := range lp.Packages {
		for _, r := range p.Counted {
			link(pkgs[i], rows[r])
		}
	}
	for i, p := range lp.Packages {
		for _, d := range p.Depends {
			if d == p.ID {
				continue // reported by the parser
			}
			for _, dt := range byRO[d] {
				if !link(pkgs[i], dt) {
					notes = append(notes, fmt.Sprintf("%s: depends on %s, which would make a dependency loop; not linked", p.ID, d))
				}
			}
		}
	}
	for _, p := range lp.Packages {
		for _, d := range p.Depends {
			if d == p.ID {
				continue
			}
			for _, dt := range byRO[d] {
				var skipped []string
				for _, r := range p.Counted {
					if !link(rows[r], dt) {
						skipped = append(skipped, lp.Tickets[r].ID)
					}
				}
				if len(skipped) > 0 {
					notes = append(notes, fmt.Sprintf("%s: %s not made to wait on %s, which already depends on it; not linked", p.ID, strings.Join(skipped, ", "), d))
				}
			}
		}
	}

	for _, ph := range touched {
		if err := pr.plan.Save(ph); err != nil {
			return err
		}
	}
	printLegacyReport(e, pr, lp, touched, rows, pkgs, notes)
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

func printLegacyReport(e *env, pr *project, lp *importer.LegacyProject, phases []*plan.Phase, rows, pkgs []*plan.Ticket, notes []string) {
	w := e.stdout
	all := append(append([]*plan.Ticket(nil), rows...), pkgs...)
	fmt.Fprintf(w, "Imported %d ticket(s) from %d WBS file(s) and %d work package(s).\n", len(all), len(lp.Config.Sources), len(pkgs))
	count := func(ts []*plan.Ticket) string {
		n := map[plan.State]int{}
		for _, t := range ts {
			n[t.Status.State]++
		}
		return fmt.Sprintf("%d done, %d in progress, %d pending", n[plan.Done], n[plan.InProgress], n[plan.Pending])
	}
	for _, ph := range phases {
		var ts []*plan.Ticket
		for _, t := range all {
			if t.Phase == ph.Slug {
				ts = append(ts, t)
			}
		}
		fmt.Fprintf(w, "  %s: %d ticket(s): %s\n", ph.Slug, len(ts), count(ts))
	}
	fmt.Fprintf(w, "  total: %d ticket(s): %s\n", len(all), count(all))

	// Old IDs on more than one ticket, in the order first seen.
	type use struct{ old, id, phase string }
	var order []string
	uses := map[string][]use{}
	for i, t := range rows {
		old := lp.Tickets[i].ID
		if _, ok := uses[old]; !ok {
			order = append(order, old)
		}
		uses[old] = append(uses[old], use{old, t.ID, t.Phase})
	}
	for i, t := range pkgs {
		old := lp.Packages[i].ID
		if _, ok := uses[old]; !ok {
			order = append(order, old)
		}
		uses[old] = append(uses[old], use{old, t.ID, t.Phase})
	}
	var dups []string
	for _, old := range order {
		if len(uses[old]) < 2 {
			continue
		}
		var where []string
		for _, u := range uses[old] {
			where = append(where, fmt.Sprintf("%s (%s)", u.id, u.phase))
		}
		dups = append(dups, fmt.Sprintf("%s: %s", old, strings.Join(where, ", ")))
	}
	section := func(head string, lines []string) {
		if len(lines) == 0 {
			return
		}
		fmt.Fprintln(w, head)
		for _, l := range lines {
			fmt.Fprintf(w, "  %s\n", l)
		}
	}
	section("Old IDs used by more than one ticket, each now its own ticket ('ajiya ticket show <old ID>' finds the first):", dups)

	var xrefs []string
	for _, p := range lp.Packages {
		for _, r := range p.Aliased {
			t := lp.Tickets[r]
			xrefs = append(xrefs, fmt.Sprintf("%s: %s (%s) is the same work as %s, so it counts once", p.ID, t.ID, rows[r].ID, t.AliasOf))
		}
	}
	section("Cross-references:", xrefs)

	// An old ID shaped like a new ID would be shadowed by the real ticket.
	var shadow []string
	for _, old := range order {
		if plan.ValidID(pr.cfg.Project.Prefix, old) {
			shadow = append(shadow, fmt.Sprintf("%s: looks like a new ID, so 'ajiya ticket show %s' finds that ticket, not the imported one", old, old))
		}
	}
	section("Not mapped, or read with a guess:", append(notes, shadow...))
}
