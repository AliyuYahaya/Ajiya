package cli

import (
	"strings"
	"testing"
)

// A change with nothing to apply is refused; a decision-only change is
// approved with no replacement decision.
func TestChangeApproveDecisionOnlyAndEmpty(t *testing.T) {
	root := chSupabase(t)
	chMust(t, root, "change", "open", "Nothing found", "--paths", "nowhere/**")
	chMust(t, root, "change", "impact", "CHG-0001")
	chCommitAll(t, root, "Ajiya: chore")
	chRefused(t, root, ExitRefused, "nothing to apply", "change", "approve", "CHG-0001", "--by", "Aliyu")
	chMust(t, root, "change", "reject", "CHG-0001", "--by", "Aliyu", "--reason", "empty")
	chCommitAll(t, root, "Ajiya: chore")

	chMust(t, root, "change", "open", "Drop the RDS decision", "--decision", "DEC-0001")
	chMust(t, root, "change", "impact", "CHG-0002")
	chMust(t, root, "change", "assess", "CHG-0002", "MI-0002", "--verdict", "unaffected", "--reason", "fine")
	chCommitAll(t, root, "Ajiya: chore")
	out := chMust(t, root, "change", "approve", "CHG-0002", "--by", "Aliyu")
	if !strings.Contains(out, "Superseded decision DEC-0001 with no replacement") || !strings.Contains(out, "No --new-decision was given") {
		t.Errorf("approve output:\n%s", out)
	}
	dec := chPlanFiles(t, root)["ajiya/decisions/DEC-0001.md"]
	if !strings.Contains(dec, "Superseded (CHG-0002)") {
		t.Errorf("DEC-0001:\n%s", dec)
	}
	if out, code := chRun(t, root, 0, "check"); code != 0 {
		t.Errorf("check: %d\n%s", code, out)
	}
}
