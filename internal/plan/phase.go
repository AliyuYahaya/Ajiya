package plan

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Header is the first line of every phase file.
const Header = "<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->"

var columns = []string{"ID", "App", "Ticket", "Done when", "Depends", "Tags", "Status"}

// oldColumns is the form before the Tags column. It is still read; the writer
// always upgrades it.
var oldColumns = []string{"ID", "App", "Ticket", "Done when", "Depends", "Status"}

// SlugRE matches a phase slug, which is also its file name.
var SlugRE = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Phase is one ajiya/<slug>.md file.
type Phase struct {
	Slug  string
	Title string
	Goal  string
	// Description is optional prose under the goal: one or more paragraphs.
	Description string
	Tickets     []*Ticket

	// SixColumn is set when the file was read in the old form, without Tags.
	SixColumn bool
}

// Ticket is one row of a phase table.
type Ticket struct {
	ID       string
	App      string
	Title    string
	DoneWhen string
	Depends  []string
	Tags     Tags
	Status   Status

	Phase string // slug of the phase holding the ticket
	Line  int    // 1-based line in the phase file, 0 if not read from a file
}

// TitleFromSlug is the default title for a phase: the slug with its first
// letter in capitals.
func TitleFromSlug(slug string) string {
	if slug == "" {
		return ""
	}
	return strings.ToUpper(slug[:1]) + slug[1:]
}

// ParseError is a phase file that cannot be read, with the line at fault.
type ParseError struct {
	Line int
	Msg  string
}

func (e *ParseError) Error() string { return fmt.Sprintf("line %d: %s", e.Line, e.Msg) }

// ParsePhase reads a phase file. It is strict about structure but not about
// spacing; Format shows the canonical form.
func ParsePhase(slug string, data []byte) (*Phase, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	p := &Phase{Slug: slug}
	fail := func(i int, format string, a ...any) (*Phase, error) {
		i = min(i, len(lines)-1) // at the end of the file, point at the last line
		return nil, &ParseError{Line: i + 1, Msg: fmt.Sprintf(format, a...)}
	}

	i := 0
	next := func() (string, bool) { // next non-blank line
		for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}
		if i >= len(lines) {
			return "", false
		}
		return lines[i], true
	}

	l, ok := next()
	if ok && strings.TrimSpace(l) == Header {
		i++
		l, ok = next()
	}
	if !ok || !strings.HasPrefix(l, "# ") {
		return fail(i, `expected the phase title as "# <title>"`)
	}
	p.Title = strings.TrimSpace(strings.TrimPrefix(l, "# "))
	i++

	l, ok = next()
	if !ok || !strings.HasPrefix(l, "Goal:") {
		return fail(i, `expected "Goal: <goal>" after the title`)
	}
	p.Goal = strings.TrimSpace(strings.TrimPrefix(l, "Goal:"))
	i++

	// Anything between the goal and the table is the description.
	start := i
	for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "|") {
		i++
	}
	if i >= len(lines) {
		return fail(i, "expected the ticket table after the goal")
	}
	desc, err := NormalizeDescription(strings.Join(lines[start:i], "\n"))
	if err != nil {
		return fail(start, "%v", err)
	}
	p.Description = desc
	l = lines[i]
	cols := columns
	if cells, ok := splitRow(l); ok && strings.Join(cells, "|") == strings.Join(oldColumns, "|") {
		cols, p.SixColumn = oldColumns, true
	} else if !ok || strings.Join(cells, "|") != strings.Join(columns, "|") {
		return fail(i, "expected the table header | %s |", strings.Join(columns, " | "))
	}
	i++
	if i >= len(lines) {
		return fail(i, "expected the table separator |---|---|...")
	}
	if cells, ok := splitRow(lines[i]); !ok || len(cells) != len(cols) || !isSeparator(cells) {
		return fail(i, "expected the table separator |---|---|... right after the header")
	}
	i++

	for ; i < len(lines); i++ {
		l := lines[i]
		if strings.TrimSpace(l) == "" {
			continue
		}
		cells, ok := splitRow(l)
		if !ok {
			return fail(i, "unexpected text after the ticket table")
		}
		if len(cells) != len(cols) {
			return fail(i, "row has %d cells, needs %d (%s)", len(cells), len(cols), strings.Join(cols, ", "))
		}
		t := &Ticket{
			ID:       cells[0],
			App:      cells[1],
			Title:    unescape(cells[2]),
			DoneWhen: unescape(cells[3]),
			Depends:  ParseIDList(cells[4]),
			Phase:    slug,
			Line:     i + 1,
		}
		if t.ID == "" {
			return fail(i, "row has no ID")
		}
		if !p.SixColumn {
			var err error
			if t.Tags, err = ParseTags(cells[5]); err != nil {
				return fail(i, "%s: %v", t.ID, err)
			}
		}
		st, err := ParseStatus(unescape(cells[len(cells)-1]))
		if err != nil {
			return fail(i, "%s: %v", t.ID, err)
		}
		t.Status = st
		p.Tickets = append(p.Tickets, t)
	}
	return p, nil
}

