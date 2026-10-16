package plan

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// DecisionDir is the folder holding the decision files, relative to the project root.
const DecisionDir = Dir + "/decisions"

// DecPrefix is the prefix of decision IDs: DEC-0003.
const DecPrefix = "DEC"

// DecState is where a decision stands.
type DecState int

const (
	DecProposed DecState = iota
	DecAccepted
	DecSuperseded
)

func (s DecState) String() string {
	switch s {
	case DecAccepted:
		return "Accepted"
	case DecSuperseded:
		return "Superseded"
	}
	return "Proposed"
}

// DecStatus is the Status line of a decision file.
//
//	Proposed · 2026-09-12 · proposed by Aliyu
//	Accepted · 2026-09-12 · decided by Aliyu
//	Superseded by DEC-0007 (CHG-0014) · was: Accepted · 2026-09-12 · decided by Aliyu
//
// Date and By belong to the state the decision is in; once superseded they keep
// the values of the state it was in before (Was), so history is not lost.
// Superseded without a replacement is written "Superseded (CHG-0014)".
type DecStatus struct {
	State        DecState
	Date, By     string
	SupersededBy string   // DEC-xxxx of the replacement, or ""
	Change       string   // CHG-xxxx that superseded it
	Was          DecState // for Superseded: Proposed or Accepted
}

var supersededRE = regexp.MustCompile(`^Superseded(?: by (DEC-[0-9]{4,}))? \((CHG-[0-9]{4,})\)$`)

// verb is the phrase before the person's name for the state the date and name belong to.
func (s DecStatus) verb() string {
	st := s.State
	if st == DecSuperseded {
		st = s.Was
	}
	if st == DecAccepted {
		return "decided by "
	}
	return "proposed by "
}

func (s DecStatus) String() string {
	var parts []string
	switch s.State {
	case DecSuperseded:
		head := "Superseded"
		if s.SupersededBy != "" {
			head += " by " + s.SupersededBy
		}
		parts = append(parts, head+" ("+s.Change+")", "was: "+s.Was.String())
	default:
		parts = append(parts, s.State.String())
	}
	if s.Date != "" {
		parts = append(parts, s.Date)
	}
	if s.By != "" {
		parts = append(parts, s.verb()+s.By)
	}
	return strings.Join(parts, " · ")
}

// Open reports whether the decision still stands (it is not superseded).
func (s DecStatus) Open() bool { return s.State != DecSuperseded }

func parseDecStatus(text string) (DecStatus, error) {
	var s DecStatus
	parts := strings.Split(text, " · ")
	head := parts[0]
	rest := parts[1:]
	switch {
	case head == "Proposed":
		s.State = DecProposed
	case head == "Accepted":
		s.State = DecAccepted
	case strings.HasPrefix(head, "Superseded"):
		m := supersededRE.FindStringSubmatch(head)
		if m == nil {
			return s, fmt.Errorf("status %q: want 'Superseded by DEC-xxxx (CHG-xxxx)'", head)
		}
		s.State, s.SupersededBy, s.Change = DecSuperseded, m[1], m[2]
		if len(rest) == 0 || !strings.HasPrefix(rest[0], "was: ") {
			return s, errors.New("a superseded status needs '· was: Proposed' or '· was: Accepted'")
		}
		switch strings.TrimPrefix(rest[0], "was: ") {
		case "Proposed":
			s.Was = DecProposed
		case "Accepted":
			s.Was = DecAccepted
		default:
			return s, fmt.Errorf("status %q: was must be Proposed or Accepted", rest[0])
		}
		rest = rest[1:]
	default:
		return s, fmt.Errorf("status %q is not Proposed, Accepted or Superseded", head)
	}
	for _, r := range rest {
		switch {
		case dateRE.MatchString(r) && s.Date == "" && s.By == "":
			s.Date = r
		case s.By == "" && strings.HasPrefix(r, s.verb()) && len(r) > len(s.verb()):
			s.By = strings.TrimPrefix(r, s.verb())
		default:
			return s, fmt.Errorf("status part %q is not a date or '%sNAME'", r, s.verb())
		}
	}
	return s, nil
}

// Decision is one decision record, ajiya/decisions/DEC-xxxx.md.
type Decision struct {
	ID           string
	Title        string
	Status       DecStatus
	Tags         Tags // only tech tags are allowed in a decision
	Context      string
	Decision     string
	Consequences string
}

