package plan

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ChangeDir is the folder holding the change files, relative to the project root.
const ChangeDir = Dir + "/changes"

// ChgPrefix is the prefix of change IDs: CHG-0014.
const ChgPrefix = "CHG"

// ChangeState is where a change request stands.
type ChangeState int

const (
	ChgOpen ChangeState = iota
	ChgApproved
	ChgRejected
)

func (s ChangeState) String() string {
	switch s {
	case ChgApproved:
		return "Approved"
	case ChgRejected:
		return "Rejected"
	}
	return "Open"
}

// ChangeStatus is the Status line of a change file.
//
//	Open · 2026-10-16
//	Approved · 2026-10-20 · by Aliyu
//	Rejected · 2026-10-20 · by Aliyu · the move was cancelled
//
// Date is when the change was opened (Open) or closed (Approved, Rejected).
// The opening date of a closed change is not kept in this line; the snapshot
// tag and git history hold it.
type ChangeStatus struct {
	State  ChangeState
	Date   string
	By     string // who approved or rejected; empty while Open
	Reason string // why it was rejected; empty otherwise
}

func (s ChangeStatus) String() string {
	parts := []string{s.State.String(), s.Date}
	if s.State != ChgOpen {
		parts = append(parts, "by "+s.By)
	}
	if s.State == ChgRejected {
		parts = append(parts, s.Reason)
	}
	return strings.Join(parts, " · ")
}

// Open reports whether the change is still being worked on.
func (s ChangeStatus) Open() bool { return s.State == ChgOpen }

func parseChangeStatus(text string) (ChangeStatus, error) {
	var s ChangeStatus
	parts := strings.Split(text, " · ")
	if len(parts) < 2 || !dateRE.MatchString(parts[1]) {
		return s, fmt.Errorf("status %q: want 'Open · YYYY-MM-DD', 'Approved · YYYY-MM-DD · by NAME' or 'Rejected · YYYY-MM-DD · by NAME · REASON'", text)
	}
	s.Date = parts[1]
	rest := parts[2:]
	switch parts[0] {
	case "Open":
		if len(rest) != 0 {
			return s, errors.New("an Open status is just 'Open · YYYY-MM-DD'")
		}
		return s, nil
	case "Approved":
		s.State = ChgApproved
	case "Rejected":
		s.State = ChgRejected
	default:
		return s, fmt.Errorf("status %q is not Open, Approved or Rejected", parts[0])
	}
	if len(rest) == 0 || !strings.HasPrefix(rest[0], "by ") || len(rest[0]) == 3 {
		return s, fmt.Errorf("a %s status needs '· by NAME' after the date", strings.ToLower(parts[0]))
	}
	s.By = strings.TrimPrefix(rest[0], "by ")
	rest = rest[1:]
	if s.State == ChgApproved {
		if len(rest) != 0 {
			return s, errors.New("an Approved status is just 'Approved · YYYY-MM-DD · by NAME'")
		}
		return s, nil
	}
	if len(rest) == 0 {
		return s, errors.New("a Rejected status needs '· REASON' after the name")
	}
	s.Reason = strings.Join(rest, " · ")
	return s, nil
}

// Verdict is the agent's judgement of one candidate. "" means none yet.
type Verdict string

const (
	VerdictUnaffected  Verdict = "unaffected"
	VerdictModify      Verdict = "modify"
	VerdictReplace     Verdict = "replace"
	VerdictInvalidate  Verdict = "invalidate"
	VerdictInvestigate Verdict = "investigate"
)

// Verdicts lists the verdicts, and nothing else is accepted.
var Verdicts = []Verdict{VerdictUnaffected, VerdictModify, VerdictReplace, VerdictInvalidate, VerdictInvestigate}

// ParseVerdict returns the verdict named s, or an error naming the choices.
func ParseVerdict(s string) (Verdict, error) {
	for _, v := range Verdicts {
		if string(v) == s {
			return v, nil
		}
	}
	names := make([]string, len(Verdicts))
	for i, v := range Verdicts {
		names[i] = string(v)
	}
	return "", fmt.Errorf("verdict %q is not one of %s", s, strings.Join(names, ", "))
}

// Reason kinds: why a ticket is a candidate. A reason is written kind:detail.
const (
	ReasonTag        = "tag"        // detail: key:value, a replaced tag the ticket has
	ReasonDecision   = "decision"   // detail: DEC-xxxx, a named decision the ticket links
	ReasonPath       = "path"       // detail: a path tag, or a file from its commits, matching a --paths glob
	ReasonDownstream = "downstream" // detail: the candidate ticket it directly depends on
)

