package kit

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The plugin for Claude Code lives in plugin/ at the repository root.
var pluginDir = filepath.Join("..", "..", "plugin")

func readJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}

// The plugin's skill is the kit's skill. When the kit's SKILL.md changes, run
//
//	cp internal/kit/files/SKILL.md plugin/skills/ajiya/SKILL.md
//
// (a copy, not a symlink, because symlinks break on Windows checkouts).
func TestPluginSkillIsCurrent(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginDir, "skills", "ajiya", "SKILL.md"))
	if err != nil || string(b) != Source("SKILL.md") {
		t.Errorf("plugin/skills/ajiya/SKILL.md differs from the kit's SKILL.md (err %v); copy it over and commit", err)
	}
}

// claude plugin validate is the authoritative check and was run by hand; this
// test pins the fields the documented schema requires so 'go test' catches a
// broken manifest without the claude binary.
// https://code.claude.com/docs/en/plugins-reference
// https://code.claude.com/docs/en/plugin-marketplaces
func TestPluginManifests(t *testing.T) {
	var pj struct {
		Name, Version, Description string
		Author                     struct{ Name string }
	}
	readJSON(t, filepath.Join(pluginDir, ".claude-plugin", "plugin.json"), &pj)
	if pj.Name != "ajiya" || pj.Version == "" || pj.Description == "" || pj.Author.Name == "" {
		t.Errorf("plugin.json: %+v", pj)
	}

	// The plugin leaves MCP registration to 'ajiya mcp install': declaring the
	// server here as well would list Ajiya's tools twice in a session.
	if _, err := os.Stat(filepath.Join(pluginDir, ".mcp.json")); err == nil {
		t.Error("plugin/.mcp.json must not exist (AJ-0097)")
	}
	raw, err := os.ReadFile(filepath.Join(pluginDir, ".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "mcpServers") || strings.Contains(string(raw), ".mcp.json") {
		t.Error("plugin.json must not declare an MCP server")
	}

	var hooks struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Type, Command string
				Args          []string
			}
		}
	}
	readJSON(t, filepath.Join(pluginDir, "hooks", "hooks.json"), &hooks)
	ss := hooks.Hooks["SessionStart"]
	// Two shell-form hooks, one per platform family, each ending in "exit 0" so
	// a missing sh or powershell.exe is not reported as a hook failure:
	// https://code.claude.com/docs/en/hooks (command hook fields)
	if len(ss) != 1 || len(ss[0].Hooks) != 2 {
		t.Fatalf("hooks.json: want one SessionStart group with two hooks, got %+v", hooks)
	}
	for i, script := range []string{"session-start.sh", "session-start.ps1"} {
		h := ss[0].Hooks[i]
		if h.Type != "command" || !strings.Contains(h.Command, "\"${CLAUDE_PLUGIN_ROOT}/hooks/"+script+"\"") ||
			!strings.HasSuffix(h.Command, "; exit 0") {
			t.Errorf("hooks.json hook %d = %+v", i, h)
		}
		if _, err := os.Stat(filepath.Join(pluginDir, "hooks", script)); err != nil {
			t.Error(err)
		}
	}

	var mk struct {
		Name    string
		Owner   struct{ Name string }
		Plugins []struct{ Name, Source string }
	}
	readJSON(t, filepath.Join("..", "..", ".claude-plugin", "marketplace.json"), &mk)
	if mk.Name == "" || mk.Owner.Name == "" || len(mk.Plugins) != 1 {
		t.Fatalf("marketplace.json: %+v", mk)
	}
	p := mk.Plugins[0]
	if p.Name != pj.Name || !strings.HasPrefix(p.Source, "./") || strings.Contains(p.Source, "..") {
		t.Errorf("marketplace entry %+v must match the plugin name and use a ./ relative source", p)
	}
	if _, err := os.Stat(filepath.Join("..", "..", p.Source, ".claude-plugin", "plugin.json")); err != nil {
		t.Errorf("marketplace source has no plugin: %v", err)
	}
}

// What the fake ajiya answers to 'mcp status --claude --quiet'.
const (
	regYes   = 0 // registered with Claude Code
	regNo    = 1 // not registered
	regOlder = 2 // an older ajiya that does not know the flag
)

// hookCase is one behaviour of the session-start hook, for both platforms.
type hookCase struct {
	name    string
	project bool // run in a folder with ajiya.toml
	withBin bool // the fake ajiya is on PATH
	reg     int  // the fake's answer to 'mcp status --claude --quiet'
	want    string
	install bool // want the install-the-binary line instead of want
}

const (
	fakeStatus  = "FAKE STATUS\n"
	registerMsg = "Ajiya is not registered with Claude Code, so a session cannot use its tools. Register it with: ajiya mcp install --claude\n"
)

var hookCases = []hookCase{
	{"registered: status only", true, true, regYes, fakeStatus, false},
	{"not registered: status plus the install line", true, true, regNo, fakeStatus + registerMsg, false},
	{"older ajiya: status only", true, true, regOlder, fakeStatus, false},
	{"no ajiya.toml, not registered", false, true, regNo, "", false},
	{"no ajiya.toml, registered", false, true, regYes, "", false},
	{"no ajiya: install the binary", true, false, regNo, "", true},
	{"no ajiya, no ajiya.toml", false, false, regNo, "", false},
}

func checkHookOutput(t *testing.T, tc hookCase, got string) {
	t.Helper()
	got = strings.ReplaceAll(got, "\r\n", "\n")
	if tc.install {
		lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
		if len(lines) != 1 || !strings.Contains(got, "go install github.com/AliyuYahaya/Ajiya/cmd/ajiya@latest") || strings.Contains(got, "mcp install") {
			t.Errorf("want only the install-the-binary line, got %q", got)
		}
		return
	}
	if got != tc.want {
		t.Errorf("stdout = %q, want %q", got, tc.want)
	}
}

