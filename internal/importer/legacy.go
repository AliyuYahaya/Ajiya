// Package importer reads work tracked elsewhere so the CLI can turn it into
// Ajiya tickets.
//
// The legacy importer reads the older format that reference/collate.py
// collates: several WBS files whose table rows start with a ticket ID and
// carry a status mark, and a rollout WBS of RO- work packages with a scope and
// dependencies. It follows collate.py's parsing rules:
//
//   - A row belongs to a source when its first cell is <PREFIX>-<digits> for
//     one of the configured prefixes. Rows with another prefix (other than
//     RO-) are skipped and reported.
//   - The status is the last cell containing 🟩 (done), 🟨 (partial) or 🟥
//     (pending), checked in that order within a cell. A row with none is
//     unknown.
//   - "## " and "### " headings set the epic of the rows below them.
//   - A title or status saying "tracked as X", "cross-reference only: X",
//     "duplicate of X" or "same as X" makes the row a cross-reference to X.
//   - Rollout rows are | RO-n | name | scope | depends | launch | status |.
//     Scope is comma separated: API-013, API-037..API-044 (inclusive, with
//     three-digit numbers), API[EPIC G] (every API ticket under a heading
//     starting "EPIC G", case-insensitive). "", "-" and "none" mean no scope.
//     A cross-reference whose target is also in scope counts once.
//   - A package with a scope takes its status from its tickets; one without
//     keeps its typed status: a mark, or Done, In progress, Not started and
//     similar words.
//
// Unlike collate.py, which strips markdown from titles, titles are kept
// exactly as written (only "\|" becomes "|"), because imports never reword.
// And two rows with the same ID in the same file stay two tickets.
package importer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// LegacyStatus is a row's status as collate.py names it.
type LegacyStatus string

const (
	LegacyDone    LegacyStatus = "done"
	LegacyPartial LegacyStatus = "partial"
	LegacyPending LegacyStatus = "pending"
	LegacyUnknown LegacyStatus = "unknown"
)

// LegacyConfig describes a legacy project, like the constants at the top of
// collate.py. Paths are relative to the config file.
//
//	prefixes = ["API", "WEB", "OPS"]
//	rollout = "rollout/rollout-wbs.md"   # optional
//	rollout_phase = "rollout"            # optional, the default
//
//	[[sources]]
//	area = "API"
//	path = "docs/api/wbs.md"
//	phase = "api"                        # optional: the area as a slug
type LegacyConfig struct {
	Prefixes     []string       `toml:"prefixes"`
	Rollout      string         `toml:"rollout"`
	RolloutPhase string         `toml:"rollout_phase"`
	Sources      []LegacySource `toml:"sources"`

	Dir string `toml:"-"` // the directory holding the config file
}

// LegacySource is one WBS file.
type LegacySource struct {
	Area  string `toml:"area"`
	Path  string `toml:"path"`
	Phase string `toml:"phase"`
}

var (
	legacyPrefixRE = regexp.MustCompile(`^[A-Z][A-Z0-9]*$`)
	legacySlugRE   = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	nonSlugRE      = regexp.MustCompile(`[^a-z0-9]+`)
)

// LegacySlug turns an area name into a phase slug: "Web app" is web-app.
func LegacySlug(name string) string {
	return strings.Trim(nonSlugRE.ReplaceAllString(strings.ToLower(name), "-"), "-")
}

// LoadLegacyConfig reads and validates a legacy config file.
func LoadLegacyConfig(path string) (*LegacyConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c, err := ParseLegacyConfig(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", path, err)
	}
	c.Dir = filepath.Dir(path)
	return c, nil
}