// StaleMark is written after the reasons of a candidate that kept its verdict
// but no longer qualifies on a later impact run.
const StaleMark = "stale"

var reasonKinds = []string{ReasonTag, ReasonDecision, ReasonPath, ReasonDownstream}

// Candidate is one ticket a change might affect: a row of the Candidates table.
type Candidate struct {
	ID      string
	Title   string   // the ticket's title when impact last ran
	Reasons []string // kind:detail, in the order found
	// Stale is set when a later impact run no longer finds the ticket but a
	// verdict was already given, so the row is kept.
	Stale       bool
	Verdict     Verdict // "" until assessed
	Why         string  // the reason given with the verdict
	NewTitle    string  // title of the ticket a replace or modify creates
	NewDoneWhen string  // its done-when
}

// Change is one change request, ajiya/changes/CHG-xxxx.md.
type Change struct {
	ID         string
	Title      string
	Status     ChangeStatus
	Replaces   Tags     // tickets with these tags are candidates
	With       Tags     // what replaces them; the agent's guide for new tickets
	Decisions  []string // decisions the change would supersede
	Paths      []string // file globs the change touches
	Snapshot   string   // snapshot name, without the ajiya/plan/ prefix
	Candidates []Candidate
	// Results lists what approving did: tickets created, superseded, edited.
	// Empty until the change is approved.
	Results string
}

// ChangePath is the path of a change file, relative to the project root, with / separators.
func ChangePath(id string) string { return ChangeDir + "/" + id + ".md" }

// SnapshotName is the snapshot that change open makes for a change.
func SnapshotName(id string) string { return "before-" + id }

const changeHeader = "<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->\n"

const candidateHead = "| ID | Ticket | Reasons | Verdict | Verdict reason | New title | New done when |"
const candidateSep = "|---|---|---|---|---|---|---|"

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func undash(s string) string {
	if s == "-" {
		return ""
	}
	return s
}

func listOrDash(l []string) string { return dash(strings.Join(l, " ")) }

// reasonsCell is the Reasons cell of a candidate.
func (c Candidate) reasonsCell() string {
	r := append([]string(nil), c.Reasons...)
	if c.Stale {
		r = append(r, StaleMark)
	}
	return listOrDash(r)
}

// Format is the one canonical form of a change file. An empty value is
// written as "-", like an empty cell in a phase file.
func (c *Change) Format() []byte {
	var b bytes.Buffer
	b.WriteString(changeHeader)
	fmt.Fprintf(&b, "# %s: %s\n\nStatus: %s\n", c.ID, c.Title, c.Status)
	fmt.Fprintf(&b, "Replaces: %s\nWith: %s\nDecisions: %s\nPaths: %s\nSnapshot: %s\n",
		c.Replaces, c.With, listOrDash(c.Decisions), listOrDash(c.Paths), dash(c.Snapshot))
	b.WriteString("\n## Candidates\n" + candidateHead + "\n" + candidateSep + "\n")
	for _, k := range c.Candidates {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s |\n", k.ID, escape(k.Title), escape(k.reasonsCell()),
			dash(string(k.Verdict)), escape(dash(k.Why)), escape(dash(k.NewTitle)), escape(dash(k.NewDoneWhen)))
	}
	fmt.Fprintf(&b, "\n## Results\n%s\n", dash(c.Results))
	return b.Bytes()
}

// CheckChangeText reports why a multi-line section cannot be stored (see CheckDecisionText).
func CheckChangeText(what, v string) error { return CheckDecisionText(what, v) }

func parseReasons(cell string) (reasons []string, stale bool, err error) {
	if cell == "-" {
		return nil, false, nil
	}
	for _, tok := range strings.Fields(cell) {
		if tok == StaleMark {
			stale = true
			continue
		}
		kind, detail, ok := strings.Cut(tok, ":")
		known := false
		for _, k := range reasonKinds {
			known = known || k == kind
		}
		if !ok || !known || detail == "" || stale {
			return nil, false, fmt.Errorf("reason %q is not one of tag:, decision:, path: or downstream: followed by a value (and %q comes last)", tok, StaleMark)
		}
		reasons = append(reasons, tok)
	}
	return reasons, stale, nil
}