// DecisionPath is the path of a decision file, relative to the project root, with / separators.
func DecisionPath(id string) string { return DecisionDir + "/" + id + ".md" }

const decisionHeader = "<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->\n"

// Format is the one canonical form of a decision file. An empty section is
// written as "-", like an empty cell in a phase file.
func (d *Decision) Format() []byte {
	var b bytes.Buffer
	b.WriteString(decisionHeader)
	fmt.Fprintf(&b, "# %s: %s\n\nStatus: %s\nTags: %s\n", d.ID, d.Title, d.Status, d.Tags)
	for _, s := range []struct{ name, text string }{{"Context", d.Context}, {"Decision", d.Decision}, {"Consequences", d.Consequences}} {
		text := s.text
		if text == "" {
			text = "-"
		}
		fmt.Fprintf(&b, "\n## %s\n%s\n", s.name, text)
	}
	return b.Bytes()
}

// CheckDecisionTags refuses any tag but tech: a decision is about technology.
func CheckDecisionTags(t Tags) error {
	if len(t.Dec) > 0 || len(t.Path) > 0 {
		return errors.New("a decision takes only tech tags")
	}
	return nil
}

// CheckDecisionText reports why a multi-line section cannot be stored: a line
// starting with "## " would be read as a heading.
func CheckDecisionText(what, v string) error {
	if strings.TrimSpace(v) != v {
		return fmt.Errorf("%s must not start or end with spaces or blank lines", what)
	}
	if strings.Contains(v, "\r") {
		return fmt.Errorf("%s must not contain carriage returns", what)
	}
	for _, l := range strings.Split(v, "\n") {
		if strings.HasPrefix(l, "## ") {
			return fmt.Errorf("%s must not have a line starting with '## ', which would read as a heading", what)
		}
	}
	return nil
}

var decSections = []string{"Context", "Decision", "Consequences"}

// ParseDecision reads a decision file. It is strict about structure but not
// about spacing; Format shows the canonical form.
func ParseDecision(id string, data []byte) (*Decision, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	fail := func(line int, format string, a ...any) (*Decision, error) {
		return nil, &ParseError{Line: line, Msg: fmt.Sprintf(format, a...)}
	}
	i := 0
	skipBlank := func() {
		for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}
	}
	skipBlank()
	if i < len(lines) && strings.HasPrefix(lines[i], "<!--") {
		i++
	}
	skipBlank()
	if i >= len(lines) || !strings.HasPrefix(lines[i], "# ") {
		return fail(i+1, "expected '# %s: <title>'", id)
	}
	gotID, title, ok := strings.Cut(strings.TrimPrefix(lines[i], "# "), ": ")
	if !ok || title == "" || strings.TrimSpace(title) != title {
		return fail(i+1, "expected '# %s: <title>'", id)
	}
	if gotID != id {
		return fail(i+1, "heading names %s, but the file is %s.md", gotID, id)
	}
	d := &Decision{ID: id, Title: title}
	i++
	skipBlank()
	if i >= len(lines) || !strings.HasPrefix(lines[i], "Status: ") {
		return fail(i+1, "expected 'Status: ...'")
	}
	var err error
	if d.Status, err = parseDecStatus(strings.TrimPrefix(lines[i], "Status: ")); err != nil {
		return fail(i+1, "%v", err)
	}
	i++
	if i >= len(lines) || !strings.HasPrefix(lines[i], "Tags: ") {
		return fail(i+1, "expected 'Tags: ...' (use '-' for none)")
	}
	if d.Tags, err = ParseTags(strings.TrimSpace(strings.TrimPrefix(lines[i], "Tags: "))); err != nil {
		return fail(i+1, "%v", err)
	}
	if err := CheckDecisionTags(d.Tags); err != nil {
		return fail(i+1, "%v", err)
	}
	i++
	bodies := map[string]*string{"Context": &d.Context, "Decision": &d.Decision, "Consequences": &d.Consequences}
	for _, name := range decSections {
		skipBlank()
		if i >= len(lines) || lines[i] != "## "+name {
			return fail(i+1, "expected '## %s'", name)
		}
		i++
		start := i
		for i < len(lines) && !strings.HasPrefix(lines[i], "## ") {
			i++
		}
		body := strings.TrimSpace(strings.Join(lines[start:i], "\n"))
		if body == "-" {
			body = ""
		}
		*bodies[name] = body
	}
	if i < len(lines) {
		return fail(i+1, "unexpected section %q", lines[i])
	}
	return d, nil
}

