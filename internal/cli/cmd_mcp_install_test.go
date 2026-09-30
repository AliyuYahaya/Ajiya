package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AliyuYahaya/Ajiya/internal/agents"
)

// agentMachine is a temporary home folder with a fake agent environment.
type agentMachine struct {
	t     *testing.T
	env   *agents.Env
	path  map[string]bool
	calls [][]string
}

func newAgentMachine(t *testing.T) *agentMachine {
	t.Helper()
	root := t.TempDir()
	m := &agentMachine{t: t, path: map[string]bool{}}
	bin := filepath.Join(root, "ajiya")
	if err := os.WriteFile(bin, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	proj := filepath.Join(root, "proj")
	if err := os.MkdirAll(proj, 0o755); err != nil {
		t.Fatal(err)
	}
	m.env = &agents.Env{
		Home: filepath.Join(root, "home"), CodexHome: filepath.Join(root, "home", ".codex"), Dir: proj, Binary: bin,
		LookPath: func(n string) (string, error) {
			if m.path[n] {
				return "/fake/" + n, nil
			}
			return "", errors.New("no")
		},
		Run: func(dir, name string, args ...string) (string, error) {
			m.calls = append(m.calls, append([]string{name}, args...))
			return "", nil
		},
	}
	return m
}

func (m *agentMachine) run(interactive bool, stdin string, args ...string) (string, string, int) {
	var o, e bytes.Buffer
	code := run(&env{stdin: strings.NewReader(stdin), stdout: &o, stderr: &e, dir: m.env.Dir, interactive: interactive, agents: m.env}, args)
	return o.String(), e.String(), code
}

func (m *agentMachine) put(p, s string) {
	m.t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		m.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		m.t.Fatal(err)
	}
}

func (m *agentMachine) get(p string) string {
	b, _ := os.ReadFile(p)
	return string(b)
}

func TestMCPInstallNeedsYesWhenNotInteractive(t *testing.T) {
	m := newAgentMachine(t)
	out, _, code := m.run(false, "", "mcp", "install", "--cursor")
	p := m.env.ConfigPath(agents.Cursor, agents.User)
	if code != ExitRefused || !strings.Contains(out, "--yes") || !strings.Contains(out, `"ajiya"`) {
		t.Errorf("code %d, out:\n%s", code, out)
	}
	if _, err := os.Stat(p); err == nil {
		t.Error("wrote without --yes")
	}
}

func TestMCPInstallUninstallAllAgents(t *testing.T) {
	m := newAgentMachine(t)
	m.put(m.env.ConfigPath(agents.Codex, agents.User), "# mine\nmodel = \"x\"\n")
	out, errOut, code := m.run(false, "", "mcp", "install", "--all", "--yes")
	if code != 0 {
		t.Fatalf("code %d\n%s%s", code, out, errOut)
	}
	for _, a := range agents.All {
		p := m.env.ConfigPath(a, agents.User)
		if !strings.Contains(m.get(p), "ajiya") {
			t.Errorf("%s: not registered in %s", a, p)
		}
	}
	if !strings.HasPrefix(m.get(m.env.ConfigPath(agents.Codex, agents.User)), "# mine\nmodel = \"x\"\n") {
		t.Error("codex file changed above the new table")
	}

	// twice: nothing changes
	out, _, code = m.run(false, "", "mcp", "install", "--all", "--yes")
	if code != 0 || strings.Count(out, "nothing to change") != 3 {
		t.Errorf("second install, code %d:\n%s", code, out)
	}

	st, _, _ := m.run(false, "", "mcp", "status")
	if strings.Count(st, "registered") != 3 || strings.Contains(st, "not registered") || strings.Contains(st, "Register with") {
		t.Errorf("status:\n%s", st)
	}

	if _, _, code = m.run(false, "", "mcp", "uninstall", "--all", "--yes"); code != 0 {
		t.Fatal("uninstall failed")
	}
	for _, a := range agents.All {
		if strings.Contains(m.get(m.env.ConfigPath(a, agents.User)), "ajiya") {
			t.Errorf("%s: still registered", a)
		}
	}
	if got := m.get(m.env.ConfigPath(agents.Codex, agents.User)); got != "# mine\nmodel = \"x\"\n" {
		t.Errorf("codex file after uninstall: %q", got)
	}
}