// ParseLegacyConfig decodes a legacy config and fills in the defaults.
func ParseLegacyConfig(data []byte) (*LegacyConfig, error) {
	var c LegacyConfig
	md, err := toml.Decode(string(data), &c)
	if err != nil {
		return nil, errors.New(strings.TrimPrefix(err.Error(), "toml: "))
	}
	if und := md.Undecoded(); len(und) > 0 {
		return nil, fmt.Errorf("unknown key %q; check the spelling", und[0].String())
	}
	if len(c.Prefixes) == 0 {
		return nil, errors.New(`prefixes is empty; list the ticket ID prefixes, like prefixes = ["API", "WEB"]`)
	}
	for _, p := range c.Prefixes {
		if !legacyPrefixRE.MatchString(p) || p == "RO" {
			return nil, fmt.Errorf("prefix %q must be capital letters and digits, and not RO (used by work packages)", p)
		}
	}
	if len(c.Sources) == 0 {
		return nil, errors.New("no [[sources]]; add one per WBS file with area and path")
	}
	if c.RolloutPhase == "" {
		c.RolloutPhase = "rollout"
	}
	phases := map[string]bool{}
	for i := range c.Sources {
		s := &c.Sources[i]
		if s.Area == "" || s.Path == "" {
			return nil, fmt.Errorf("source %d needs both area and path", i+1)
		}
		if strings.TrimSpace(s.Area) != s.Area || strings.ContainsAny(s.Area, "\r\n|") {
			return nil, fmt.Errorf("area %q must be one line with no '|' and no spaces around it", s.Area)
		}
		if s.Phase == "" {
			s.Phase = LegacySlug(s.Area)
		}
		if !legacySlugRE.MatchString(s.Phase) {
			return nil, fmt.Errorf("phase %q of source %s must be lower case letters and digits separated by hyphens; set phase", s.Phase, s.Area)
		}
		if phases[s.Phase] {
			return nil, fmt.Errorf("two sources use phase %q; set a different phase on one", s.Phase)
		}
		phases[s.Phase] = true
	}
	if c.Rollout != "" {
		if !legacySlugRE.MatchString(c.RolloutPhase) {
			return nil, fmt.Errorf("rollout_phase %q must be lower case letters and digits separated by hyphens", c.RolloutPhase)
		}
		if phases[c.RolloutPhase] {
			return nil, fmt.Errorf("rollout_phase %q is also a source phase; choose another", c.RolloutPhase)
		}
	}
	return &c, nil
}

// LegacyTicket is one row of a source WBS.
type LegacyTicket struct {
	ID      string
	Title   string
	Status  LegacyStatus
	Epic    string // the heading above the row
	Source  int    // index into LegacyConfig.Sources
	Line    int    // 1-based
	AliasOf string // the ticket this row is a cross-reference to, if any
}

// LegacyPackage is one RO- row of the rollout WBS.
type LegacyPackage struct {
	ID      string
	Name    string
	Scope   string // "" when the package has none
	Depends []string
	Launch  string
	Typed   string // the status cell as written
	Status  LegacyStatus
	Line    int // 1-based

	Counted   []int    // indexes into LegacyProject.Tickets that the scope counts
	Aliased   []int    // in scope, but cross-references to a counted ticket
	Missing   []string // scope entries that matched nothing
	Ambiguous []string // scope IDs that name more than one row
}

// LegacyProject is everything read from a legacy project.
type LegacyProject struct {
	Config   *LegacyConfig
	Tickets  []LegacyTicket
	Packages []LegacyPackage
	// Notes lists rows and references that were skipped or guessed, one line
	// each, as "<where>: <what>".
	Notes []string
}

// legacyRE holds the patterns that depend on the configured prefixes.
type legacyRE struct {
	id, anyID, group, rng, single, alias, ro, rorow, heading *regexp.Regexp
}

