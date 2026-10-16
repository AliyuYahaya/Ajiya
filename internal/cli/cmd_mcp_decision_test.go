package cli

import (
	"strings"
	"testing"
)

// The four decision tools: add, show, list and accept (accept is allowed over
// MCP; only change approve and reject are person-only).
func TestMCPDecisions(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cs := mcpConnect(t, dir, "")
	for _, step := range []struct {
		tool string
		args map[string]any
	}{
		{"ajiya_init", map[string]any{"name": "Demo", "prefix": "DE"}},
		{"ajiya_phase_add", map[string]any{"slug": "core", "goal": "Core works"}},
		{"ajiya_decision_add", map[string]any{"title": "-Use PostgreSQL", "decision": "Use PostgreSQL on AWS RDS", "tech": "postgres,aws-rds", "by": "Sam", "context": "Need transactions"}},
		{"ajiya_ticket_add", map[string]any{"phase": "core", "app": "infra", "title": "Backups", "done_when": "ok", "decisions": "DEC-0001"}},
	} {
		if r := mcpCall(t, cs, step.tool, step.args); r.IsError {
			t.Fatalf("%s: %s", step.tool, resultText(r, 0))
		}
	}

	r := mcpCall(t, cs, "ajiya_decision_show", map[string]any{"id": "DEC-0001"})
	text := resultText(r, 1)
	for _, want := range []string{`"id": "DEC-0001"`, `"state": "proposed"`, `"by": "Sam"`, `"DE-0001"`, `"aws-rds"`, `"context": "Need transactions"`} {
		if r.IsError || !strings.Contains(text, want) {
			t.Errorf("decision show lacks %s: %v %q", want, r.IsError, text)
		}
	}
	r = mcpCall(t, cs, "ajiya_decision_list", map[string]any{"tag": "tech:postgres"})
	if r.IsError || !strings.Contains(resultText(r, 0), "1 decision(s)") {
		t.Errorf("decision list: %v %q", r.IsError, resultText(r, 0))
	}
	r = mcpCall(t, cs, "ajiya_decision_list", map[string]any{"tag": "tech:redis"})
	if r.IsError || strings.TrimSpace(resultText(r, 1)) != "[]" {
		t.Errorf("decision list with no match: %v %q", r.IsError, resultText(r, 1))
	}

	r = mcpCall(t, cs, "ajiya_decision_accept", map[string]any{"id": "DEC-0001", "by": "Aliyu"})
	if r.IsError || !strings.Contains(resultText(r, 0), "Accepted") {
		t.Fatalf("decision accept: %v %q", r.IsError, resultText(r, 0))
	}
	r = mcpCall(t, cs, "ajiya_decision_accept", map[string]any{"id": "DEC-0001", "by": "Aliyu"})
	if !r.IsError || !strings.Contains(resultText(r, 0), "already accepted") {
		t.Errorf("accepting twice: %v %q", r.IsError, resultText(r, 0))
	}

	// ticket show lists the linked decision with its status.
	r = mcpCall(t, cs, "ajiya_ticket_show", map[string]any{"id": "DE-0001"})
	if !strings.Contains(resultText(r, 1), `"state": "accepted"`) || !strings.Contains(resultText(r, 1), `"DEC-0001"`) {
		t.Errorf("ticket show lacks the decision: %q", resultText(r, 1))
	}
	r = mcpCall(t, cs, "ajiya_check", map[string]any{})
	if r.IsError || strings.TrimSpace(resultText(r, 1)) != "[]" {
		t.Errorf("check: %v %q", r.IsError, resultText(r, 1))
	}
}