// Format writes the phase in canonical form: rows sorted by ID, always with
// the Tags column, so writing a phase read in the old form upgrades it.
func (p *Phase) Format() []byte { return p.format(true) }

// FormatSixColumn writes the phase in the old form, without Tags. Only check
// uses it, to tell an old file from a hand-edited one.
func (p *Phase) FormatSixColumn() []byte { return p.format(false) }

func (p *Phase) format(withTags bool) []byte {
	cols := oldColumns
	if withTags {
		cols = columns
	}
	var b strings.Builder
	b.WriteString(Header + "\n")
	b.WriteString("# " + p.Title + "\n\n")
	b.WriteString("Goal: " + p.Goal + "\n\n")
	if p.Description != "" {
		b.WriteString(p.Description + "\n\n")
	}
	b.WriteString("| " + strings.Join(cols, " | ") + " |\n")
	b.WriteString(strings.Repeat("|---", len(cols)) + "|\n")
	tickets := append([]*Ticket(nil), p.Tickets...)
	sort.SliceStable(tickets, func(i, j int) bool { return LessID(tickets[i].ID, tickets[j].ID) })
	for _, t := range tickets {
		deps := "-"
		if len(t.Depends) > 0 {
			ids := append([]string(nil), t.Depends...)
			SortIDs(ids)
			deps = strings.Join(ids, ", ")
		}
		cells := []string{t.ID, t.App, escape(t.Title), escape(t.DoneWhen), deps}
		if withTags {
			cells = append(cells, t.Tags.String())
		}
		cells = append(cells, escape(t.Status.String()))
		b.WriteString("| " + strings.Join(cells, " | ") + " |\n")
	}
	return []byte(b.String())
}

// splitRow splits a markdown table row on unescaped pipes and trims each cell.
func splitRow(line string) ([]string, bool) {
	line = strings.TrimSpace(line)
	if len(line) < 2 || line[0] != '|' || line[len(line)-1] != '|' || strings.HasSuffix(line, `\|`) {
		return nil, false
	}
	line = line[1 : len(line)-1]
	var cells []string
	start := 0
	for j := 0; j < len(line); j++ {
		if line[j] == '|' && (j == 0 || line[j-1] != '\\') {
			cells = append(cells, strings.TrimSpace(line[start:j]))
			start = j + 1
		}
	}
	return append(cells, strings.TrimSpace(line[start:])), true
}

func isSeparator(cells []string) bool {
	for _, c := range cells {
		if strings.Trim(c, ":-") != "" || !strings.Contains(c, "-") {
			return false
		}
	}
	return true
}

func escape(s string) string   { return strings.ReplaceAll(s, "|", `\|`) }
func unescape(s string) string { return strings.ReplaceAll(s, `\|`, "|") }

// NormalizeDescription checks a phase description and returns it in canonical
// form: lines without trailing spaces, at most one blank line in a row, none at
// either end. It may be several paragraphs. A line must not start with "|" (it
// would read as the table) or "Goal:". An empty result means no description.
func NormalizeDescription(s string) (string, error) {
	if strings.Contains(s, "\r") {
		return "", fmt.Errorf("description must not contain carriage returns")
	}
	var out []string
	for _, l := range strings.Split(strings.TrimSpace(s), "\n") {
		l = strings.TrimRight(l, " \t")
		t := strings.TrimLeft(l, " \t")
		if strings.HasPrefix(t, "|") {
			return "", fmt.Errorf("description lines must not start with \"|\"")
		}
		if strings.HasPrefix(t, "Goal:") {
			return "", fmt.Errorf("description lines must not start with \"Goal:\"")
		}
		if l == "" && (len(out) == 0 || out[len(out)-1] == "") {
			continue
		}
		out = append(out, l)
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return strings.Join(out, "\n"), nil
}