func newLegacyRE(prefixes []string) *legacyRE {
	q := make([]string, len(prefixes))
	for i, p := range prefixes {
		q[i] = regexp.QuoteMeta(p)
	}
	pfx := "(?:" + strings.Join(q, "|") + ")"
	return &legacyRE{
		id:      regexp.MustCompile(`^\|\s*(` + pfx + `-\d+)\s*\|`),
		anyID:   regexp.MustCompile(`^\|\s*([A-Z][A-Z0-9]*)-\d+\s*\|`),
		group:   regexp.MustCompile(`^(` + pfx + `)\[(.+)\]$`),
		rng:     regexp.MustCompile(`^(` + pfx + `-\d+)\.\.(` + pfx + `-\d+)$`),
		single:  regexp.MustCompile(`^` + pfx + `-\d+$`),
		alias:   regexp.MustCompile(`(?i)(?:tracked\s+as|cross-reference\s+only\W+|duplicate\s+of|same\s+as)\s+(` + pfx + `-\d+)`),
		ro:      regexp.MustCompile(`\bRO-\d+\b`),
		rorow:   regexp.MustCompile(`^\|\s*RO-\d+\s*\|`),
		heading: regexp.MustCompile(`^#{2,3}\s+(.*)`),
	}
}

// legacyCells splits a table row on unescaped pipes, like collate.py's cells.
func legacyCells(line string) []string {
	line = strings.TrimSpace(line)
	var parts []string
	start := 0
	for i := 0; i < len(line); i++ {
		if line[i] == '|' && (i == 0 || line[i-1] != '\\') {
			parts = append(parts, line[start:i])
			start = i + 1
		}
	}
	parts = append(parts, line[start:])
	if len(parts) < 2 {
		return nil
	}
	parts = parts[1 : len(parts)-1]
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

var legacyMarks = []struct {
	mark   string
	status LegacyStatus
}{{"🟩", LegacyDone}, {"🟨", LegacyPartial}, {"🟥", LegacyPending}}

// legacyStatusOf returns the status from the last cell holding a mark.
func legacyStatusOf(row []string) LegacyStatus {
	for i := len(row) - 1; i >= 0; i-- {
		for _, m := range legacyMarks {
			if strings.Contains(row[i], m.mark) {
				return m.status
			}
		}
	}
	return LegacyUnknown
}

var legacyWords = []struct {
	re     *regexp.Regexp
	status LegacyStatus
}{
	{regexp.MustCompile(`(?i)^\s*(done|complete|completed)\b`), LegacyDone},
	{regexp.MustCompile(`(?i)^\s*(in progress|partial|started)\b`), LegacyPartial},
	{regexp.MustCompile(`(?i)^\s*(not started|pending|to do|todo|blocked)\b`), LegacyPending},
}

// legacyTyped reads a typed package status, and whether it starts with a mark.
func legacyTyped(s string) (LegacyStatus, bool) {
	for _, m := range legacyMarks {
		if strings.HasPrefix(s, m.mark) {
			return m.status, true
		}
	}
	for _, w := range legacyWords {
		if w.re.MatchString(s) {
			return w.status, false
		}
	}
	return LegacyUnknown, false
}

// shortList joins ids, cut after n.
func shortList(ids []string, n int) string {
	if len(ids) <= n {
		return strings.Join(ids, ", ")
	}
	return strings.Join(ids[:n], ", ") + fmt.Sprintf(" and %d more", len(ids)-n)
}

// ParseLegacySource reads the ticket rows of one WBS file. where names the
// file in notes; src is its index in the config.
func ParseLegacySource(data, where string, src int, prefixes []string) ([]LegacyTicket, []string) {
	re := newLegacyRE(prefixes)
	var tickets []LegacyTicket
	var notes []string
	epic := ""
	unknown := map[string][]string{}
	var order []string
	var noMark []string
	for i, line := range strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n") {
		if h := re.heading.FindStringSubmatch(line); h != nil {
			epic = strings.TrimSpace(h[1])
			continue
		}
		m := re.id.FindStringSubmatch(line)
		if m == nil {
			if o := re.anyID.FindStringSubmatch(line); o != nil && o[1] != "RO" {
				if _, seen := unknown[o[1]]; !seen {
					order = append(order, o[1])
				}
				unknown[o[1]] = append(unknown[o[1]], legacyCells(line)[0])
			}
			continue
		}
		row := legacyCells(line)
		t := LegacyTicket{ID: m[1], Status: legacyStatusOf(row), Epic: epic, Source: src, Line: i + 1}
		if len(row) > 1 {
			t.Title = strings.ReplaceAll(row[1], `\|`, "|")
		}
		if a := re.alias.FindStringSubmatch(strings.Join(row[1:], " ")); a != nil && a[1] != t.ID {
			t.AliasOf = a[1]
		}
		if t.Status == LegacyUnknown {
			noMark = append(noMark, t.ID)
		}
		tickets = append(tickets, t)
	}
	if len(tickets) == 0 {
		notes = append(notes, where+": no ticket rows found; the ticket ID must be in the first column")
	}
	for _, p := range order {
		ids := unknown[p]
		notes = append(notes, fmt.Sprintf("%s: %d row(s) skipped because the prefix %s- is not in prefixes: %s", where, len(ids), p, shortList(ids, 12)))
	}
	if len(noMark) > 0 {
		notes = append(notes, fmt.Sprintf("%s: %d ticket(s) with no 🟩/🟨/🟥 mark, imported as pending: %s", where, len(noMark), shortList(noMark, 12)))
	}
	return tickets, notes
}

