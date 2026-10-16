package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Reading a canonical change file and writing it back gives the same bytes.
func TestChangeGoldenRoundTrip(t *testing.T) {
	files, err := filepath.Glob("testdata/golden-changes/CHG-*.md")
	if err != nil || len(files) < 4 {
		t.Fatalf("golden files: %v %v", files, err)
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			want, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			c, err := ParseChange(strings.TrimSuffix(filepath.Base(f), ".md"), want)
			if err != nil {
				t.Fatal(err)
			}
			if got := c.Format(); string(got) != string(want) {
				t.Errorf("round trip changed the file:\n--- got\n%s\n--- want\n%s", got, want)
			}
		})
	}
}

func TestParseChangeFields(t *testing.T) {
	data, _ := os.ReadFile("testdata/golden-changes/CHG-0002.md")
	c, err := ParseChange("CHG-0002", data)
	if err != nil {
		t.Fatal(err)
	}
	if c.Title != "Replace the RDS backup job" || !c.Status.Open() || c.Status.Date != "2026-10-17" || c.Snapshot != "before-CHG-0002" {
		t.Errorf("change = %+v", c)
	}
	if strings.Join(c.Decisions, ",") != "DEC-0003,DEC-0004" || strings.Join(c.Paths, ",") != "infra/rds/**,services/*/db.go" ||
		c.Replaces.String() != "tech:aws-rds dec:DEC-0003" || c.With.String() != "tech:supabase" {
		t.Errorf("header = %+v", c)
	}
	if len(c.Candidates) != 4 {
		t.Fatalf("candidates = %+v", c.Candidates)
	}
	k := c.Candidates[0]
	if k.Title != "Provision RDS | size the instance" || k.Verdict != VerdictReplace || k.NewDoneWhen != "Project exists, | connection string stored" || len(k.Reasons) != 3 {
		t.Errorf("candidate 0 = %+v", k)
	}
	if k := c.Candidates[1]; k.Verdict != "" || k.Why != "" || k.Stale {
		t.Errorf("candidate 1 = %+v", k)
	}
	if !c.Candidates[3].Stale || len(c.Candidates[3].Reasons) != 1 {
		t.Errorf("candidate 3 = %+v", c.Candidates[3])
	}
	data, _ = os.ReadFile("testdata/golden-changes/CHG-0004.md")
	c, _ = ParseChange("CHG-0004", data)
	if c.Status.State != ChgRejected || c.Status.By != "Aliyu" || c.Status.Reason != "The team decided to keep two repositories" {
		t.Errorf("rejected status = %+v", c.Status)
	}
}

func TestParseChangeErrors(t *testing.T) {
	good, _ := os.ReadFile("testdata/golden-changes/CHG-0002.md")
	for name, edit := range map[string][2]string{
		"wrong id":           {"# CHG-0002:", "# CHG-0009:"},
		"bad status":         {"Open · 2026-10-17", "Maybe · 2026-10-17"},
		"approved no name":   {"Open · 2026-10-17", "Approved · 2026-10-17"},
		"bad tag":            {"tech:aws-rds dec", "tech:AWS dec"},
		"bad decision":       {"DEC-0004\n", "DEC-4\n"},
		"bad verdict":        {"| replace |", "| rework |"},
		"bad reason":         {"tag:tech:aws-rds decision", "because:tech:aws-rds decision"},
		"short row":          {"| - | - | - |\n| AJ-0107", "| - |\n| AJ-0107"},
		"missing results":    {"## Results\n-\n", ""},
		"extra section":      {"## Results\n-\n", "## Results\n-\n## More\nx\n"},
		"twice":              {"| AJ-0105 |", "| AJ-0101 |"},
		"no candidates head": {"## Candidates", "## Candidate"},
	} {
		if _, err := ParseChange("CHG-0002", []byte(strings.Replace(string(good), edit[0], edit[1], 1))); err == nil {
			t.Errorf("%s: no error", name)
		} else if _, ok := err.(*ParseError); !ok {
			t.Errorf("%s: %T, want *ParseError", name, err)
		}
	}
}

func TestGlobMatch(t *testing.T) {
	for _, c := range []struct {
		glob, file string
		want       bool
	}{
		{"infra/rds/**", "infra/rds/main.tf", true},
		{"infra/rds/**", "infra/rds/deep/er/main.tf", true},
		{"infra/rds/**", "infra/rds", true},
		{"infra/**/main.tf", "infra/main.tf", true},
		{"infra/**/main.tf", "infra/a/b/main.tf", true},
		{"infra/**/main.tf", "infra/a/b/other.tf", false},
		{"infra/rds", "infra/rds/main.tf", true},
		{"infra/*", "infra/rds/main.tf", true},
		{"**/*.sql", "db/migrations/001.sql", true},
		{"**/*.sql", "db/migrations/001.go", false},
		{"services/*/db.go", "services/api/db.go", true},
		{"services/*/db.go", "services/api/x/db.go", false},
		{"infra/rds/**", "infra/s3/main.tf", false},
		{"infra/rd", "infra/rds/main.tf", false},
	} {
		if got := GlobMatch(c.glob, c.file); got != c.want {
			t.Errorf("GlobMatch(%q, %q) = %v", c.glob, c.file, got)
		}
	}
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"infra/rds/**", "infra/rds/**", true},
		{"infra/rds/**", "infra/rds/backup.tf", true},
		{"infra/**", "infra/rds/**", true},
		{"infra/rds/**", "infra/s3/**", false},
		{"infra/rds/**", "web/**", false},
	} {
		if got := GlobsOverlap(c.a, c.b); got != c.want {
			t.Errorf("GlobsOverlap(%q, %q) = %v", c.a, c.b, got)
		}
	}
}

