package cli

import (
	"encoding/json"
	"github.com/AliyuYahaya/Ajiya/internal/config"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// chPlanFiles reads every file under ajiya/ except the change files, keyed by path.
func chPlanFiles(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(filepath.Join(root, "ajiya"), func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		data, err := os.ReadFile(p)
		out[filepath.ToSlash(rel)] = string(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := os.ReadFile(filepath.Join(root, "ajiya.toml"))
	if err != nil {
		t.Fatal(err)
	}
	out["ajiya.toml"] = string(cfg)
	return out
}

type chTotals struct {
	done, total, required, closed, percent int
}

func chCounts(t *testing.T, root string) chTotals {
	t.Helper()
	var ph []struct {
		Slug        string
		Done, Total int
	}
	if err := json.Unmarshal([]byte(chMust(t, root, "phase", "list", "--json")), &ph); err != nil || len(ph) != 1 {
		t.Fatalf("phase list --json: %v %v", err, ph)
	}
	var ms []struct {
		Name                      string
		Required, Closed, Percent int
	}
	if err := json.Unmarshal([]byte(chMust(t, root, "milestone", "list", "--json")), &ms); err != nil || len(ms) < 1 || ms[0].Name != "staging" {
		t.Fatalf("milestone list --json: %v %v", err, ms)
	}
	return chTotals{ph[0].Done, ph[0].Total, ms[0].Required, ms[0].Closed, ms[0].Percent}
}

// chSupabase is the Supabase example after chPlan, with evidence:
//
//	MI-0001 done (commit), tech:aws-rds
//	MI-0002 open, dec:DEC-0001
//	MI-0003 open, path:infra/rds/**, depends on MI-0007
//	MI-0004 done (commit changed infra/rds/b.tf), no tags
//	MI-0005 open, no tags, depends on MI-0001
//	MI-0006 open, no tags, depends on MI-0005 and MI-0003
//	MI-0007 done, unrelated
//
// Milestones: staging targets the build phase, v1 targets MI-0001 and MI-0007,
// qa targets MI-0003 and MI-0007. Everything is committed.
func chSupabase(t *testing.T) string {
	t.Helper()
	root := chPlan(t)
	chMust(t, root, "decision", "accept", "DEC-0001", "--by", "Aliyu")
	tfCommit(t, root, "Ajiya: MI-0001", "docs/rds.go")
	for _, id := range []string{"MI-0001", "MI-0004", "MI-0007"} {
		chMust(t, root, "ticket", "done", id)
	}
	chMust(t, root, "ticket", "edit", "MI-0003", "--depends", "MI-0007")
	chMust(t, root, "ticket", "edit", "MI-0006", "--depends", "MI-0003,MI-0005")
	chMust(t, root, "milestone", "add", "staging", "--targets", "build")
	chMust(t, root, "milestone", "add", "v1", "--targets", "MI-0001,MI-0007", "--after", "staging")
	chMust(t, root, "milestone", "add", "qa", "--targets", "MI-0003,MI-0007", "--after", "v1")
	chCommitAll(t, root, "Ajiya: chore")
	return root
}

func TestChangeApproveSupabaseScenario(t *testing.T) {
	root := chSupabase(t)
	phaseFile := func() string { return chPlanFiles(t, root)["ajiya/build.md"] }
	row := func(file, id string) string {
		for _, l := range strings.Split(file, "\n") {
			if strings.HasPrefix(l, "| "+id+" |") {
				return l
			}
		}
		t.Fatalf("no row %s in\n%s", id, file)
		return ""
	}
	oldFile := phaseFile()
	evidence := strings.TrimSpace(regexp.MustCompile(`Done · [0-9a-f]{7,}[^|]*`).FindString(row(oldFile, "MI-0001")))
	if evidence == "" {
		t.Fatalf("MI-0001 has no evidence:\n%s", row(oldFile, "MI-0001"))
	}
	if got := chCounts(t, root); got != (chTotals{3, 7, 7, 3, 42}) {
		t.Fatalf("totals before = %+v", got)
	}

	// Open the change, find the candidates.
	chMust(t, root, "change", "open", "Move the database from AWS RDS to Supabase",
		"--replaces", "tech:aws-rds", "--with", "tech:supabase", "--decision", "DEC-0001", "--paths", "infra/rds/**")
	out := chMust(t, root, "change", "impact", "CHG-0001")
	if !strings.Contains(out, "6 candidate(s)") { // 0001 tag, 0002 decision, 0003 path tag, 0004 commit path, 0005 downstream of 0001, 0006 downstream of 0003
		t.Fatalf("impact:\n%s", out)
	}
	asm := func(id, verdict, reason string, extra ...string) {
		t.Helper()
		chMust(t, root, append([]string{"change", "assess", "CHG-0001", id, "--verdict", verdict, "--reason", reason}, extra...)...)
	}
	asm("MI-0001", "replace", "Uses the RDS backup API", "--title", "Configure Supabase backups", "--done-when", "Daily backup, 7-day retention")
	asm("MI-0002", "modify", "Wording follows the new choice", "--title", "Record the Supabase choice")
	asm("MI-0003", "invalidate", "Only RDS needs a path rule")
	asm("MI-0004", "modify", "Schema script must target Supabase", "--title", "Port the schema script to Supabase", "--done-when", "Script runs on Supabase")
	asm("MI-0005", "investigate", "Not sure it relies on RDS")
	asm("MI-0006", "unaffected", "Generic")

	// Refused: a verdict is investigate; the plan is not touched.
	chCommitAll(t, root, "Ajiya: chore")
	filesBefore := chPlanFiles(t, root)
	chRefused(t, root, ExitRefused, "still marked investigate: MI-0005", "change", "approve", "CHG-0001", "--by", "Aliyu")
	if !mapsEqual(filesBefore, chPlanFiles(t, root)) {
		t.Error("a refused approve changed files")
	}
	// Re-assess, then the commit rule: ajiya/ must be clean.
	asm("MI-0005", "unaffected", "It only reads the connection string")
	chRefused(t, root, ExitRefused, "uncommitted changes", "change", "approve", "CHG-0001", "--by", "Aliyu")
	chCommitAll(t, root, "Ajiya: chore")

	// Other refusals: no --by, a decision that does not exist.
	chRefused(t, root, ExitUsage, "--by is required", "change", "approve", "CHG-0001")
	chRefused(t, root, ExitRefused, "--new-decision DEC-0009 does not exist", "change", "approve", "CHG-0001", "--by", "Aliyu", "--new-decision", "DEC-0009")
	if phaseFile() != oldFile {
		t.Error("a refused approve changed the plan")
	}

	// A new decision to replace DEC-0001, then approve.
	chMust(t, root, "decision", "add", "Use Supabase", "--decision", "Postgres on Supabase", "--tech", "supabase", "--by", "Aliyu")
	chCommitAll(t, root, "Ajiya: chore")
	out = chMust(t, root, "change", "approve", "CHG-0001", "--by", "Aliyu", "--new-decision", "DEC-0002")
	if !strings.Contains(out, "Approved CHG-0001 by Aliyu") {
		t.Errorf("approve output:\n%s", out)
	}

	// ajiya.toml was retargeted in the same write: the replaced target follows
	// its replacement, the invalidated one is swapped for its dependencies
	// (MI-0007 was already a target of qa, so it is not listed twice), and the
	// phase target is untouched.
	cfgText, err := config.ReadText(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`targets = ["build"]`, `targets = ["MI-0008", "MI-0007"]`, `targets = ["MI-0007"]`} {
		if !strings.Contains(cfgText, want) {
			t.Errorf("ajiya.toml lacks %s:\n%s", want, cfgText)
		}
	}
	if strings.Contains(cfgText, "MI-0001") || strings.Contains(cfgText, "MI-0003") {
		t.Errorf("ajiya.toml still targets a superseded ticket:\n%s", cfgText)
	}

	file := phaseFile()
	// New tickets: the replacement (MI-0008) and the follow-up (MI-0009), same phase and app.
	r8 := row(file, "MI-0008")
	for _, want := range []string{"| docs |", "Configure Supabase backups", "Daily backup, 7-day retention", "| - |", "tech:supabase", "🟥 Pending"} {
		if !strings.Contains(r8, want) {
			t.Errorf("MI-0008 lacks %q: %s", want, r8)
		}
	}
	if strings.Contains(r8, "aws-rds") {
		t.Errorf("MI-0008 kept the replaced tag: %s", r8)
	}
	r9 := row(file, "MI-0009")
	for _, want := range []string{"Port the schema script to Supabase", "Script runs on Supabase", "| MI-0004 |", "🟥 Pending"} {
		if !strings.Contains(r9, want) {
			t.Errorf("MI-0009 lacks %q: %s", want, r9)
		}
	}
	// Superseded tickets keep their evidence.
	if r := row(file, "MI-0001"); !strings.Contains(r, "🟩 Superseded by MI-0008 · CHG-0001 · was: "+evidence) {
		t.Errorf("MI-0001 = %s\nwant evidence %q kept", r, evidence)
	}
	if r := row(file, "MI-0003"); !strings.Contains(r, "🟩 Superseded · CHG-0001 · was: 🟥 Pending") && !strings.Contains(r, "🟩 Superseded · CHG-0001 · was: Pending") {
		t.Errorf("MI-0003 = %s", r)
	}
	// Edited in place: only the title.
	if r := row(file, "MI-0002"); !strings.Contains(r, "Record the Supabase choice") || !strings.Contains(r, "dec:DEC-0002") || strings.Contains(r, "DEC-0001") || !strings.Contains(r, "🟥 Pending") {
		t.Errorf("MI-0002 = %s", r)
	}
	// Left alone: the done follow-up's origin and the unaffected ones.
	if row(file, "MI-0004") != row(oldFile, "MI-0004") || row(file, "MI-0007") != row(oldFile, "MI-0007") {
		t.Error("a done ticket that was not superseded changed")
	}
	// Dependencies: MI-0005 moved from the replaced MI-0001 to MI-0008; MI-0006 moved off
	// the invalidated MI-0003 to what MI-0003 waited on (MI-0007).
	if r := row(file, "MI-0005"); !strings.Contains(r, "| MI-0008 |") || strings.Contains(r, "MI-0001 |") {
		t.Errorf("MI-0005 = %s", r)
	}
	if r := row(file, "MI-0006"); !strings.Contains(r, "| MI-0005, MI-0007 |") {
		t.Errorf("MI-0006 = %s", r)
	}

	// The superseded decision, and the new one untouched.
	dec := chPlanFiles(t, root)["ajiya/decisions/DEC-0001.md"]
	if !strings.Contains(dec, "Status: Superseded by DEC-0002 (CHG-0001) · was: Accepted") && !strings.Contains(dec, "Status: Superseded by DEC-0002 (CHG-0001)") {
		t.Errorf("DEC-0001:\n%s", dec)
	}

	// Totals leave out superseded work: 7 live (0002, 0004..0009), 2 done.
	if got := chCounts(t, root); got != (chTotals{2, 7, 7, 2, 28}) {
		t.Errorf("totals after = %+v, want 2 of 7 and 28%%", got)
	}
	// Nothing for the closed work to do: next offers the new tickets, not the superseded ones.
	next := chMust(t, root, "next")
	if strings.Contains(next, "MI-0001") || strings.Contains(next, "MI-0003") || !strings.Contains(next, "MI-0008") {
		t.Errorf("next:\n%s", next)
	}

	// The change file records who, when and what.
	chg := chRead(t, root, "CHG-0001")
	for _, want := range []string{
		"Status: Approved · 2026-10-16 · by Aliyu",
		"- Created MI-0008: Configure Supabase backups (replaces MI-0001)",
		"- Created MI-0009: Port the schema script to Supabase (follow-up to MI-0004)",
		"- Edited MI-0002: Record the Supabase choice (title, decision tag)",
		"- Superseded MI-0001 by MI-0008 (was: " + evidence,
		"- Superseded MI-0003 with no replacement (was: Pending)",
		"- Moved dependencies of MI-0005: MI-0001 -> MI-0008",
		"- Moved dependencies of MI-0006: MI-0003, MI-0005 -> MI-0005, MI-0007",
		"- Superseded decision DEC-0001 by DEC-0002",
		"- milestone v1: target MI-0001 → MI-0008",
		"- milestone qa: target MI-0003 → MI-0007 (invalidated; its dependencies)",
	} {
		if !strings.Contains(chg, want) {
			t.Errorf("change file lacks %q:\n%s", want, chg)
		}
	}
	if out, code := chRun(t, root, 0, "check"); code != 0 {
		t.Errorf("check after approve: %d\n%s", code, out)
	}
	// No open ticket is left pointing at the superseded decision (W014).
	if out, _ := chRun(t, root, 0, "check"); strings.Contains(out, "W014") {
		t.Errorf("check warns about a superseded decision:\n%s", out)
	}
	// Ticket show reaches the whole story from either end.
	if out := chMust(t, root, "ticket", "show", "MI-0001"); !strings.Contains(out, "Superseded by MI-0008") || !strings.Contains(out, "CHG-0001 approved, replace") {
		t.Errorf("ticket show MI-0001:\n%s", out)
	}

	// Closed: no second approve, reject, or assess.
	chCommitAll(t, root, "Ajiya: chore")
	chRefused(t, root, ExitRefused, "approved", "change", "approve", "CHG-0001", "--by", "Aliyu")
	chRefused(t, root, ExitRefused, "approved", "change", "reject", "CHG-0001", "--by", "Aliyu", "--reason", "late")
	chRefused(t, root, ExitRefused, "approved", "change", "assess", "CHG-0001", "MI-0002", "--verdict", "unaffected", "--reason", "late")
	if out := chMust(t, root, "change", "list"); !strings.Contains(out, "Approved") {
		t.Errorf("list:\n%s", out)
	}
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func TestChangeRejectLeavesThePlanAlone(t *testing.T) {
	root := chSupabase(t)
	chMust(t, root, "change", "open", "Move the database", "--replaces", "tech:aws-rds", "--decision", "DEC-0001")
	chMust(t, root, "change", "impact", "CHG-0001")
	chMust(t, root, "change", "assess", "CHG-0001", "MI-0001", "--verdict", "invalidate", "--reason", "x")
	before := chPlanFiles(t, root)
	delete(before, "ajiya/changes/CHG-0001.md")

	chRefused(t, root, ExitUsage, "--by and --reason are both required", "change", "reject", "CHG-0001", "--by", "Aliyu")
	chRefused(t, root, ExitUsage, "--by and --reason are both required", "change", "reject", "CHG-0001", "--reason", "x")
	chRefused(t, root, ExitUsage, "'·'", "change", "reject", "CHG-0001", "--by", "Aliyu", "--reason", "a · b")
	chRefused(t, root, ExitRefused, "no change CHG-0009", "change", "reject", "CHG-0009", "--by", "Aliyu", "--reason", "x")
	chMust(t, root, "change", "reject", "CHG-0001", "--by", "Aliyu", "--reason", "The move was cancelled")

	after := chPlanFiles(t, root)
	chg := after["ajiya/changes/CHG-0001.md"]
	delete(after, "ajiya/changes/CHG-0001.md")
	if !mapsEqual(before, after) {
		t.Errorf("reject changed plan files besides the change file:\nbefore %v\nafter %v", keys(before), keys(after))
	}
	if !strings.Contains(chg, "Status: Rejected · 2026-10-16 · by Aliyu · The move was cancelled") {
		t.Errorf("change file:\n%s", chg)
	}
	if !strings.Contains(chg, "| MI-0001 |") || !strings.Contains(chg, "invalidate") {
		t.Errorf("reject dropped the candidates and verdicts:\n%s", chg)
	}
	if out, code := chRun(t, root, 0, "check"); code != 0 {
		t.Errorf("check after reject: %d\n%s", code, out)
	}
	// No snapshot or ticket was removed either: the tag stays.
	if tags := mrGitOut(t, root, "tag", "-l", "ajiya/plan/before-CHG-0001"); !strings.Contains(tags, "before-CHG-0001") {
		t.Error("the snapshot was removed")
	}
	chRefused(t, root, ExitRefused, "rejected", "change", "approve", "CHG-0001", "--by", "Aliyu")
}

func keys(m map[string]string) []string {
	var k []string
	for x := range m {
		k = append(k, x)
	}
	return k
}

// An approve that would leave errors in the plan writes nothing.
func TestChangeApproveWritesNothingWhenTheResultIsInvalid(t *testing.T) {
	root := chSupabase(t)
	// The new ticket would belong to an app that is not registered (E010).
	chMust(t, root, "change", "open", "Replace work", "--replaces", "tech:aws-rds")
	chMust(t, root, "change", "impact", "CHG-0001")
	chMust(t, root, "change", "assess", "CHG-0001", "MI-0001", "--verdict", "replace", "--reason", "x", "--title", "New", "--done-when", "ok")
	chMust(t, root, "change", "assess", "CHG-0001", "MI-0005", "--verdict", "unaffected", "--reason", "x")
	toml := filepath.Join(root, "ajiya.toml")
	cfg, _ := os.ReadFile(toml)
	if err := os.WriteFile(toml, []byte(strings.Replace(string(cfg), `name = "docs"`, `name = "docs2"`, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	chCommitAll(t, root, "Ajiya: chore")
	before := chPlanFiles(t, root)
	chRefused(t, root, ExitRefused, "would leave the plan with errors, so nothing was written", "change", "approve", "CHG-0001", "--by", "Aliyu")
	if !mapsEqual(before, chPlanFiles(t, root)) {
		t.Error("approve wrote files although the result was invalid")
	}
}
