package kit

import (
	"encoding/json"
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

	var mcp struct {
		McpServers map[string]struct {
			Command string
			Args    []string
		}
	}
	readJSON(t, filepath.Join(pluginDir, ".mcp.json"), &mcp)
	if s := mcp.McpServers["ajiya"]; s.Command != "ajiya" || len(s.Args) != 1 || s.Args[0] != "mcp" {
		t.Errorf(".mcp.json server = %+v, want ajiya mcp", s)
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
	if len(ss) != 1 || len(ss[0].Hooks) != 1 {
		t.Fatalf("hooks.json: want one SessionStart hook, got %+v", hooks)
	}
	h := ss[0].Hooks[0]
	if h.Type != "command" || h.Command != "sh" || len(h.Args) != 1 ||
		h.Args[0] != "${CLAUDE_PLUGIN_ROOT}/hooks/session-start.sh" {
		t.Errorf("hooks.json hook = %+v", h)
	}
	if _, err := os.Stat(filepath.Join(pluginDir, "hooks", "session-start.sh")); err != nil {
		t.Error(err)
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

// runHook runs the session-start script in dir, with a PATH of binDirs plus the
// system directories, and returns its stdout.
func runHook(t *testing.T, dir string, binDirs ...string) string {
	t.Helper()
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("no sh on PATH")
	}
	script, err := filepath.Abs(filepath.Join(pluginDir, "hooks", "session-start.sh"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(sh, script)
	cmd.Dir = dir
	cmd.Env = []string{"PATH=" + strings.Join(append(binDirs, "/usr/bin", "/bin"), ":"), "HOME=" + dir}
	var out strings.Builder
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("hook failed: %v", err)
	}
	return out.String()
}

func TestSessionStartHook(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the hook is a POSIX sh script; on Windows Claude Code runs it through Git Bash, which CI does not exercise")
	}
	bin := t.TempDir()
	fake := "#!/bin/sh\n[ \"$1 $2\" = \"status --brief\" ] && echo 'FAKE STATUS'\nexit 0\n"
	if err := os.WriteFile(filepath.Join(bin, "ajiya"), []byte(fake), 0o755); err != nil {
		t.Fatal(err)
	}

	project := t.TempDir()
	os.WriteFile(filepath.Join(project, "ajiya.toml"), []byte("name = \"x\"\n"), 0o644)
	plain := t.TempDir()

	tests := []struct {
		name    string
		dir     string
		bins    []string
		want    string // exact stdout, or a prefix when contains is set
		install bool
	}{
		{"ajiya on PATH, project", project, []string{bin}, "FAKE STATUS\n", false},
		{"ajiya on PATH, no ajiya.toml", plain, []string{bin}, "", false},
		{"no ajiya, project", project, nil, "", true},
		{"no ajiya, no ajiya.toml", plain, nil, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runHook(t, tt.dir, tt.bins...)
			if tt.install {
				lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
				if len(lines) > 2 || !strings.Contains(got, "go install github.com/AliyuYahaya/Ajiya/cmd/ajiya@latest") {
					t.Errorf("want the install line (at most two lines), got %q", got)
				}
				return
			}
			if got != tt.want {
				t.Errorf("stdout = %q, want %q", got, tt.want)
			}
		})
	}

	// CLAUDE_PROJECT_DIR wins over the working directory.
	t.Run("CLAUDE_PROJECT_DIR", func(t *testing.T) {
		sh, _ := exec.LookPath("sh")
		script, _ := filepath.Abs(filepath.Join(pluginDir, "hooks", "session-start.sh"))
		cmd := exec.Command(sh, script)
		cmd.Dir = plain
		cmd.Env = []string{"PATH=" + bin + ":/usr/bin:/bin", "CLAUDE_PROJECT_DIR=" + project}
		out, err := cmd.Output()
		if err != nil || string(out) != "FAKE STATUS\n" {
			t.Errorf("out %q err %v", out, err)
		}
	})
}
