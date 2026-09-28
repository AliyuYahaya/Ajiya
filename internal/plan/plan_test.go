package plan

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Reading a canonical file and writing it back gives the same bytes.
func TestGoldenRoundTrip(t *testing.T) {
	files, err := filepath.Glob("testdata/golden/*.md")
	if err != nil || len(files) == 0 {
		t.Fatalf("no golden files: %v", err)
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			want, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			ph, err := ParsePhase("x", want)
			if err != nil {
				t.Fatal(err)
			}
			if got := ph.Format(); string(got) != string(want) {
				t.Errorf("round trip changed the file:\n--- got\n%s\n--- want\n%s", got, want)
			}
		})
	}
}

func TestParseSpecExample(t *testing.T) {
	data, _ := os.ReadFile("testdata/golden/bookings.md")
	ph, err := ParsePhase("bookings", data)
	if err != nil {
		t.Fatal(err)
	}
	if ph.Title != "Bookings" || ph.Goal != "A patient can book a slot end to end" || len(ph.Tickets) != 4 {
		t.Fatalf("unexpected phase: %+v", ph)
	}
	t7 := ph.Tickets[0]
	want := Status{State: Done, Commit: "4f2a91c", Date: "2026-09-28", Tests: true}
	if t7.ID != "CB-0007" || t7.App != "api" || !reflect.DeepEqual(t7.Depends, []string{"CB-0002"}) ||
		t7.Status != want || t7.Line != 8 || t7.Phase != "bookings" {
		t.Errorf("CB-0007 = %+v", t7)
	}
	if s := ph.Tickets[1].Status; s.State != InProgress || s.Note != "validation done, tests to do" {
		t.Errorf("CB-0008 status = %+v", s)
	}
	if s := ph.Tickets[3].Status; s.State != Pending || s.Human != "hosting account" || len(ph.Tickets[3].Depends) != 0 {
		t.Errorf("CB-0010 = %+v", ph.Tickets[3])
	}
}

func TestParseEscapedPipes(t *testing.T) {
	data, _ := os.ReadFile("testdata/golden/every-status.md")
	ph, err := ParsePhase("x", data)
	if err != nil {
		t.Fatal(err)
	}
	if got := ph.Tickets[0].Title; got != "Parse a | b" {
		t.Errorf("title = %q", got)
	}
}

// Hand edits that still parse come out in canonical form.
func TestFormatCanonicalises(t *testing.T) {
	in := "# Bookings\nGoal:   Book a slot\n\n|ID|App|Ticket|Done when|Depends|Status|\n| :-- | --- | --- | --- | --- | --- |\n" +
		"| CB-0009 | web | B | x | CB-0008,CB-0007 CB-0007 | 🟥 Pending |\n" +
		"|CB-0002|api|A|x|-|🟩 Done · 2026-09-28 · 4f2a91c|\n"
	ph, err := ParsePhase("bookings", []byte(in))
	if err != nil {
		t.Fatal(err)
	}
	want := Header + "\n# Bookings\n\nGoal: Book a slot\n\n| ID | App | Ticket | Done when | Depends | Status |\n|---|---|---|---|---|---|\n" +
		"| CB-0002 | api | A | x | - | 🟩 Done · 4f2a91c · 2026-09-28 |\n" +
		"| CB-0009 | web | B | x | CB-0007, CB-0008 | 🟥 Pending |\n"
	if got := string(ph.Format()); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestParseErrors(t *testing.T) {
	const top = Header + "\n# P\n\nGoal: g\n\n| ID | App | Ticket | Done when | Depends | Status |\n|---|---|---|---|---|---|\n"
	tests := []struct {
		name, in string
		line     int
		want     string
	}{
		{"empty", "", 1, "phase title"},
		{"no goal", Header + "\n# P\n\n| ID |\n", 4, "Goal:"},
		{"no table", Header + "\n# P\n\nGoal: g\n", 4, "ticket table"},
		{"wrong header", Header + "\n# P\n\nGoal: g\n\n| ID | Ticket |\n|---|---|\n", 6, "table header"},
		{"no separator", Header + "\n# P\n\nGoal: g\n\n| ID | App | Ticket | Done when | Depends | Status |\n| a | b | c | d | e | f |\n", 7, "separator"},
		{"short row", top + "| CB-0001 | api | t | d | 🟥 Pending |\n", 8, "has 5 cells"},
		{"text after", top + "| CB-0001 | api | t | d | - | 🟥 Pending |\nSome notes\n", 9, "unexpected text"},
		{"no id", top + "|  | api | t | d | - | 🟥 Pending |\n", 8, "no ID"},
		{"bad mark", top + "| CB-0001 | api | t | d | - | Pending |\n", 8, "must start with"},
		{"bad part", top + "| CB-0001 | api | t | d | - | 🟥 Pending · soon |\n", 8, `"soon" is not understood`},
		{"hash on pending", top + "| CB-0001 | api | t | d | - | 🟥 Pending · 4f2a91c |\n", 8, "not understood"},
		{"repeated", top + "| CB-0001 | api | t | d | - | 🟩 Done · 2026-01-01 · 2026-01-02 |\n", 8, "repeated"},
		{"dropped no name", top + "| CB-0001 | api | t | d | - | 🟩 Dropped: no longer needed |\n", 8, "decided by"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePhase("p", []byte(tt.in))
			pe, ok := err.(*ParseError)
			if !ok {
				t.Fatalf("err = %v, want a ParseError", err)
			}
			if pe.Line != tt.line || !strings.Contains(pe.Msg, tt.want) {
				t.Errorf("err = %q at line %d, want %q at line %d", pe.Msg, pe.Line, tt.want, tt.line)
			}
		})
	}
}