func TestMCPInstallDefaultsToAgentsFound(t *testing.T) {
	m := newAgentMachine(t)
	if _, e, code := m.run(false, "", "mcp", "install", "--yes"); code != ExitRefused || !strings.Contains(e, "no supported agent found") {
		t.Errorf("empty machine: code %d, %q", code, e)
	}
	m.put(filepath.Join(m.env.Home, ".cursor", "x"), "")
	if _, _, code := m.run(false, "", "mcp", "install", "--yes"); code != 0 {
		t.Fatal("install failed")
	}
	if !strings.Contains(m.get(m.env.ConfigPath(agents.Cursor, agents.User)), "ajiya") {
		t.Error("cursor not registered")
	}
	if _, err := os.Stat(m.env.ConfigPath(agents.Codex, agents.User)); err == nil {
		t.Error("registered with an agent that is not on the machine")
	}
}

func TestMCPInstallProjectScope(t *testing.T) {
	m := newAgentMachine(t)
	if _, _, code := m.run(false, "", "mcp", "install", "--all", "--scope", "project", "--yes"); code != 0 {
		t.Fatal("install failed")
	}
	for _, rel := range []string{".mcp.json", filepath.Join(".cursor", "mcp.json"), filepath.Join(".codex", "config.toml")} {
		if !strings.Contains(m.get(filepath.Join(m.env.Dir, rel)), "ajiya") {
			t.Errorf("%s missing the entry", rel)
		}
	}
	if _, err := os.Stat(m.env.Home); err == nil {
		t.Error("project scope wrote in the home folder")
	}
	if _, e, code := m.run(false, "", "mcp", "install", "--scope", "everywhere"); code != ExitUsage || !strings.Contains(e, "--scope") {
		t.Errorf("bad scope: %d %q", code, e)
	}
}

func TestMCPInstallMalformedWritesNothing(t *testing.T) {
	m := newAgentMachine(t)
	bad := m.env.ConfigPath(agents.Codex, agents.User)
	m.put(bad, "[oops")
	_, e, code := m.run(false, "", "mcp", "install", "--all", "--yes")
	if code != ExitRefused || !strings.Contains(e, "not valid TOML") || !strings.Contains(e, "Codex") {
		t.Errorf("code %d, %q", code, e)
	}
	if m.get(bad) != "[oops" {
		t.Error("malformed file changed")
	}
	// the others were not written either: all or nothing
	for _, a := range []agents.Agent{agents.Claude, agents.Cursor} {
		if _, err := os.Stat(m.env.ConfigPath(a, agents.User)); err == nil {
			t.Errorf("%s was written although another config was refused", a)
		}
	}
}

func TestMCPInstallAsks(t *testing.T) {
	m := newAgentMachine(t)
	p := m.env.ConfigPath(agents.Cursor, agents.User)
	if out, _, code := m.run(true, "n\n", "mcp", "install", "--cursor"); code != ExitRefused || !strings.Contains(out, "Nothing was changed") {
		t.Errorf("no: code %d\n%s", code, out)
	}
	if _, err := os.Stat(p); err == nil {
		t.Error("wrote after a no")
	}
	if _, _, code := m.run(true, "y\n", "mcp", "install", "--cursor"); code != 0 {
		t.Fatal("yes failed")
	}
	if !strings.Contains(m.get(p), "ajiya") {
		t.Error("not written after a yes")
	}
}

func TestMCPInstallClaudeCommand(t *testing.T) {
	m := newAgentMachine(t)
	m.path["claude"] = true
	if _, _, code := m.run(false, "", "mcp", "install", "--claude", "--scope", "project", "--yes"); code != 0 {
		t.Fatal("install failed")
	}
	if len(m.calls) != 1 || strings.Join(m.calls[0], " ") != "claude mcp add --scope project ajiya -- "+m.env.Binary+" mcp" {
		t.Errorf("calls: %v", m.calls)
	}
}

