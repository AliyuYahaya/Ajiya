package agents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestClaudeConfigDir: with CLAUDE_CONFIG_DIR set, Claude Code's user config is
// $CLAUDE_CONFIG_DIR/.claude.json and ~/.claude.json is left alone; unset, it
// is ~/.claude.json.
func TestClaudeConfigDir(t *testing.T) {
	for _, set := range []bool{false, true} {
		name := "unset"
		if set {
			name = "set"
		}
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			home := filepath.Join(f.env.Home, ".claude.json")
			want := home
			if set {
				f.env.ClaudeConfigDir = filepath.Join(filepath.Dir(f.env.Home), "elsewhere")
				want = filepath.Join(f.env.ClaudeConfigDir, ".claude.json")
			}
			if got := f.env.ConfigPath(Claude, User); got != want {
				t.Fatalf("ConfigPath = %s, want %s", got, want)
			}
			if got := f.env.ConfigPath(Claude, Project); got != filepath.Join(f.env.Dir, ".mcp.json") {
				t.Errorf("project path changed: %s", got)
			}
			f.write(want, `{"theme": "dark"}`)

			if st, err := f.env.Status(Claude, User); err != nil || st != NotRegistered {
				t.Fatalf("status before = %v, %v", st, err)
			}
			f.apply(Claude, User, true)
			if st, err := f.env.Status(Claude, User); err != nil || st != Registered {
				t.Fatalf("status after install = %v, %v", st, err)
			}
			f.wantEntry(f.servers(want)["ajiya"])
			if !strings.Contains(f.read(want), `"theme"`) {
				t.Error("other settings were lost")
			}
			if set {
				if _, err := os.Stat(home); err == nil {
					t.Error("~/.claude.json was written although CLAUDE_CONFIG_DIR is set")
				}
			}
			f.apply(Claude, User, false)
			if st, _ := f.env.Status(Claude, User); st != NotRegistered {
				t.Errorf("status after uninstall = %v", st)
			}
		})
	}
}

// TestClaudeConfigDirIgnoresHomeFile: an entry in ~/.claude.json does not count
// while CLAUDE_CONFIG_DIR points elsewhere.
func TestClaudeConfigDirIgnoresHomeFile(t *testing.T) {
	f := newFixture(t)
	f.apply(Claude, User, true) // into ~/.claude.json
	f.env.ClaudeConfigDir = filepath.Join(filepath.Dir(f.env.Home), "elsewhere")
	if st, err := f.env.Status(Claude, User); err != nil || st != NotRegistered {
		t.Errorf("status = %v, %v; want not registered", st, err)
	}
}

func TestClaudeConfigDirDetect(t *testing.T) {
	f := newFixture(t)
	f.env.ClaudeConfigDir = filepath.Join(filepath.Dir(f.env.Home), "elsewhere")
	f.write(filepath.Join(f.env.Home, ".claude.json"), "{}") // the default location no longer counts
	if got := f.env.Detect(); len(got) != 0 {
		t.Errorf("detected %v", got)
	}
	f.write(filepath.Join(f.env.ClaudeConfigDir, "settings.json"), "{}")
	if got := f.env.Detect(); len(got) != 1 || got[0] != Claude {
		t.Errorf("detected %v, want claude", got)
	}
}

// TestClaudeCommandPreviewNamesConfigDirFile: with the claude command, the
// preview names the file under CLAUDE_CONFIG_DIR.
func TestClaudeCommandPreviewNamesConfigDirFile(t *testing.T) {
	f := newFixture(t)
	f.path["claude"] = true
	f.env.ClaudeConfigDir = filepath.Join(filepath.Dir(f.env.Home), "elsewhere")
	c, err := f.env.PlanInstall(Claude, User)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(f.env.ClaudeConfigDir, ".claude.json")
	if c.Path != want || !strings.Contains(strings.Join(c.Preview, "\n"), want) {
		t.Errorf("path %s, preview %v; want %s", c.Path, c.Preview, want)
	}
}

// TestNewEnvReadsClaudeConfigDir uses a temporary home and never the real one.
func TestNewEnvReadsClaudeConfigDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	dir := filepath.Join(home, "cc")
	for _, v := range []string{"", dir} {
		t.Setenv("CLAUDE_CONFIG_DIR", v)
		e, err := NewEnv(home)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(home, ".claude.json")
		if v != "" {
			want = filepath.Join(dir, ".claude.json")
		}
		if e.ClaudeConfigDir != v || e.ConfigPath(Claude, User) != want {
			t.Errorf("CLAUDE_CONFIG_DIR=%q: dir %q, path %s, want %s", v, e.ClaudeConfigDir, e.ConfigPath(Claude, User), want)
		}
	}
}

// TestRunnerPassesClaudeConfigDir: the claude command gets CLAUDE_CONFIG_DIR.
// The test binary stands in for it (see TestMain in helper_test.go).
func TestRunnerPassesClaudeConfigDir(t *testing.T) {
	t.Setenv("AJIYA_TEST_PRINT_ENV", "1")
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	out, err := runner("/some/dir")("", os.Args[0], "-test.run=^$")
	if err != nil {
		t.Fatal(err, out)
	}
	if !strings.Contains(out, "CLAUDE_CONFIG_DIR=/some/dir\n") {
		t.Errorf("child env: %q", out)
	}
}
