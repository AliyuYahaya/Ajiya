package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

var chgDay = time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)

// chRun runs a command with the clock at chgDay plus the given days.
func chRun(t *testing.T, dir string, days int, args ...string) (string, int) {
	t.Helper()
	var out, errOut bytes.Buffer
	now := chgDay.AddDate(0, 0, days)
	e := &env{stdout: &out, stderr: &errOut, dir: dir, noAgents: true, now: func() time.Time { return now }}
	code := run(e, args)
	return out.String() + errOut.String(), code
}

func chMust(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, code := chRun(t, dir, 0, args...)
	if code != 0 {
		t.Fatalf("ajiya %v: exit %d\n%s", args, code, out)
	}
	return out
}

func chRefused(t *testing.T, dir string, wantCode int, wantText string, args ...string) {
	t.Helper()
	out, code := chRun(t, dir, 0, args...)
	if code != wantCode || !strings.Contains(out, wantText) {
		t.Errorf("ajiya %v: exit %d, want %d containing %q:\n%s", args, code, wantCode, wantText, out)
	}
}

func chCommitAll(t *testing.T, root, trailer string) {
	t.Helper()
	mrGit(t, root, "add", "ajiya", "ajiya.toml")
	mrGit(t, root, "commit", "-q", "--allow-empty", "-m", "Work", "-m", trailer)
}

// chPlan builds the Supabase example in one git repository:
//
//	MI-0001 done, tech:aws-rds                         found by tag
//	MI-0002 dec:DEC-0001                               found by decision
//	MI-0003 path:infra/rds/** (open, no commits)       found by path tag
//	MI-0004 no tags, commits changed infra/rds/b.tf    found by the files in its commits
//	MI-0005 no tags, depends on MI-0001                found as downstream only
//	MI-0006 no tags, depends on MI-0005                two levels away: not found
//	MI-0007 no tags, no links, commit changed docs/x.go not found
//
// Everything is committed, with the plan clean.
func chPlan(t *testing.T) string {
	t.Helper()
	root := tfPlan(t) // MI-0001 to MI-0005
	for _, title := range []string{"Sixth", "Seventh"} {
		chMust(t, root, "ticket", "add", "--phase", "build", "--app", "docs", title, "--done-when", "ok")
	}
	chMust(t, root, "decision", "add", "Use PostgreSQL on AWS RDS", "--decision", "RDS", "--tech", "aws-rds")
	chMust(t, root, "ticket", "edit", "MI-0001", "--tech", "aws-rds")
	chMust(t, root, "ticket", "edit", "MI-0002", "--decisions", "DEC-0001")
	chMust(t, root, "ticket", "edit", "MI-0003", "--paths", "infra/rds/**")
	chMust(t, root, "ticket", "edit", "MI-0005", "--depends", "MI-0001")
	chMust(t, root, "ticket", "edit", "MI-0006", "--depends", "MI-0005")
	chCommitAll(t, root, "Ajiya: chore")
	tfCommit(t, root, "Ajiya: MI-0004", "infra/rds/b.tf", "README.md")
	tfCommit(t, root, "Ajiya: MI-0007", "docs/x.go")
	return root
}

