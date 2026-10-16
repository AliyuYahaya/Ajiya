package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// chAssessPlan is chPlan with MI-0001 done (committed) and CHG-0001 open with
// its impact found: candidates MI-0001 (done), MI-0002, MI-0003, MI-0004, MI-0005.
func chAssessPlan(t *testing.T) string {
	t.Helper()
	root := chPlan(t)
	tfCommit(t, root, "Ajiya: MI-0001", "docs/rds.go")
	chMust(t, root, "ticket", "done", "MI-0001")
	chCommitAll(t, root, "Ajiya: MI-0001")
	chMust(t, root, "change", "open", "Move to Supabase", "--replaces", "tech:aws-rds", "--with", "tech:supabase",
		"--decision", "DEC-0001", "--paths", "infra/rds/**")
	chMust(t, root, "change", "impact", "CHG-0001")
	return root
}

func TestChangeAssessRefusals(t *testing.T) {
	root := chAssessPlan(t)
	a := func(args ...string) []string { return append([]string{"change", "assess", "CHG-0001"}, args...) }
	before := chRead(t, root, "CHG-0001")

	chRefused(t, root, ExitUsage, "--verdict is required", a("MI-0002", "--reason", "x")...)
	chRefused(t, root, ExitUsage, "not one of", a("MI-0002", "--verdict", "maybe", "--reason", "x")...)
	chRefused(t, root, ExitUsage, "--reason is required", a("MI-0002", "--verdict", "unaffected")...)
	chRefused(t, root, ExitUsage, "--reason is required", a("MI-0002", "--verdict", "unaffected", "--reason", "")...)
	chRefused(t, root, ExitUsage, "one line", a("MI-0002", "--verdict", "unaffected", "--reason", "a\nb")...)
	// Only replace and modify take the new text.
	for _, v := range []string{"unaffected", "invalidate", "investigate"} {
		chRefused(t, root, ExitUsage, "--title and --done-when are for replace and modify", a("MI-0002", "--verdict", v, "--reason", "x", "--title", "T")...)
		chRefused(t, root, ExitUsage, "--title and --done-when are for replace and modify", a("MI-0002", "--verdict", v, "--reason", "x", "--done-when", "D")...)
	}
	chRefused(t, root, ExitUsage, "replace needs --title and --done-when", a("MI-0002", "--verdict", "replace", "--reason", "x", "--title", "T")...)
	chRefused(t, root, ExitUsage, "replace needs --title and --done-when", a("MI-0002", "--verdict", "replace", "--reason", "x")...)
	// MI-0001 is done: modify adds a follow-up, so it needs both.
	chRefused(t, root, ExitUsage, "follow-up", a("MI-0001", "--verdict", "modify", "--reason", "x", "--title", "T")...)
	chRefused(t, root, ExitUsage, "follow-up", a("MI-0001", "--verdict", "modify", "--reason", "x")...)
	// MI-0002 is open: modify edits it, so it needs at least one.
	chRefused(t, root, ExitUsage, "edits it in place", a("MI-0002", "--verdict", "modify", "--reason", "x")...)
	chRefused(t, root, ExitUsage, "--title cannot be empty", a("MI-0002", "--verdict", "modify", "--reason", "x", "--title", "")...)
	// Not a candidate, no such change, no such ticket.
	chRefused(t, root, ExitRefused, "MI-0006 is not a candidate of CHG-0001", a("MI-0006", "--verdict", "unaffected", "--reason", "x")...)
	chRefused(t, root, ExitRefused, "no change CHG-0009", "change", "assess", "CHG-0009", "MI-0002", "--verdict", "unaffected", "--reason", "x")
	if got := chRead(t, root, "CHG-0001"); got != before {
		t.Errorf("a refused assess changed the change file:\n%s", got)
	}

	// A closed change takes no more verdicts.
	pr, err := loadAt(root)
	if err != nil {
		t.Fatal(err)
	}
	c := pr.plan.Change("CHG-0001")
	c.Status = plan.ChangeStatus{State: plan.ChgRejected, Date: "2026-10-17", By: "Aliyu", Reason: "cancelled"}
	if err := pr.plan.SaveChange(c); err != nil {
		t.Fatal(err)
	}
	chRefused(t, root, ExitRefused, "rejected", a("MI-0002", "--verdict", "unaffected", "--reason", "x")...)
}