func TestStatusRoundTrip(t *testing.T) {
	for _, s := range []Status{
		{State: Pending},
		{State: Pending, Human: "h", Blocked: "b"},
		{State: InProgress},
		{State: InProgress, Note: "n", Human: "h"},
		{State: Done, Commit: "abcdef1", Date: "2026-01-01", Tests: true, Note: "n"},
		{State: Done, By: "A Person", Date: "2026-01-01", Note: "n"},
		{State: Dropped, Reason: "r, decided by nobody", By: "Me"},
	} {
		got, err := ParseStatus(s.String())
		if err != nil || got != s {
			t.Errorf("ParseStatus(%q) = %+v, %v; want %+v", s.String(), got, err, s)
		}
	}
}

func TestIDs(t *testing.T) {
	for id, want := range map[string]bool{"CB-0001": true, "CB-12345": true, "CB-001": false, "cb-0001": false, "XY-0001": false, "CB-00001": false} {
		if got := ValidID("CB", id); got != want {
			t.Errorf("ValidID(CB, %q) = %v", id, got)
		}
	}
	ids := ParseIDList("CB-0010, CB-0002 CB-0010,-, X, CB-10000")
	if got := strings.Join(ids, " "); got != "CB-0002 CB-0010 CB-10000 X" {
		t.Errorf("ParseIDList = %s", got)
	}
	if ParseIDList("-") != nil || ParseIDList("") != nil {
		t.Error("ParseIDList(-) is not empty")
	}
}

func tk(id string, deps ...string) *Ticket { return &Ticket{ID: id, Depends: deps} }

func TestCycles(t *testing.T) {
	g := NewGraph([]*Ticket{
		tk("CB-0001", "CB-0003"), tk("CB-0002", "CB-0001"), tk("CB-0003", "CB-0002"), // loop 1-3-2-1
		tk("CB-0004", "CB-0004"),            // self: not a loop here
		tk("CB-0005", "CB-0006", "CB-0099"), // unknown ignored
		tk("CB-0006", "CB-0005"),            // loop 5-6-5
		tk("CB-0007", "CB-0001"),            // hangs off a loop, not in it
	})
	var got []string
	for _, c := range g.Cycles() {
		got = append(got, strings.Join(c, ">"))
	}
	want := []string{"CB-0001>CB-0003>CB-0002>CB-0001", "CB-0005>CB-0006>CB-0005"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Cycles = %v, want %v", got, want)
	}
}

func TestDownstreamAndLoopWith(t *testing.T) {
	ts := []*Ticket{tk("CB-0001"), tk("CB-0002", "CB-0001"), tk("CB-0003", "CB-0002"), tk("CB-0004", "CB-0001", "CB-0003")}
	g := NewGraph(ts)
	if got := strings.Join(g.Downstream("CB-0001"), " "); got != "CB-0002 CB-0003 CB-0004" {
		t.Errorf("Downstream = %s", got)
	}
	if got := strings.Join(g.Dependants("CB-0001"), " "); got != "CB-0002 CB-0004" {
		t.Errorf("Dependants = %s", got)
	}
	if loop := LoopWith(ts, "CB-0001", []string{"CB-0004"}); strings.Join(loop, ">") != "CB-0001>CB-0004>CB-0001" {
		t.Errorf("LoopWith = %v", loop)
	}
	if loop := LoopWith(ts, "CB-0004", []string{"CB-0002"}); loop != nil {
		t.Errorf("LoopWith found a loop that is not there: %v", loop)
	}
	if ts[3].Depends[0] != "CB-0001" {
		t.Error("LoopWith changed the input")
	}
}

func TestLoadSaveNextID(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, Dir)
	os.MkdirAll(dir, 0o755)
	golden, _ := os.ReadFile("testdata/golden/bookings.md")
	crlf := strings.ReplaceAll(string(golden), "\n", "\r\n")
	os.WriteFile(filepath.Join(dir, "bookings.md"), []byte(crlf), 0o644)
	os.WriteFile(filepath.Join(dir, "broken.md"), []byte("nonsense\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "PROGRESS.md"), []byte("generated\n"), 0o644)

	p, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Phases) != 1 || p.Broken["broken"] == nil || len(p.Broken) != 1 {
		t.Fatalf("phases %d, broken %v", len(p.Phases), p.Broken)
	}
	if string(p.Raw["bookings"]) != string(golden) {
		t.Error("CRLF was not normalised")
	}
	if got := p.NextID("CB"); got != "CB-0011" {
		t.Errorf("NextID = %s", got)
	}
	if p.Ticket("CB-0009") == nil || p.Ticket("CB-0999") != nil {
		t.Error("Ticket lookup wrong")
	}

	// Saving an unchanged phase leaves the CRLF file alone.
	if err := p.Save(p.Phases[0]); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "bookings.md")); string(b) != crlf {
		t.Error("unchanged phase was rewritten")
	}

	// Moving a ticket writes both files.
	golive := &Phase{Slug: "go-live", Title: TitleFromSlug("go-live"), Goal: "Live"}
	p.Move(p.Ticket("CB-0010"), golive)
	if err := p.Save(golive); err != nil {
		t.Fatal(err)
	}
	if err := p.Save(p.Phase("bookings")); err != nil {
		t.Fatal(err)
	}
	p2, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := p2.Ticket("CB-0010"); got == nil || got.Phase != "go-live" || len(p2.Phase("bookings").Tickets) != 3 {
		t.Errorf("move not saved: %+v", got)
	}
	if p2.Phase("go-live").Title != "Go-live" {
		t.Errorf("title = %q", p2.Phase("go-live").Title)
	}
}