// ParseLegacyRollout reads the RO- rows of a rollout WBS and resolves their
// scopes against tickets.
func ParseLegacyRollout(data, where string, tickets []LegacyTicket, prefixes []string) ([]LegacyPackage, []string) {
	re := newLegacyRE(prefixes)
	var pkgs []LegacyPackage
	var notes []string
	for i, line := range strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n") {
		if !re.rorow.MatchString(line) {
			continue
		}
		row := legacyCells(line)
		if len(row) < 6 {
			notes = append(notes, fmt.Sprintf("%s:%d: %s has %d columns, needs 6 (ID, name, scope, depends, launch, status); skipped", where, i+1, row[0], len(row)))
			continue
		}
		p := LegacyPackage{ID: row[0], Name: strings.ReplaceAll(row[1], `\|`, "|"), Scope: row[2], Depends: re.ro.FindAllString(row[3], -1), Launch: row[4], Typed: row[5], Line: i + 1}
		if p.Scope == "-" || p.Scope == "none" {
			p.Scope = ""
		}
		if p.Scope != "" {
			p.Counted, p.Aliased, p.Missing, p.Ambiguous = resolveLegacyScope(re, p.Scope, tickets)
			p.Status = legacyDerived(p.Counted, tickets)
			if len(p.Missing) > 0 {
				notes = append(notes, fmt.Sprintf("%s: scope entries not found: %s", p.ID, strings.Join(p.Missing, ", ")))
			}
			if len(p.Ambiguous) > 0 {
				notes = append(notes, fmt.Sprintf("%s: scope names %s, used by more than one ticket, so every one of them is included", p.ID, shortList(p.Ambiguous, 12)))
			}
			if len(p.Counted) == 0 {
				notes = append(notes, fmt.Sprintf("%s: scope matched no tickets, imported as pending", p.ID))
			}
		} else {
			st, marked := legacyTyped(p.Typed)
			p.Status = st
			switch {
			case st == LegacyUnknown:
				notes = append(notes, fmt.Sprintf("%s: typed status %q is not recognised, imported as pending", p.ID, p.Typed))
			case !marked:
				notes = append(notes, fmt.Sprintf("%s: typed status %q has no mark, read as %s", p.ID, p.Typed, st))
			}
		}
		pkgs = append(pkgs, p)
	}
	known := map[string]bool{}
	for _, p := range pkgs {
		known[p.ID] = true
	}
	for _, p := range pkgs {
		for _, d := range p.Depends {
			switch {
			case d == p.ID:
				notes = append(notes, fmt.Sprintf("%s: depends on itself; not linked", p.ID))
			case !known[d]:
				notes = append(notes, fmt.Sprintf("%s: depends on %s, which does not exist; not linked", p.ID, d))
			}
		}
	}
	return pkgs, notes
}

