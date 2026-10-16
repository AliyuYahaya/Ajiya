package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// A milestone whose only target is invalidated, with nothing it depended on,
// would be left empty: approve is refused and writes nothing, ajiya.toml
// included.
func TestChangeApproveRefusesAnEmptiedMilestone(t *testing.T) {
	root := chSupabase(t)
	chMust(t, root, "milestone", "add", "solo", "--targets", "MI-0002")
	chMust(t, root, "change", "open", "Drop the RDS decision", "--decision", "DEC-0001")
	chMust(t, root, "change", "impact", "CHG-0001")
	chMust(t, root, "change", "assess", "CHG-0001", "MI-0002", "--verdict", "invalidate", "--reason", "No longer needed")
	chCommitAll(t, root, "Ajiya: chore")
	before := chPlanFiles(t, root)
	cfgBefore, _ := config.ReadText(root)
	chRefused(t, root, ExitRefused, "milestone solo would be left with no targets", "change", "approve", "CHG-0001", "--by", "Aliyu")
	if !mapsEqual(before, chPlanFiles(t, root)) {
		t.Error("a refused approve changed the plan")
	}
	if cfgAfter, _ := config.ReadText(root); cfgAfter != cfgBefore {
		t.Error("a refused approve changed ajiya.toml")
	}
}

// The one-target [launch] form takes one target: a replaced ticket is swapped
// in place, comments kept; two targets are refused.
func TestRetargetConfigLaunchForm(t *testing.T) {
	dir := t.TempDir()
	text := "# my project\n[project]\nname = \"x\"\nprefix = \"MI\"\n\n[launch]\ntarget = \"MI-0003\" # ship it\n"
	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	pr := &project{cfg: &config.Config{Root: dir, Launch: config.Launch{Target: "MI-0003"}}}
	_, got, err := retargetConfig(pr, []plan.Retarget{{Name: "launch", Targets: []string{"MI-0009"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `target = "MI-0009" # ship it`) || !strings.HasPrefix(got, "# my project\n") {
		t.Errorf("retargeted text:\n%s", got)
	}
	if _, _, err := retargetConfig(pr, []plan.Retarget{{Name: "launch", Targets: []string{"MI-0001", "MI-0002"}}}); err == nil || !strings.Contains(err.Error(), "takes one") {
		t.Errorf("two launch targets: %v", err)
	}
}

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