// ParseChange reads a change file. It is strict about structure but not about
// spacing; Format shows the canonical form.
func ParseChange(id string, data []byte) (*Change, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	fail := func(line int, format string, a ...any) (*Change, error) {
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
	c := &Change{ID: id, Title: title}
	i++
	skipBlank()
	field := func(name string) (string, bool) {
		if i >= len(lines) || !strings.HasPrefix(lines[i], name+": ") {
			return "", false
		}
		v := strings.TrimSpace(strings.TrimPrefix(lines[i], name+": "))
		i++
		return v, true
	}
	var err error
	v, ok := field("Status")
	if !ok {
		return fail(i+1, "expected 'Status: ...'")
	}
	if c.Status, err = parseChangeStatus(v); err != nil {
		return fail(i, "%v", err)
	}
	for _, name := range []string{"Replaces", "With"} {
		v, ok := field(name)
		if !ok {
			return fail(i+1, "expected '%s: ...' (use '-' for none)", name)
		}
		t, err := ParseTags(v)
		if err != nil {
			return fail(i, "%s: %v", name, err)
		}
		if name == "Replaces" {
			c.Replaces = t
		} else {
			c.With = t
		}
	}
	if v, ok = field("Decisions"); !ok {
		return fail(i+1, "expected 'Decisions: ...' (use '-' for none)")
	}
	for _, d := range ParseIDList(v) {
		if !ValidID(DecPrefix, d) {
			return fail(i, "%q is not a decision ID such as DEC-0003", d)
		}
		c.Decisions = append(c.Decisions, d)
	}
	if v, ok = field("Paths"); !ok {
		return fail(i+1, "expected 'Paths: ...' (use '-' for none)")
	}
	if v != "-" {
		for _, g := range strings.Fields(v) {
			if err := CheckTag(TagPath, g); err != nil {
				return fail(i, "%v", err)
			}
			c.Paths = append(c.Paths, g)
		}
	}
	if v, ok = field("Snapshot"); !ok {
		return fail(i+1, "expected 'Snapshot: ...'")
	}
	c.Snapshot = undash(v)
	skipBlank()
	if i >= len(lines) || lines[i] != "## Candidates" {
		return fail(i+1, "expected '## Candidates'")
	}
	i++
	if i >= len(lines) || strings.Join(strings.Fields(lines[i]), "") != strings.Join(strings.Fields(candidateHead), "") {
		return fail(i+1, "expected the table header %s", candidateHead)
	}
	i++
	if cells, ok := splitRowAt(lines, i); !ok || !isSeparator(cells) {
		return fail(i+1, "expected the table separator %s", candidateSep)
	}
	i++
	seen := map[string]bool{}
	for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "|") {
		cells, ok := splitRow(lines[i])
		if !ok || len(cells) != 7 {
			return fail(i+1, "a candidate row needs 7 cells: ID, Ticket, Reasons, Verdict, Verdict reason, New title, New done when")
		}
		k := Candidate{ID: cells[0], Title: unescape(cells[1])}
		if _, _, ok := SplitID(k.ID); !ok || k.Title == "" {
			return fail(i+1, "candidate %q needs a ticket ID and a title", k.ID)
		}
		if seen[k.ID] {
			return fail(i+1, "%s is listed twice", k.ID)
		}
		seen[k.ID] = true
		if k.Reasons, k.Stale, err = parseReasons(unescape(cells[2])); err != nil {
			return fail(i+1, "%v", err)
		}
		if cells[3] != "-" {
			if k.Verdict, err = ParseVerdict(cells[3]); err != nil {
				return fail(i+1, "%v", err)
			}
		}
		k.Why, k.NewTitle, k.NewDoneWhen = undash(unescape(cells[4])), undash(unescape(cells[5])), undash(unescape(cells[6]))
		c.Candidates = append(c.Candidates, k)
		i++
	}
	skipBlank()
	if i >= len(lines) || lines[i] != "## Results" {
		return fail(i+1, "expected '## Results'")
	}
	i++
	start := i
	for i < len(lines) && !strings.HasPrefix(lines[i], "## ") {
		i++
	}
	c.Results = undash(strings.TrimSpace(strings.Join(lines[start:i], "\n")))
	if i < len(lines) {
		return fail(i+1, "unexpected section %q", lines[i])
	}
	return c, nil
}

func splitRowAt(lines []string, i int) ([]string, bool) {
	if i >= len(lines) {
		return nil, false
	}
	return splitRow(lines[i])
}