func newHookProject(t *testing.T) (project, plain string) {
	t.Helper()
	project, plain = t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "ajiya.toml"), []byte("name = \"x\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return project, plain
}

// runHook runs cmd, which must exit 0, and returns its stdout.
func runHook(t *testing.T, cmd *exec.Cmd) string {
	t.Helper()
	var out strings.Builder
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("hook failed: %v", err)
	}
	return out.String()
}

func TestSessionStartHook(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("session-start.sh is for macOS and Linux; Windows runs session-start.ps1 (TestSessionStartHookPowerShell)")
	}
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("no sh on PATH")
	}
	script, err := filepath.Abs(filepath.Join(pluginDir, "hooks", "session-start.sh"))
	if err != nil {
		t.Fatal(err)
	}
	bins := map[int]string{}
	for _, reg := range []int{regYes, regNo, regOlder} {
		bin := t.TempDir()
		fake := fmt.Sprintf("#!/bin/sh\n[ \"$1 $2\" = \"status --brief\" ] && echo 'FAKE STATUS'\n[ \"$1 $2 $3 $4\" = \"mcp status --claude --quiet\" ] && exit %d\nexit 0\n", reg)
		if err := os.WriteFile(filepath.Join(bin, "ajiya"), []byte(fake), 0o755); err != nil {
			t.Fatal(err)
		}
		bins[reg] = bin
	}
	bin := bins[regYes]
	project, plain := newHookProject(t)

	for _, tc := range hookCases {
		t.Run(tc.name, func(t *testing.T) {
			dir, path := plain, "/usr/bin:/bin"
			if tc.project {
				dir = project
			}
			if tc.withBin {
				path = bins[tc.reg] + ":" + path
			}
			cmd := exec.Command(sh, script)
			cmd.Dir = dir
			cmd.Env = []string{"PATH=" + path, "HOME=" + dir}
			checkHookOutput(t, tc, runHook(t, cmd))
		})
	}

	// CLAUDE_PROJECT_DIR wins over the working directory.
	t.Run("CLAUDE_PROJECT_DIR", func(t *testing.T) {
		cmd := exec.Command(sh, script)
		cmd.Dir = plain
		cmd.Env = []string{"PATH=" + bin + ":/usr/bin:/bin", "CLAUDE_PROJECT_DIR=" + project}
		if got := runHook(t, cmd); got != fakeStatus {
			t.Errorf("out %q", got)
		}
	})
}

// The same cases for session-start.ps1, which Claude Code runs on Windows
// through powershell.exe (Windows PowerShell 5.1, see plugin/hooks/hooks.json).
// It runs on the Windows CI runner and is skipped elsewhere: the script does
// nothing off Windows.
func TestSessionStartHookPowerShell(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("session-start.ps1 only acts on Windows")
	}
	ps, err := exec.LookPath("powershell")
	if err != nil {
		t.Skip("no powershell on PATH")
	}
	script, err := filepath.Abs(filepath.Join(pluginDir, "hooks", "session-start.ps1"))
	if err != nil {
		t.Fatal(err)
	}
	bins := map[int]string{}
	for _, reg := range []int{regYes, regNo, regOlder} {
		bin := t.TempDir()
		fake := fmt.Sprintf("@echo off\r\nif \"%%1 %%2\"==\"status --brief\" echo FAKE STATUS\r\nif \"%%1 %%2 %%3 %%4\"==\"mcp status --claude --quiet\" exit /b %d\r\nexit /b 0\r\n", reg)
		if err := os.WriteFile(filepath.Join(bin, "ajiya.cmd"), []byte(fake), 0o644); err != nil {
			t.Fatal(err)
		}
		bins[reg] = bin
	}
	bin := bins[regYes]
	project, plain := newHookProject(t)

	root := os.Getenv("SystemRoot")
	sys := filepath.Join(root, "System32")
	sysPath := strings.Join([]string{sys, root, filepath.Join(sys, "WindowsPowerShell", "v1.0")}, ";")
	run := func(t *testing.T, dir, path string, extra ...string) string {
		cmd := exec.Command(ps, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script)
		cmd.Dir = dir
		cmd.Env = append([]string{"PATH=" + path, "SystemRoot=" + root, "ComSpec=" + os.Getenv("ComSpec"), "PATHEXT=.COM;.EXE;.BAT;.CMD", "OS=Windows_NT"}, extra...)
		return runHook(t, cmd)
	}

	for _, tc := range hookCases {
		t.Run(tc.name, func(t *testing.T) {
			dir, path := plain, sysPath
			if tc.project {
				dir = project
			}
			if tc.withBin {
				path = bins[tc.reg] + ";" + path
			}
			checkHookOutput(t, tc, run(t, dir, path))
		})
	}

	// CLAUDE_PROJECT_DIR wins over the working directory.
	t.Run("CLAUDE_PROJECT_DIR", func(t *testing.T) {
		got := run(t, plain, bin+";"+sysPath, "CLAUDE_PROJECT_DIR="+project)
		if strings.ReplaceAll(got, "\r\n", "\n") != "FAKE STATUS\n" {
			t.Errorf("out %q", got)
		}
	})

	// A CLAUDE_PROJECT_DIR that does not exist is silent and exits 0.
	t.Run("bad CLAUDE_PROJECT_DIR", func(t *testing.T) {
		if got := run(t, plain, bin+";"+sysPath, "CLAUDE_PROJECT_DIR="+filepath.Join(plain, "missing")); got != "" {
			t.Errorf("out %q", got)
		}
	})
}