func chRead(t *testing.T, root, id string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "ajiya", "changes", id+".md"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// chVerdict gives a verdict the way 'change assess' will: with SetVerdict and SaveChange.
func chVerdict(t *testing.T, root, chg, ticket string, v plan.Verdict, why string) {
	t.Helper()
	pr, err := loadAt(root)
	if err != nil {
		t.Fatal(err)
	}
	c := pr.plan.Change(chg)
	if c == nil {
		t.Fatalf("no change %s", chg)
	}
	if err := c.SetVerdict(ticket, v, why, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := pr.plan.SaveChange(c); err != nil {
		t.Fatal(err)
	}
}

func TestChangeOpenRefusals(t *testing.T) {
	root := chPlan(t)
	chRefused(t, root, ExitUsage, "give at least one of --replaces, --decision or --paths", "change", "open", "Nothing")
	chRefused(t, root, ExitUsage, "--with", "change", "open", "T", "--replaces", "tech:aws-rds", "--with", "tech:Bad")
	chRefused(t, root, ExitUsage, "--replaces", "change", "open", "T", "--replaces", "colour:red")
	chRefused(t, root, ExitUsage, "--replaces", "change", "open", "T", "--replaces", "aws-rds")
	chRefused(t, root, ExitUsage, "--paths", "change", "open", "T", "--paths", "a b")
	chRefused(t, root, ExitUsage, "not a decision ID", "change", "open", "T", "--decision", "DEC-3")
	chRefused(t, root, ExitRefused, "no decision DEC-0009", "change", "open", "T", "--decision", "DEC-0009")
	chRefused(t, root, ExitUsage, "one line", "change", "open", "two\nlines", "--paths", "x/**")
	if _, err := os.Stat(filepath.Join(root, "ajiya", "changes")); err == nil {
		t.Error("a refused open left a changes folder behind")
	}
	if out := mrGitOut(t, root, "tag"); strings.Contains(out, "CHG") {
		t.Errorf("a refused open left a snapshot: %s", out)
	}

	// Uncommitted plan changes: the snapshot would not hold them.
	chMust(t, root, "ticket", "add", "--phase", "build", "--app", "docs", "Extra", "--done-when", "ok")
	chRefused(t, root, ExitRefused, "uncommitted", "change", "open", "T", "--replaces", "tech:aws-rds")
	if _, err := os.Stat(filepath.Join(root, "ajiya", "changes")); err == nil {
		t.Error("open with a dirty plan wrote a change file")
	}
}

func mrGitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func TestChangeOpenImpactScenario(t *testing.T) {
	root := chPlan(t)
	out := chMust(t, root, "change", "open", "Move the database from AWS RDS to Supabase",
		"--replaces", "tech:aws-rds", "--with", "tech:supabase", "--decision", "DEC-0001", "--paths", "infra/rds/**")
	for _, want := range []string{"Opened CHG-0001", "Snapshot before-CHG-0001", "uncommitted", "ajiya change impact CHG-0001"} {
		if !strings.Contains(out, want) {
			t.Errorf("open output lacks %q:\n%s", want, out)
		}
	}
	wantOpen := `<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# CHG-0001: Move the database from AWS RDS to Supabase

Status: Open · 2026-10-16
Replaces: tech:aws-rds
With: tech:supabase
Decisions: DEC-0001
Paths: infra/rds/**
Snapshot: before-CHG-0001

## Candidates
| ID | Ticket | Reasons | Verdict | Verdict reason | New title | New done when |
|---|---|---|---|---|---|---|

## Results
-
`
	if got := chRead(t, root, "CHG-0001"); got != wantOpen {
		t.Errorf("change file:\n%s\nwant:\n%s", got, wantOpen)
	}
	if tags := mrGitOut(t, root, "tag", "-l", "ajiya/plan/before-CHG-0001"); !strings.Contains(tags, "before-CHG-0001") {
		t.Errorf("snapshot tag missing: %q", tags)
	}
	if out, code := chRun(t, root, 0, "check"); code != 0 {
		t.Errorf("check after open: %d\n%s", code, out)
	}

	// Impact: tag, decision, path tag, path from commits (no tags), downstream (no tags).
	out = chMust(t, root, "change", "impact", "CHG-0001")
	if !strings.Contains(out, "5 candidate(s): 5 new, 0 kept, 0 stale, 0 dropped.") {
		t.Errorf("impact summary:\n%s", out)
	}
	wantRows := `| MI-0001 | Work on docs | tag:tech:aws-rds | - | - | - | - |
| MI-0002 | More | decision:DEC-0001 | - | - | - | - |
| MI-0003 | More | path:infra/rds/** | - | - | - | - |
| MI-0004 | More | path:infra/rds/b.tf | - | - | - | - |
| MI-0005 | More | downstream:MI-0001 | - | - | - | - |
`
	if got := chRead(t, root, "CHG-0001"); !strings.Contains(got, wantRows) {
		t.Errorf("candidate rows:\n%s\nwant:\n%s", got, wantRows)
	}
	for _, id := range []string{"MI-0006", "MI-0007"} {
		if strings.Contains(chRead(t, root, "CHG-0001"), "| "+id+" ") {
			t.Errorf("%s was picked but should not be", id)
		}
	}

	// Rerun with nothing changed: the file is byte for byte the same.
	before := chRead(t, root, "CHG-0001")
	out = chMust(t, root, "change", "impact", "CHG-0001")
	if !strings.Contains(out, "5 candidate(s): 0 new, 5 kept, 0 stale, 0 dropped.") || chRead(t, root, "CHG-0001") != before {
		t.Errorf("rerun changed the file or the summary:\n%s", out)
	}

	// Verdicts survive a rerun; a candidate with a verdict that stops matching is
	// kept and marked; one without a verdict is dropped.
	chVerdict(t, root, "CHG-0001", "MI-0001", plan.VerdictReplace, "Uses the RDS API | directly")
	chVerdict(t, root, "CHG-0001", "MI-0002", plan.VerdictUnaffected, "Only wording")
	chMust(t, root, "ticket", "edit", "MI-0002", "--decisions", "-")
	chMust(t, root, "ticket", "edit", "MI-0003", "--paths", "-")
	out = chMust(t, root, "change", "impact", "CHG-0001")
	if !strings.Contains(out, "4 candidate(s): 0 new, 3 kept, 1 stale, 1 dropped.") {
		t.Errorf("rerun summary:\n%s", out)
	}
	got := chRead(t, root, "CHG-0001")
	for _, want := range []string{
		`| MI-0001 | Work on docs | tag:tech:aws-rds | replace | Uses the RDS API \| directly | - | - |`,
		`| MI-0002 | More | decision:DEC-0001 stale | unaffected | Only wording | - | - |`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("after rerun lacks %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "| MI-0003 |") {
		t.Errorf("MI-0003 should have been dropped:\n%s", got)
	}
	if out, code := chRun(t, root, 0, "check"); code != 0 {
		t.Errorf("check after impact: %d\n%s", code, out)
	}

	// A second change needs the first one committed (its file is untracked).
	chRefused(t, root, ExitRefused, "uncommitted", "change", "open", "Another", "--paths", "docs/**")
	chCommitAll(t, root, "Ajiya: chore")
	out = chMust(t, root, "change", "open", "Another", "--paths", "docs/**")
	if !strings.Contains(out, "Opened CHG-0002") {
		t.Errorf("second open:\n%s", out)
	}
	// The no-tags ticket whose only link is a commit file is found by the second change too.
	out = chMust(t, root, "change", "impact", "CHG-0002")
	if !strings.Contains(out, "MI-0007") || !strings.Contains(out, "path:docs/x.go") {
		t.Errorf("CHG-0002 impact:\n%s", out)
	}

	// list and show
	out = chMust(t, root, "change", "list")
	if !strings.Contains(out, "2 change(s)") || !strings.Contains(out, "CHG-0001") || !strings.Contains(out, "Open") {
		t.Errorf("list:\n%s", out)
	}
	out = chMust(t, root, "change", "show", "CHG-0001")
	for _, want := range []string{"Status:     Open · 2026-10-16", "Replaces:   tech:aws-rds", "Decisions:  DEC-0001", "Paths:      infra/rds/**", "replace", "Uses the RDS API | directly"} {
		if !strings.Contains(out, want) {
			t.Errorf("show lacks %q:\n%s", want, out)
		}
	}
	var js struct {
		ID         string   `json:"id"`
		Decisions  []string `json:"decisions"`
		Candidates []struct {
			ID      string   `json:"id"`
			Reasons []string `json:"reasons"`
			Verdict string   `json:"verdict"`
			Stale   bool     `json:"stale"`
		} `json:"candidates"`
		Status struct {
			State string `json:"state"`
		} `json:"status"`
	}
	if err := json.Unmarshal([]byte(chMust(t, root, "change", "show", "CHG-0001", "--json")), &js); err != nil {
		t.Fatal(err)
	}
	if js.ID != "CHG-0001" || js.Status.State != "open" || len(js.Candidates) != 4 || js.Candidates[1].ID != "MI-0002" || !js.Candidates[1].Stale ||
		js.Candidates[0].Verdict != "replace" || strings.Join(js.Candidates[3].Reasons, ",") != "downstream:MI-0001" {
		t.Errorf("show --json = %+v", js)
	}
	var list []map[string]any
	if err := json.Unmarshal([]byte(chMust(t, root, "change", "list", "--json")), &list); err != nil || len(list) != 2 {
		t.Errorf("list --json: %v %v", err, list)
	}

	// ticket show lists the changes that touched the ticket.
	out = chMust(t, root, "ticket", "show", "MI-0001")
	if !strings.Contains(out, "Changes:    CHG-0001 open, replace: Move the database from AWS RDS to Supabase") {
		t.Errorf("ticket show:\n%s", out)
	}
	if out := chMust(t, root, "ticket", "show", "MI-0006"); strings.Contains(out, "Changes:") {
		t.Errorf("ticket show lists a change for an untouched ticket:\n%s", out)
	}
	var ts struct {
		Changes []changeLink `json:"changes"`
	}
	if err := json.Unmarshal([]byte(chMust(t, root, "ticket", "show", "MI-0001", "--json")), &ts); err != nil || len(ts.Changes) != 1 || ts.Changes[0].Verdict != "replace" {
		t.Errorf("ticket show --json: %v %+v", err, ts)
	}
	if err := json.Unmarshal([]byte(chMust(t, root, "ticket", "show", "MI-0006", "--json")), &ts); err != nil || ts.Changes == nil || len(ts.Changes) != 0 {
		t.Errorf("ticket show --json for an untouched ticket: %v %+v", err, ts)
	}

	// The dashboard data carries the changes.
	chMust(t, root, "build")
	data, err := os.ReadFile(filepath.Join(root, "ajiya", "data.js"))
	if err != nil || !strings.Contains(string(data), `"changes"`) || !strings.Contains(string(data), `"CHG-0001"`) || !strings.Contains(string(data), `"downstream:MI-0001"`) {
		t.Errorf("data.js lacks the changes: %v", err)
	}
}

func TestChangeImpactRefusals(t *testing.T) {
	root := chPlan(t)
	chMust(t, root, "change", "open", "Move", "--replaces", "tech:aws-rds")
	chRefused(t, root, ExitRefused, "no change CHG-0009", "change", "impact", "CHG-0009")
	chRefused(t, root, ExitRefused, "no change CHG-0009", "change", "show", "CHG-0009")
	chRefused(t, root, ExitUsage, "expected 1 argument", "change", "impact")

	// A closed change keeps its candidates: impact refuses.
	pr, _ := loadAt(root)
	c := pr.plan.Change("CHG-0001")
	c.Status = plan.ChangeStatus{State: plan.ChgRejected, Date: "2026-10-17", By: "Sam", Reason: "cancelled"}
	if err := pr.plan.SaveChange(c); err != nil {
		t.Fatal(err)
	}
	chRefused(t, root, ExitRefused, "CHG-0001 is rejected", "change", "impact", "CHG-0001")
	if err := os.WriteFile(filepath.Join(root, "ajiya", "changes", "CHG-0001.md"), []byte("garbage\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	chRefused(t, root, ExitRefused, "see 'ajiya check'", "change", "impact", "CHG-0001")
	chRefused(t, root, ExitRefused, "see 'ajiya check'", "change", "show", "CHG-0001")
}

func TestChangeChecks(t *testing.T) {
	root := chPlan(t)
	chMust(t, root, "change", "open", "Move", "--replaces", "tech:aws-rds")
	if out, code := chRun(t, root, 0, "check", "--strict"); code != 0 {
		t.Fatalf("check --strict on a fresh change: %d\n%s", code, out)
	}
	// 14 days after opening is fine; 15 is a warning (W015), which fails --strict only.
	if out, code := chRun(t, root, 14, "check", "--strict"); code != 0 {
		t.Errorf("14 days: %d\n%s", code, out)
	}
	out, code := chRun(t, root, 15, "check")
	if code != 0 || !strings.Contains(out, "W015 ajiya/changes/CHG-0001.md:1: CHG-0001 has been open for 15 days") {
		t.Errorf("15 days: %d\n%s", code, out)
	}
	if _, code := chRun(t, root, 15, "check", "--strict"); code == 0 {
		t.Error("--strict passed with W015")
	}

	// A hand edit that still parses is E019; one that does not parse is E018.
	p := filepath.Join(root, "ajiya", "changes", "CHG-0001.md")
	good := chRead(t, root, "CHG-0001")
	if err := os.WriteFile(p, []byte(strings.Replace(good, "## Results", "\n## Results", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, code := chRun(t, root, 0, "check"); code != ExitRefused || !strings.Contains(out, "E019 ajiya/changes/CHG-0001.md:") {
		t.Errorf("hand edit: %d\n%s", code, out)
	}
	if err := os.WriteFile(p, []byte(strings.Replace(good, "Status: Open", "Status: Maybe", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, code := chRun(t, root, 0, "check"); code != ExitRefused || !strings.Contains(out, "E018 ajiya/changes/CHG-0001.md:4") {
		t.Errorf("broken file: %d\n%s", code, out)
	}
}