// Candidate returns the candidate row for a ticket, or nil.
func (c *Change) Candidate(ticket string) *Candidate {
	for i := range c.Candidates {
		if c.Candidates[i].ID == ticket {
			return &c.Candidates[i]
		}
	}
	return nil
}

// SetVerdict records the agent's judgement of a candidate: its verdict, the
// reason for it and, for a replacement, the title and done-when of the new
// ticket (pass "" when there is none). It replaces any earlier judgement, and
// only an Open change takes one. The caller saves the change with SaveChange.
func (c *Change) SetVerdict(ticket string, v Verdict, why, newTitle, newDoneWhen string) error {
	if !c.Status.Open() {
		return fmt.Errorf("%s is %s, so its verdicts are fixed", c.ID, strings.ToLower(c.Status.State.String()))
	}
	k := c.Candidate(ticket)
	if k == nil {
		return fmt.Errorf("%s is not a candidate of %s; run 'ajiya change impact %s' to find candidates", ticket, c.ID, c.ID)
	}
	if _, err := ParseVerdict(string(v)); err != nil {
		return err
	}
	if why == "" {
		return errors.New("a verdict needs a reason")
	}
	for _, f := range []struct{ name, v string }{{"reason", why}, {"new title", newTitle}, {"new done-when", newDoneWhen}} {
		if err := CheckText(f.name, f.v); err != nil {
			return err
		}
	}
	k.Verdict, k.Why, k.NewTitle, k.NewDoneWhen = v, why, newTitle, newDoneWhen
	return nil
}

// Found is a ticket impact picked, with why.
type Found struct {
	ID      string
	Reasons []string
}

// FindImpact lists the tickets the change might affect, deterministically, by
// ticket ID. Dropped and superseded tickets are skipped. files maps a ticket ID
// to the files its commits changed, as paths from the project root (nil when
// not needed: path reasons then come from path tags only).
//
//	tag        the ticket has a tag listed in Replaces
//	decision   the ticket links a decision in Decisions with a dec tag
//	path       a path tag of the ticket overlaps a Paths glob, or a file from
//	           its commits matches one
//	downstream the ticket depends directly on a ticket picked for one of the
//	           reasons above (one level only)
//
// A ticket with no tags is found through its commits' files or a dependency.
func (p *Plan) FindImpact(c *Change, files map[string][]string) []Found {
	reasons := map[string][]string{}
	add := func(id, r string) {
		for _, have := range reasons[id] {
			if have == r {
				return
			}
		}
		reasons[id] = append(reasons[id], r)
	}
	live := map[string]*Ticket{}
	var ids []string
	for _, t := range p.Tickets() {
		if t.Status.Excluded() {
			continue
		}
		live[t.ID] = t
		ids = append(ids, t.ID)
	}
	for _, id := range ids {
		t := live[id]
		for _, key := range TagKeys {
			for _, v := range c.Replaces.Values(key) {
				for _, have := range t.Tags.Values(key) {
					if have == v {
						add(id, ReasonTag+":"+key+":"+v)
					}
				}
			}
		}
		for _, d := range c.Decisions {
			for _, have := range t.Tags.Dec {
				if have == d {
					add(id, ReasonDecision+":"+d)
				}
			}
		}
		fs := append([]string(nil), files[id]...)
		sort.Strings(fs)
		for _, g := range c.Paths {
			for _, pt := range t.Tags.Path {
				if GlobsOverlap(g, pt) {
					add(id, ReasonPath+":"+pt)
				}
			}
			for _, f := range fs {
				if GlobMatch(g, f) {
					add(id, ReasonPath+":"+strings.ReplaceAll(f, " ", "%20"))
					break
				}
			}
		}
	}
	base := map[string]bool{}
	for id := range reasons {
		base[id] = true
	}
	for _, id := range ids {
		var deps []string
		for _, d := range live[id].Depends {
			if base[d] && d != id {
				deps = append(deps, d)
			}
		}
		SortIDs(deps)
		for _, d := range deps {
			add(id, ReasonDownstream+":"+d)
		}
	}
	var out []Found
	for _, id := range ids {
		if r := reasons[id]; len(r) > 0 {
			out = append(out, Found{id, r})
		}
	}
	return out
}

// MergeResult says what MergeCandidates did, by ticket ID.
type MergeResult struct {
	Added   []string // new candidates
	Kept    []string // candidates found again (verdicts untouched)
	Stale   []string // no longer found, but kept because they have a verdict
	Dropped []string // no longer found and had no verdict, so removed
}