// resolveLegacyScope follows collate.py's resolve_scope. Rows are told apart
// by position, so two rows with one ID are two tickets.
func resolveLegacyScope(re *legacyRE, scope string, tickets []LegacyTicket) (counted, aliased []int, missing, ambiguous []string) {
	byID := map[string][]int{}
	for i, t := range tickets {
		byID[t.ID] = append(byID[t.ID], i)
	}
	var found []int
	one := func(id string) {
		hits, ok := byID[id]
		if !ok {
			missing = append(missing, id)
			return
		}
		found = append(found, hits...)
		if len(hits) > 1 {
			ambiguous = append(ambiguous, id)
		}
	}
	for _, tok := range strings.Split(scope, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if g := re.group.FindStringSubmatch(tok); g != nil {
			var hits []int
			for i, t := range tickets {
				if strings.HasPrefix(t.ID, g[1]+"-") && strings.HasPrefix(strings.ToLower(t.Epic), strings.ToLower(g[2])) {
					hits = append(hits, i)
				}
			}
			if len(hits) == 0 {
				missing = append(missing, tok)
			}
			found = append(found, hits...)
			continue
		}
		if r := re.rng.FindStringSubmatch(tok); r != nil {
			p1, n1 := splitLegacyID(r[1])
			p2, n2 := splitLegacyID(r[2])
			if p1 != p2 || n2 < n1 {
				missing = append(missing, tok)
				continue
			}
			for n := n1; n <= n2; n++ {
				one(fmt.Sprintf("%s-%03d", p1, n))
			}
			continue
		}
		if re.single.MatchString(tok) {
			one(tok)
			continue
		}
		missing = append(missing, tok)
	}

	seen := map[int]bool{}
	var unique []int
	inScope := map[string]bool{}
	for _, i := range found {
		if !seen[i] {
			seen[i] = true
			unique = append(unique, i)
			inScope[tickets[i].ID] = true
		}
	}
	for _, i := range unique {
		if a := tickets[i].AliasOf; a != "" && inScope[a] {
			aliased = append(aliased, i)
		} else {
			counted = append(counted, i)
		}
	}
	return counted, aliased, missing, ambiguous
}

func splitLegacyID(id string) (string, int) {
	i := strings.LastIndex(id, "-")
	n, _ := strconv.Atoi(id[i+1:])
	return id[:i], n
}

// legacyDerived is a scoped package's status: done when every ticket is,
// partial when any is done or partial, pending otherwise.
func legacyDerived(idx []int, tickets []LegacyTicket) LegacyStatus {
	if len(idx) == 0 {
		return LegacyUnknown
	}
	done, partial := 0, 0
	for _, i := range idx {
		switch tickets[i].Status {
		case LegacyDone:
			done++
		case LegacyPartial:
			partial++
		}
	}
	switch {
	case done == len(idx):
		return LegacyDone
	case done > 0 || partial > 0:
		return LegacyPartial
	}
	return LegacyPending
}

// ReadLegacy reads every file a config names. A missing file is an error.
func ReadLegacy(c *LegacyConfig) (*LegacyProject, error) {
	lp := &LegacyProject{Config: c}
	read := func(rel string) (string, error) {
		data, err := os.ReadFile(filepath.Join(c.Dir, filepath.FromSlash(rel)))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return "", fmt.Errorf("%s does not exist; fix the path in the config (paths are relative to it)", rel)
			}
			return "", err
		}
		return string(data), nil
	}
	for i, s := range c.Sources {
		data, err := read(s.Path)
		if err != nil {
			return nil, err
		}
		ts, notes := ParseLegacySource(data, s.Path, i, c.Prefixes)
		lp.Tickets = append(lp.Tickets, ts...)
		lp.Notes = append(lp.Notes, notes...)
	}
	if c.Rollout != "" {
		data, err := read(c.Rollout)
		if err != nil {
			return nil, err
		}
		pkgs, notes := ParseLegacyRollout(data, c.Rollout, lp.Tickets, c.Prefixes)
		lp.Packages = pkgs
		lp.Notes = append(lp.Notes, notes...)
	}
	return lp, nil
}