func TestChangeAssessRecords(t *testing.T) {
	root := chAssessPlan(t)
	a := func(args ...string) string {
		return chMust(t, root, append([]string{"change", "assess", "CHG-0001"}, args...)...)
	}
	out := a("MI-0001", "--verdict", "replace", "--reason", "Uses the RDS API | directly", "--title", "Configure Supabase backups", "--done-when", "Daily backup | 7 days")
	if !strings.Contains(out, "MI-0001 is replace") || !strings.Contains(out, "4 candidate(s) still need a verdict") {
		t.Errorf("assess output:\n%s", out)
	}
	a("MI-0002", "--verdict", "modify", "--reason", "Wording", "--title", "Record the Supabase choice")
	a("MI-0003", "--verdict", "invalidate", "--reason", "RDS only")
	a("MI-0004", "--verdict", "investigate", "--reason", "Unclear")
	out = a("MI-0005", "--verdict", "unaffected", "--reason", "Generic")
	if !strings.Contains(out, "1 candidate(s) still need a verdict other than investigate") {
		t.Errorf("assess output:\n%s", out)
	}
	got := chRead(t, root, "CHG-0001")
	for _, want := range []string{
		`| MI-0001 | Work on docs | tag:tech:aws-rds | replace | Uses the RDS API \| directly | Configure Supabase backups | Daily backup \| 7 days |`,
		`| MI-0002 | More | decision:DEC-0001 | modify | Wording | Record the Supabase choice | - |`,
		`| MI-0003 | More | path:infra/rds/** | invalidate | RDS only | - | - |`,
		`| MI-0004 | More | path:infra/rds/b.tf | investigate | Unclear | - | - |`,
		`| MI-0005 | More | downstream:MI-0001 | unaffected | Generic | - | - |`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("change file lacks %q:\n%s", want, got)
		}
	}
	if out, code := chRun(t, root, 0, "check"); code != 0 {
		t.Errorf("check after assess: %d\n%s", code, out)
	}

	// Assessing again overwrites the verdict, and drops new text a new verdict does not take.
	out = a("MI-0001", "--verdict", "unaffected", "--reason", "Second thoughts")
	got = chRead(t, root, "CHG-0001")
	if !strings.Contains(got, `| MI-0001 | Work on docs | tag:tech:aws-rds | unaffected | Second thoughts | - | - |`) {
		t.Errorf("reassess did not overwrite:\n%s", got)
	}
	if strings.Count(got, "| MI-0001 |") != 1 {
		t.Errorf("reassess duplicated the row:\n%s", got)
	}

	// --json is the change.
	var js struct {
		ID         string `json:"id"`
		Candidates []struct {
			ID          string `json:"id"`
			Verdict     string `json:"verdict"`
			Why         string `json:"why"`
			NewTitle    string `json:"new_title"`
			NewDoneWhen string `json:"new_done_when"`
		} `json:"candidates"`
	}
	raw := a("MI-0001", "--verdict", "replace", "--reason", "Back to replace", "--title", "T2", "--done-when", "D2", "--json")
	if err := json.Unmarshal([]byte(raw), &js); err != nil || js.ID != "CHG-0001" || js.Candidates[0].Verdict != "replace" || js.Candidates[0].NewTitle != "T2" {
		t.Errorf("assess --json: %v\n%s", err, raw)
	}

	// Ticket show reports the verdict.
	if out := chMust(t, root, "ticket", "show", "MI-0003"); !strings.Contains(out, "CHG-0001 open, invalidate") {
		t.Errorf("ticket show:\n%s", out)
	}
}

func TestChangeAssessSkipsClosedTickets(t *testing.T) {
	root := chAssessPlan(t)
	// A dropped ticket can only be unaffected (or investigate).
	chMust(t, root, "ticket", "drop", "MI-0002", "--reason", "gone", "--by", "Aliyu")
	chRefused(t, root, ExitRefused, "already", "change", "assess", "CHG-0001", "MI-0002", "--verdict", "replace", "--reason", "x", "--title", "T", "--done-when", "D")
	chMust(t, root, "change", "assess", "CHG-0001", "MI-0002", "--verdict", "unaffected", "--reason", "already dropped")
}