func impactPlan() *Plan {
	tk := func(id string, deps []string, tags string, st Status) *Ticket {
		tg, err := ParseTags(tags)
		if err != nil {
			panic(err)
		}
		return &Ticket{ID: id, Title: "T " + id, Depends: deps, Tags: tg, Status: st}
	}
	return &Plan{Phases: []*Phase{{Slug: "core", Tickets: []*Ticket{
		tk("AJ-0001", nil, "tech:aws-rds", Status{State: Done}),
		tk("AJ-0002", nil, "dec:DEC-0003", Status{State: Pending}),
		tk("AJ-0003", nil, "path:infra/rds/**", Status{State: Pending}),
		tk("AJ-0004", nil, "-", Status{State: Done}),                    // no tags: found through commit files
		tk("AJ-0005", []string{"AJ-0001"}, "-", Status{State: Pending}), // no tags: downstream
		tk("AJ-0006", []string{"AJ-0005"}, "-", Status{State: Pending}), // two levels away: not found
		tk("AJ-0007", nil, "tech:aws-rds", Status{State: Superseded}),   // excluded
		tk("AJ-0008", nil, "tech:other path:web/**", Status{State: Pending}),
	}}}}
}

func TestFindImpactAndMerge(t *testing.T) {
	p := impactPlan()
	c := &Change{ID: "CHG-0001", Title: "t", Status: ChangeStatus{Date: "2026-10-16"}, Snapshot: "before-CHG-0001"}
	c.Replaces, _ = ParseTags("tech:aws-rds")
	c.Decisions = []string{"DEC-0003"}
	c.Paths = []string{"infra/rds/**"}
	files := map[string][]string{"AJ-0004": {"README.md", "infra/rds/main.tf", "infra/rds/z.tf"}, "AJ-0006": {"web/a.go"}}
	found := p.FindImpact(c, files)
	var got []string
	for _, f := range found {
		got = append(got, f.ID+"="+strings.Join(f.Reasons, ","))
	}
	want := []string{
		"AJ-0001=tag:tech:aws-rds",
		"AJ-0002=decision:DEC-0003",
		"AJ-0003=path:infra/rds/**",
		"AJ-0004=path:infra/rds/main.tf",
		"AJ-0005=downstream:AJ-0001",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("found:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}

	res := c.MergeCandidates(p, found)
	if len(res.Added) != 5 || len(c.Candidates) != 5 || c.Candidates[0].Title != "T AJ-0001" {
		t.Fatalf("merge = %+v %+v", res, c.Candidates)
	}
	if err := c.SetVerdict("AJ-0001", VerdictReplace, "Uses RDS", "New", "Done"); err != nil {
		t.Fatal(err)
	}
	if err := c.SetVerdict("AJ-0004", VerdictUnaffected, "Only a readme", "", ""); err != nil {
		t.Fatal(err)
	}
	// AJ-0003 loses its path tag, AJ-0004 stops matching: AJ-0004 has a verdict so it is kept, stale.
	p.Phases[0].Tickets[2].Tags = Tags{}
	found = p.FindImpact(c, nil)
	res = c.MergeCandidates(p, found)
	if strings.Join(res.Dropped, ",") != "AJ-0003" || strings.Join(res.Stale, ",") != "AJ-0004" || len(res.Added) != 0 {
		t.Errorf("rerun = %+v", res)
	}
	if k := c.Candidate("AJ-0001"); k.Verdict != VerdictReplace || k.Why != "Uses RDS" || k.NewTitle != "New" {
		t.Errorf("verdict lost: %+v", k)
	}
	if k := c.Candidate("AJ-0004"); !k.Stale || k.Verdict != VerdictUnaffected || len(k.Reasons) != 1 {
		t.Errorf("stale row = %+v", k)
	}
	if c.Candidate("AJ-0003") != nil {
		t.Error("candidate without a verdict that no longer qualifies was kept")
	}
	// What was written reads back the same.
	back, err := ParseChange(c.ID, c.Format())
	if err != nil || string(back.Format()) != string(c.Format()) {
		t.Errorf("round trip: %v", err)
	}
}

func TestSetVerdictRefusals(t *testing.T) {
	c := &Change{ID: "CHG-0001", Candidates: []Candidate{{ID: "AJ-0001", Title: "t"}}}
	for name, err := range map[string]error{
		"unknown ticket": c.SetVerdict("AJ-0009", VerdictModify, "x", "", ""),
		"bad verdict":    c.SetVerdict("AJ-0001", Verdict("rework"), "x", "", ""),
		"no reason":      c.SetVerdict("AJ-0001", VerdictModify, "", "", ""),
		"multi-line":     c.SetVerdict("AJ-0001", VerdictModify, "a\nb", "", ""),
	} {
		if err == nil {
			t.Errorf("%s: allowed", name)
		}
	}
	c.Status.State = ChgApproved
	if err := c.SetVerdict("AJ-0001", VerdictModify, "x", "", ""); err == nil {
		t.Error("verdict on an approved change was allowed")
	}
}
