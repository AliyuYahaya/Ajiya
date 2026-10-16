package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Reading a canonical decision file and writing it back gives the same bytes.
func TestDecisionGoldenRoundTrip(t *testing.T) {
	files, err := filepath.Glob("testdata/golden-decisions/DEC-*.md")
	if err != nil || len(files) == 0 {
		t.Fatalf("no golden files: %v", err)
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			want, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			id := strings.TrimSuffix(filepath.Base(f), ".md")
			d, err := ParseDecision(id, want)
			if err != nil {
				t.Fatal(err)
			}
			if got := d.Format(); string(got) != string(want) {
				t.Errorf("round trip changed the file:\n--- got\n%s\n--- want\n%s", got, want)
			}
		})
	}
}

func TestParseDecisionFields(t *testing.T) {
	data, _ := os.ReadFile("testdata/golden-decisions/DEC-0003.md")
	d, err := ParseDecision("DEC-0003", data)
	if err != nil {
		t.Fatal(err)
	}
	if d.Title != "Use PostgreSQL on AWS RDS" || d.Status.State != DecAccepted || d.Status.Date != "2026-09-12" || d.Status.By != "Aliyu" {
		t.Errorf("decision = %+v", d)
	}
	if strings.Join(d.Tags.Tech, ",") != "aws-rds,postgres" {
		t.Errorf("tags = %v", d.Tags)
	}
	if !strings.Contains(d.Context, "\n\nBackups must be daily.") || d.Consequences == "" {
		t.Errorf("sections = %q / %q", d.Context, d.Consequences)
	}
}

func TestDecisionSupersededAndAccept(t *testing.T) {
	data, _ := os.ReadFile("testdata/golden-decisions/DEC-0003.md")
	d, _ := ParseDecision("DEC-0003", data)
	if err := d.Accept("X", "2026-01-01"); err == nil {
		t.Error("accepting an accepted decision was allowed")
	}
	if err := d.SetSuperseded("DEC-0007", "CHG-0014"); err != nil {
		t.Fatal(err)
	}
	want := "Superseded by DEC-0007 (CHG-0014) · was: Accepted · 2026-09-12 · decided by Aliyu"
	if d.Status.String() != want || d.Status.Open() {
		t.Errorf("status = %q", d.Status)
	}
	again, err := ParseDecision("DEC-0003", d.Format())
	if err != nil || again.Status != d.Status {
		t.Errorf("reparsed = %+v, %v", again.Status, err)
	}
	if err := d.SetSuperseded("", "CHG-0015"); err == nil {
		t.Error("superseding twice was allowed")
	}
	p := &Decision{ID: "DEC-0001", Title: "t"}
	if err := p.Accept("Aliyu", "2026-02-03"); err != nil || p.Status.String() != "Accepted · 2026-02-03 · decided by Aliyu" {
		t.Errorf("accept: %v %q", err, p.Status)
	}
	if err := (&Decision{ID: "DEC-0002"}).SetSuperseded("DEC-0002", "CHG-0001"); err == nil {
		t.Error("a decision superseded by itself was allowed")
	}
}

func TestParseDecisionErrors(t *testing.T) {
	good, _ := os.ReadFile("testdata/golden-decisions/DEC-0003.md")
	for name, c := range map[string]struct{ from, to, msg string }{
		"wrong id":       {"# DEC-0003:", "# DEC-0009:", "heading names DEC-0009"},
		"bad status":     {"Status: Accepted", "Status: Maybe", "is not Proposed, Accepted or Superseded"},
		"no was":         {"Status: Accepted · 2026-09-12 · decided by Aliyu", "Status: Superseded by DEC-0007 (CHG-0014)", "was:"},
		"bad tag":        {"tech:postgres", "comp:db", "tag key"},
		"dec tag":        {"tech:postgres", "dec:DEC-0001", "only tech"},
		"missing a part": {"## Consequences", "## Result", "expected '## Consequences'"},
		"wrong verb":     {"decided by", "proposed by", "is not a date"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ParseDecision("DEC-0003", []byte(strings.Replace(string(good), c.from, c.to, 1)))
			if err == nil || !strings.Contains(err.Error(), c.msg) {
				t.Errorf("err = %v, want %q", err, c.msg)
			}
		})
	}
}

func TestDecisionLoadSaveAndIDs(t *testing.T) {
	root := t.TempDir()
	p, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := p.NextDecisionID(); got != "DEC-0001" {
		t.Errorf("first id = %s", got)
	}
	d := &Decision{ID: "DEC-0001", Title: "One", Decision: "Do it", Status: DecStatus{State: DecProposed, Date: "2026-01-01"}}
	if err := p.SaveDecision(d); err != nil {
		t.Fatal(err)
	}
	// An unreadable file still holds its number.
	os.WriteFile(filepath.Join(root, "ajiya", "decisions", "DEC-0004.md"), []byte("junk\n"), 0o644)
	p, err = Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Decisions) != 1 || p.BrokenDecisions["DEC-0004"] == nil || p.NextDecisionID() != "DEC-0005" {
		t.Errorf("decisions = %v, broken = %v, next = %s", p.Decisions, p.BrokenDecisions, p.NextDecisionID())
	}
	if p.Decision("DEC-0001") == nil || p.Decision("DEC-0002") != nil {
		t.Error("Decision lookup")
	}
}

func TestCheckDecisionText(t *testing.T) {
	if err := CheckDecisionText("--context", "a\nb"); err != nil {
		t.Error(err)
	}
	for _, bad := range []string{" a", "a\n", "a\n## Decision\nb", "a\rb"} {
		if CheckDecisionText("--context", bad) == nil {
			t.Errorf("%q was allowed", bad)
		}
	}
}