// MergeCandidates writes impact results into the candidate table. A candidate
// found again keeps its verdict, reason, new title and new done-when, and gets
// the freshly found reasons and the ticket's current title. A candidate no
// longer found keeps its row and verdict, marked stale, when it has a verdict,
// and is dropped when it has none. The table ends sorted by ticket ID.
func (c *Change) MergeCandidates(p *Plan, found []Found) MergeResult {
	var res MergeResult
	byID := map[string]Found{}
	for _, f := range found {
		byID[f.ID] = f
	}
	var out []Candidate
	for _, k := range c.Candidates {
		if f, ok := byID[k.ID]; ok {
			k.Reasons, k.Stale = f.Reasons, false
			if t := p.Ticket(k.ID); t != nil {
				k.Title = t.Title
			}
			out = append(out, k)
			res.Kept = append(res.Kept, k.ID)
			delete(byID, k.ID)
			continue
		}
		if k.Verdict != "" {
			k.Stale = true
			out = append(out, k)
			res.Stale = append(res.Stale, k.ID)
		} else {
			res.Dropped = append(res.Dropped, k.ID)
		}
	}
	for _, f := range found {
		if _, ok := byID[f.ID]; !ok {
			continue
		}
		title := ""
		if t := p.Ticket(f.ID); t != nil {
			title = t.Title
		}
		out = append(out, Candidate{ID: f.ID, Title: title, Reasons: f.Reasons})
		res.Added = append(res.Added, f.ID)
	}
	sort.SliceStable(out, func(i, j int) bool { return LessID(out[i].ID, out[j].ID) })
	c.Candidates = out
	return res
}

func changeFiles(root string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(ChangeDir)))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if id, ok := strings.CutSuffix(e.Name(), ".md"); ok && !e.IsDir() && ValidID(ChgPrefix, id) {
			ids = append(ids, id)
		}
	}
	SortIDs(ids)
	return ids, nil
}

// loadChanges reads every CHG-xxxx.md. Files that do not parse are listed in
// BrokenChanges rather than failing the load.
func (p *Plan) loadChanges() error {
	ids, err := changeFiles(p.Root)
	if err != nil {
		return err
	}
	p.RawChanges, p.BrokenChanges = map[string][]byte{}, map[string]*ParseError{}
	for _, id := range ids {
		data, err := os.ReadFile(filepath.Join(p.Root, filepath.FromSlash(ChangePath(id))))
		if err != nil {
			return err
		}
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		p.RawChanges[id] = data
		c, err := ParseChange(id, data)
		if err != nil {
			var pe *ParseError
			if !errors.As(err, &pe) {
				return err
			}
			p.BrokenChanges[id] = pe
			continue
		}
		p.Changes = append(p.Changes, c)
	}
	return nil
}

// Change returns the change with the given ID, or nil.
func (p *Plan) Change(id string) *Change {
	for _, c := range p.Changes {
		if c.ID == id {
			return c
		}
	}
	return nil
}

// NextChangeID returns the next free change ID: one past the highest file name
// in use, readable or not. Changes are never deleted, so IDs are permanent.
func (p *Plan) NextChangeID() string {
	high := 0
	for id := range p.RawChanges {
		if _, n, ok := SplitID(id); ok && n > high {
			high = n
		}
	}
	return FormatID(ChgPrefix, high+1)
}

// SaveChange writes a change file if its canonical form differs from what is on disk.
func (p *Plan) SaveChange(c *Change) error {
	data := c.Format()
	if bytes.Equal(p.RawChanges[c.ID], data) {
		return nil
	}
	if err := writeFile(filepath.Join(p.Root, filepath.FromSlash(ChangePath(c.ID))), data); err != nil {
		return err
	}
	if p.RawChanges == nil {
		p.RawChanges = map[string][]byte{}
	}
	if p.Change(c.ID) == nil {
		p.Changes = append(p.Changes, c)
		sort.Slice(p.Changes, func(i, j int) bool { return LessID(p.Changes[i].ID, p.Changes[j].ID) })
	}
	p.RawChanges[c.ID] = data
	return nil
}

// TicketChanges returns the changes that list a ticket as a candidate, by change ID.
func (p *Plan) TicketChanges(ticket string) []*Change {
	var out []*Change
	for _, c := range p.Changes {
		if c.Candidate(ticket) != nil {
			out = append(out, c)
		}
	}
	return out
}
