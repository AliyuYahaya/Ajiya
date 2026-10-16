package cli

import (
	"strings"
	"testing"
)

// The five change tools: open, impact, assess, list and show. approve and reject are
// not tools (see TestMCPExcludedTools).
func TestMCPChanges(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"AUTHOR", "COMMITTER"} {
		t.Setenv("GIT_"+k+"_NAME", "T")
		t.Setenv("GIT_"+k+"_EMAIL", "t@example.com")
	}
	cs := mcpConnect(t, dir, "")
	for _, step := range []struct {
		tool string
		args map[string]any
	}{
		{"ajiya_init", map[string]any{"name": "Demo", "prefix": "DE"}},
		{"ajiya_phase_add", map[string]any{"slug": "core", "goal": "Core works"}},
		{"ajiya_ticket_add", map[string]any{"phase": "core", "app": "infra", "title": "Backups", "done_when": "ok", "tech": "aws-rds"}},
		{"ajiya_ticket_add", map[string]any{"phase": "core", "app": "infra", "title": "Alerts", "done_when": "ok", "depends": "DE-0001"}},
	} {
		if r := mcpCall(t, cs, step.tool, step.args); r.IsError {
			t.Fatalf("%s: %s", step.tool, resultText(r, 0))
		}
	}

	// Not a git repository yet, or the plan not committed: open is refused, not a crash.
	r := mcpCall(t, cs, "ajiya_change_open", map[string]any{"title": "Move", "replaces": "tech:aws-rds"})
	if !r.IsError {
		t.Fatalf("open outside a repository: %q", resultText(r, 0))
	}
	mrGit(t, dir, "init", "-q", "-b", "main")
	r = mcpCall(t, cs, "ajiya_change_open", map[string]any{"title": "Move", "replaces": "tech:aws-rds"})
	if !r.IsError || !strings.Contains(resultText(r, 0), "commit the plan first") {
		t.Fatalf("open before the first commit: %v %q", r.IsError, resultText(r, 0))
	}
	mrGit(t, dir, "add", "ajiya", "ajiya.toml")
	mrGit(t, dir, "commit", "-q", "-m", "Plan", "-m", "Ajiya: chore")

	r = mcpCall(t, cs, "ajiya_change_open", map[string]any{"title": "Move the database", "replaces": "tech:aws-rds", "with": "tech:supabase", "paths": "infra/rds/**"})
	if r.IsError || !strings.Contains(resultText(r, 0), "Opened CHG-0001") {
		t.Fatalf("change open: %v %q", r.IsError, resultText(r, 0))
	}
	r = mcpCall(t, cs, "ajiya_change_open", map[string]any{"title": "Nothing"})
	if !r.IsError {
		t.Errorf("open with nothing to touch was allowed: %q", resultText(r, 0))
	}

	r = mcpCall(t, cs, "ajiya_change_impact", map[string]any{"id": "CHG-0001"})
	text := resultText(r, 1)
	for _, want := range []string{`"id": "CHG-0001"`, `"tag:tech:aws-rds"`, `"downstream:DE-0001"`, `"added": [`, `"DE-0002"`} {
		if r.IsError || !strings.Contains(text, want) {
			t.Errorf("change impact lacks %s: %v %q", want, r.IsError, text)
		}
	}
	r = mcpCall(t, cs, "ajiya_change_assess", map[string]any{"id": "CHG-0001", "ticket": "DE-0002", "verdict": "replace", "reason": "Needs new alerts", "title": "Supabase alerts", "done_when": "alerts fire"})
	if text := resultText(r, 1); r.IsError || !strings.Contains(text, `"verdict": "replace"`) || !strings.Contains(text, `"new_title": "Supabase alerts"`) || !strings.Contains(text, `"new_done_when": "alerts fire"`) {
		t.Fatalf("change assess: %v %q", r.IsError, text)
	}
	r = mcpCall(t, cs, "ajiya_change_assess", map[string]any{"id": "CHG-0001", "ticket": "DE-0002", "verdict": "unaffected", "reason": "Second look", "title": "x"})
	if !r.IsError || !strings.Contains(resultText(r, 0), "--title and --done-when are for replace and modify") {
		t.Errorf("assess with new text on unaffected: %v %q", r.IsError, resultText(r, 0))
	}
	r = mcpCall(t, cs, "ajiya_change_assess", map[string]any{"id": "CHG-0001", "ticket": "DE-0001", "verdict": "investigate"})
	if !r.IsError {
		t.Errorf("assess without a reason was allowed: %q", resultText(r, 0))
	}
	r = mcpCall(t, cs, "ajiya_change_list", map[string]any{})
	if r.IsError || !strings.Contains(resultText(r, 1), `"CHG-0001"`) {
		t.Errorf("change list: %v %q", r.IsError, resultText(r, 1))
	}
	r = mcpCall(t, cs, "ajiya_change_show", map[string]any{"id": "CHG-0001"})
	text = resultText(r, 1)
	if r.IsError || !strings.Contains(text, `"state": "open"`) || !strings.Contains(text, `"before-CHG-0001"`) || !strings.Contains(text, `"infra/rds/**"`) {
		t.Errorf("change show: %v %q", r.IsError, text)
	}
	r = mcpCall(t, cs, "ajiya_change_show", map[string]any{"id": "CHG-0099"})
	if !r.IsError || !strings.Contains(resultText(r, 0), "no change CHG-0099") {
		t.Errorf("change show of a missing change: %v %q", r.IsError, resultText(r, 0))
	}
	r = mcpCall(t, cs, "ajiya_ticket_show", map[string]any{"id": "DE-0002"})
	if !strings.Contains(resultText(r, 1), `"changes": [`) || !strings.Contains(resultText(r, 1), `"CHG-0001"`) {
		t.Errorf("ticket show lacks the change: %q", resultText(r, 1))
	}
	r = mcpCall(t, cs, "ajiya_check", map[string]any{})
	if r.IsError || strings.TrimSpace(resultText(r, 1)) != "[]" {
		t.Errorf("check: %v %q", r.IsError, resultText(r, 1))
	}
}
