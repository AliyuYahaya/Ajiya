package cli

import (
	"strings"
	"testing"

	"github.com/AliyuYahaya/Ajiya/internal/agents"
)

func TestMCPStatusQuiet(t *testing.T) {
	m := newAgentMachine(t)
	check := func(want int, args ...string) {
		t.Helper()
		out, _, code := m.run(false, "", append([]string{"mcp", "status", "--quiet"}, args...)...)
		if code != want || out != "" {
			t.Errorf("%v: code %d (want %d), out %q", args, code, want, out)
		}
	}
	check(ExitUsage) // needs an agent
	check(ExitRefused, "--claude")
	m.run(false, "", "mcp", "install", "--claude", "--yes")
	check(0, "--claude")
	check(ExitRefused, "--claude", "--cursor")
	m.run(false, "", "mcp", "install", "--cursor", "--scope", "project", "--yes")
	check(0, "--claude", "--cursor") // project scope counts
	m.put(m.env.ConfigPath(agents.Codex, agents.User), "[oops")
	check(ExitUsage, "--codex") // unreadable config: neither yes nor no
	// the table can be limited to one agent, detected or not
	out, _, _ := m.run(false, "", "mcp", "status", "--claude")
	if !strings.Contains(out, "Claude Code") || strings.Contains(out, "Cursor") {
		t.Errorf("status --claude:\n%s", out)
	}
}