// SetSuperseded marks a proposed or accepted decision as superseded by change,
// with by the replacement decision ("" when there is none). It keeps the date
// and name of the state it was in. The caller saves it with SaveDecision.
func (d *Decision) SetSuperseded(by, change string) error {
	if d.Status.State == DecSuperseded {
		return fmt.Errorf("%s is already superseded: %s", d.ID, d.Status)
	}
	if !ValidID("CHG", change) {
		return fmt.Errorf("%q is not a change ID such as CHG-0014", change)
	}
	if by != "" && (!ValidID(DecPrefix, by) || by == d.ID) {
		return fmt.Errorf("%q is not another decision ID such as DEC-0007", by)
	}
	d.Status = DecStatus{State: DecSuperseded, Date: d.Status.Date, By: d.Status.By,
		SupersededBy: by, Change: change, Was: d.Status.State}
	return nil
}

// Accept marks a proposed decision as accepted by a person on date.
func (d *Decision) Accept(by, date string) error {
	if d.Status.State != DecProposed {
		return fmt.Errorf("%s is already %s", d.ID, strings.ToLower(d.Status.State.String()))
	}
	d.Status = DecStatus{State: DecAccepted, Date: date, By: by}
	return nil
}

func decisionFiles(root string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(DecisionDir)))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if id, ok := strings.CutSuffix(e.Name(), ".md"); ok && !e.IsDir() && ValidID(DecPrefix, id) {
			ids = append(ids, id)
		}
	}
	SortIDs(ids)
	return ids, nil
}

// loadDecisions reads every DEC-xxxx.md. Files that do not parse are listed
// in BrokenDecisions rather than failing the load.
func (p *Plan) loadDecisions() error {
	ids, err := decisionFiles(p.Root)
	if err != nil {
		return err
	}
	p.RawDecisions, p.BrokenDecisions = map[string][]byte{}, map[string]*ParseError{}
	for _, id := range ids {
		data, err := os.ReadFile(filepath.Join(p.Root, filepath.FromSlash(DecisionPath(id))))
		if err != nil {
			return err
		}
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		p.RawDecisions[id] = data
		d, err := ParseDecision(id, data)
		if err != nil {
			var pe *ParseError
			if !errors.As(err, &pe) {
				return err
			}
			p.BrokenDecisions[id] = pe
			continue
		}
		p.Decisions = append(p.Decisions, d)
	}
	return nil
}

// Decision returns the decision with the given ID, or nil.
func (p *Plan) Decision(id string) *Decision {
	for _, d := range p.Decisions {
		if d.ID == id {
			return d
		}
	}
	return nil
}

// NextDecisionID returns the next free decision ID: one past the highest file
// name in use, readable or not. Decisions are never deleted, so IDs are permanent.
func (p *Plan) NextDecisionID() string {
	high := 0
	for id := range p.RawDecisions {
		if _, n, ok := SplitID(id); ok && n > high {
			high = n
		}
	}
	return FormatID(DecPrefix, high+1)
}

// SaveDecision writes a decision file if its canonical form differs from what is on disk.
func (p *Plan) SaveDecision(d *Decision) error {
	data := d.Format()
	if bytes.Equal(p.RawDecisions[d.ID], data) {
		return nil
	}
	if err := writeFile(filepath.Join(p.Root, filepath.FromSlash(DecisionPath(d.ID))), data); err != nil {
		return err
	}
	if p.RawDecisions == nil {
		p.RawDecisions = map[string][]byte{}
	}
	if p.Decision(d.ID) == nil {
		p.Decisions = append(p.Decisions, d)
		sort.Slice(p.Decisions, func(i, j int) bool { return LessID(p.Decisions[i].ID, p.Decisions[j].ID) })
	}
	p.RawDecisions[d.ID] = data
	return nil
}

// DecisionTickets returns the tickets that link decision id with a dec tag, by ID.
func (p *Plan) DecisionTickets(id string) []*Ticket {
	var out []*Ticket
	for _, t := range p.Tickets() {
		for _, v := range t.Tags.Dec {
			if v == id {
				out = append(out, t)
			}
		}
	}
	return out
}