func TestMCPStatusListsFoundAgents(t *testing.T) {
	m := newAgentMachine(t)
	if out, _, code := m.run(false, "", "mcp", "status"); code != 0 || !strings.Contains(out, "No supported agent") {
		t.Errorf("empty: %d %s", code, out)
	}
	m.path["codex"] = true
	m.put(filepath.Join(m.env.Home, ".cursor", "mcp.json"), `{"mcpServers":{}}`)
	m.put(m.env.ConfigPath(agents.Cursor, agents.Project), `{"mcpServers":{"ajiya":{"command":"ajiya","args":["mcp"]}}}`)
	out, _, _ := m.run(false, "", "mcp", "status")
	for _, want := range []string{"Codex", "Cursor", "not registered", "project", "registered", "ajiya mcp install --codex --cursor"} {
		if !strings.Contains(out, want) {
			t.Errorf("status lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Claude Code") {
		t.Errorf("listed an agent that is not on the machine:\n%s", out)
	}
	m.put(m.env.ConfigPath(agents.Codex, agents.User), "[oops")
	if out, _, code := m.run(false, "", "mcp", "status"); code != 0 || !strings.Contains(out, "cannot read") {
		t.Errorf("malformed: %d\n%s", code, out)
	}
}

func TestInitOffersToRegister(t *testing.T) {
	newProject := func(m *agentMachine) {
		m.put(filepath.Join(m.env.Dir, "go.mod"), "module x\n")
	}
	m := newAgentMachine(t)
	newProject(m)
	m.put(filepath.Join(m.env.Home, ".cursor", "x"), "")

	// not interactive, no --yes: the one command, nothing written
	out, _, code := m.run(false, "", "init")
	if code != 0 || !strings.Contains(out, "ajiya mcp install --cursor") {
		t.Fatalf("code %d\n%s", code, out)
	}
	if _, err := os.Stat(m.env.ConfigPath(agents.Cursor, agents.User)); err == nil {
		t.Fatal("init registered without being asked")
	}

	// --yes registers (init on an already set up project does too)
	out, _, code = m.run(false, "", "init", "--yes")
	if code != 0 || !strings.Contains(out, "Registered Ajiya with Cursor") {
		t.Fatalf("code %d\n%s", code, out)
	}
	if !strings.Contains(m.get(m.env.ConfigPath(agents.Cursor, agents.User)), "ajiya") {
		t.Error("not registered")
	}
	// registered already: init says nothing about it
	if out, _, _ = m.run(false, "", "init"); strings.Contains(out, "not registered") {
		t.Errorf("offered again:\n%s", out)
	}

	// at a terminal it asks
	m2 := newAgentMachine(t)
	newProject(m2)
	m2.put(filepath.Join(m2.env.Home, ".cursor", "x"), "")
	if out, _, _ = m2.run(true, "\ny\n", "init"); !strings.Contains(out, "Registered Ajiya with Cursor") {
		t.Errorf("interactive:\n%s", out)
	}
}

func TestInitOverMCPNeverRegisters(t *testing.T) {
	// ajiya_init runs with noAgents set, so an agent calling it cannot change
	// the person's agent config.
	m := newAgentMachine(t)
	m.put(filepath.Join(m.env.Dir, "go.mod"), "module x\n")
	m.put(filepath.Join(m.env.Home, ".cursor", "x"), "")
	var o, e bytes.Buffer
	code := run(&env{stdin: strings.NewReader(""), stdout: &o, stderr: &e, dir: m.env.Dir, agents: m.env, noAgents: true}, []string{"init", "--yes"})
	if code != 0 || strings.Contains(o.String(), "Registered Ajiya") {
		t.Fatalf("code %d\n%s%s", code, o.String(), e.String())
	}
	if _, err := os.Stat(m.env.ConfigPath(agents.Cursor, agents.User)); err == nil {
		t.Error("registered over MCP")
	}
}
