package agents

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

// fixture is a machine in a temporary home folder. No agent command is on PATH
// unless a test adds one.
type fixture struct {
	t     *testing.T
	env   *Env
	calls [][]string // commands the fake runner was asked to run
	path  map[string]bool
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	root := t.TempDir()
	f := &fixture{t: t, path: map[string]bool{}}
	bin := filepath.Join(root, "bin", "ajiya")
	if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	f.env = &Env{
		Home:      filepath.Join(root, "home"),
		CodexHome: filepath.Join(root, "home", ".codex"),
		Dir:       filepath.Join(root, "project"),
		Binary:    bin,
		LookPath: func(name string) (string, error) {
			if f.path[name] {
				return "/fake/" + name, nil
			}
			return "", errors.New("not found")
		},
		Run: func(dir, name string, args ...string) (string, error) {
			f.calls = append(f.calls, append([]string{dir, name}, args...))
			return "", nil
		},
	}
	for _, d := range []string{f.env.Home, f.env.Dir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return f
}

func (f *fixture) write(p, s string) {
	f.t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) read(p string) string {
	f.t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		f.t.Fatal(err)
	}
	return string(b)
}

func (f *fixture) apply(a Agent, s Scope, install bool) *Change {
	f.t.Helper()
	plan := f.env.PlanInstall
	if !install {
		plan = f.env.PlanUninstall
	}
	c, err := plan(a, s)
	if err != nil {
		f.t.Fatalf("%s %s: %v", a, s, err)
	}
	if err := c.Apply(); err != nil {
		f.t.Fatal(err)
	}
	return c
}

// servers reads the mcpServers of a JSON config.
func (f *fixture) servers(p string) map[string]map[string]any {
	f.t.Helper()
	var doc struct {
		Servers map[string]map[string]any `json:"mcpServers"`
	}
	if err := json.Unmarshal([]byte(f.read(p)), &doc); err != nil {
		f.t.Fatalf("%s is not valid JSON: %v", p, err)
	}
	return doc.Servers
}

func (f *fixture) wantEntry(en map[string]any) {
	f.t.Helper()
	if en == nil || en["command"] != f.env.Binary || !reflect.DeepEqual(en["args"], []any{"mcp"}) || en["type"] != "stdio" {
		f.t.Errorf("entry = %v, want stdio %s mcp", en, f.env.Binary)
	}
}

// jsonAgents run the same tests: their configs are JSON with an mcpServers
// object. Claude Code is one of them when its command is missing.
var jsonAgents = []struct {
	agent Agent
	scope Scope
	rel   string // path below home (user) or the project (project)
}{
	{Claude, User, ".claude.json"},
	{Claude, Project, ".mcp.json"},
	{Cursor, User, filepath.Join(".cursor", "mcp.json")},
	{Cursor, Project, filepath.Join(".cursor", "mcp.json")},
}

func (f *fixture) jsonPath(rel string, s Scope) string {
	if s == Project {
		return filepath.Join(f.env.Dir, rel)
	}
	return filepath.Join(f.env.Home, rel)
}

const otherJSON = `{
  "numStartups": 12,
  "theme": "dark",
  "mcpServers": {
    "context7": {"type": "stdio", "command": "npx", "args": ["-y", "@upstash/context7-mcp"]}
  },
  "zzz": [1, 2, 3]
}
`

func TestJSONConfigs(t *testing.T) {
	for _, tc := range jsonAgents {
		t.Run(string(tc.agent)+"/"+string(tc.scope), func(t *testing.T) {
			f := newFixture(t)
			p := f.jsonPath(tc.rel, tc.scope)
			if got := f.env.ConfigPath(tc.agent, tc.scope); got != p {
				t.Fatalf("ConfigPath = %s, want %s", got, p)
			}

			// first install creates the file (and its folder)
			c := f.apply(tc.agent, tc.scope, true)
			if c.Noop() || c.Action != "add" || len(c.Preview) == 0 {
				t.Fatalf("first install: action %q, preview %v", c.Action, c.Preview)
			}
			f.wantEntry(f.servers(p)["ajiya"])
			if _, err := os.Stat(p + BackupSuffix); err == nil {
				t.Error("a backup was made of a file that did not exist")
			}

			// second install changes nothing
			before := f.read(p)
			c = f.apply(tc.agent, tc.scope, true)
			if !c.Noop() || f.read(p) != before {
				t.Errorf("second install changed something: %q", c.Action)
			}

			// uninstall, then nothing left to remove
			f.apply(tc.agent, tc.scope, false)
			if _, ok := f.servers(p)["ajiya"]; ok {
				t.Error("entry still there after uninstall")
			}
			if c := f.apply(tc.agent, tc.scope, false); !c.Noop() {
				t.Error("second uninstall was not a no-op")
			}

			// other servers and other settings survive, in their order, with a backup
			f.write(p, otherJSON)
			f.apply(tc.agent, tc.scope, true)
			srv := f.servers(p)
			f.wantEntry(srv["ajiya"])
			if srv["context7"]["command"] != "npx" {
				t.Errorf("other server lost: %v", srv)
			}
			if got := f.read(p + BackupSuffix); got != otherJSON {
				t.Errorf("backup = %q, want the original file", got)
			}
			text := f.read(p)
			for _, k := range []string{`"numStartups": 12`, `"theme": "dark"`, `"zzz"`} {
				if !strings.Contains(text, k) {
					t.Errorf("lost %s in\n%s", k, text)
				}
			}
			if strings.Index(text, `"numStartups"`) > strings.Index(text, `"mcpServers"`) || strings.Index(text, `"mcpServers"`) > strings.Index(text, `"zzz"`) {
				t.Errorf("keys were reordered:\n%s", text)
			}
			f.apply(tc.agent, tc.scope, false)
			srv = f.servers(p)
			if _, ok := srv["ajiya"]; ok || srv["context7"]["command"] != "npx" {
				t.Errorf("uninstall left %v", srv)
			}
			if !strings.Contains(f.read(p), `"theme": "dark"`) {
				t.Error("uninstall lost another setting")
			}
		})
	}
}

func TestJSONStaleEntryIsUpdated(t *testing.T) {
	f := newFixture(t)
	p := f.env.ConfigPath(Cursor, User)
	f.write(p, `{"mcpServers":{"ajiya":{"command":"/gone/ajiya","args":["mcp"]},"x":{"command":"y"}}}`)
	if st, _ := f.env.Status(Cursor, User); st != Different {
		t.Fatalf("status = %v, want Different", st)
	}
	c := f.apply(Cursor, User, true)
	if c.Action != "update" {
		t.Fatalf("action = %q", c.Action)
	}
	srv := f.servers(p)
	f.wantEntry(srv["ajiya"])
	if srv["x"]["command"] != "y" {
		t.Error("other server lost")
	}
}

func TestHandWrittenEntryIsLeftAlone(t *testing.T) {
	f := newFixture(t)
	p := f.env.ConfigPath(Cursor, User)
	const text = `{"mcpServers":{"ajiya":{"command":"ajiya","args":["mcp","--dir","/x"]}}}`
	f.write(p, text)
	if c := f.apply(Cursor, User, true); !c.Noop() || f.read(p) != text {
		t.Error("an entry that already runs ajiya mcp was rewritten")
	}
}

func TestMalformedJSONIsRefused(t *testing.T) {
	bad := map[string]string{
		"syntax":         `{"mcpServers": {`,
		"not an object":  `[1,2]`,
		"servers a list": `{"mcpServers": []}`,
		"servers null":   `{"mcpServers": null}`,
		"trailing":       `{} {}`,
	}
	for name, text := range bad {
		for _, tc := range jsonAgents {
			for _, install := range []bool{true, false} {
				f := newFixture(t)
				p := f.jsonPath(tc.rel, tc.scope)
				f.write(p, text)
				plan := f.env.PlanInstall
				if !install {
					plan = f.env.PlanUninstall
				}
				_, err := plan(tc.agent, tc.scope)
				if err == nil || !strings.Contains(err.Error(), "Nothing was changed") || !strings.Contains(err.Error(), p) {
					t.Errorf("%s %s/%s install=%v: err = %v", name, tc.agent, tc.scope, install, err)
				}
				if f.read(p) != text {
					t.Errorf("%s: the file was changed", name)
				}
				if _, err := os.Stat(p + BackupSuffix); err == nil {
					t.Errorf("%s: a backup was written", name)
				}
				if _, err := f.env.Status(tc.agent, tc.scope); err == nil {
					t.Errorf("%s: Status did not report the malformed file", name)
				}
			}
		}
	}
}

func TestEmptyJSONFileCountsAsEmpty(t *testing.T) {
	f := newFixture(t)
	p := f.env.ConfigPath(Cursor, User)
	f.write(p, "")
	f.apply(Cursor, User, true)
	f.wantEntry(f.servers(p)["ajiya"])
}

// The Windows form of a path must survive both formats.
func TestWindowsBinaryPath(t *testing.T) {
	const win = `C:\Users\Ada Lovelace\bin\ajiya.exe`
	f := newFixture(t)
	f.env.Binary = win
	pj := f.env.ConfigPath(Cursor, User)
	f.apply(Cursor, User, true)
	if got := f.servers(pj)["ajiya"]["command"]; got != win {
		t.Errorf("JSON command = %v, want %s", got, win)
	}
	pt := f.env.ConfigPath(Codex, User)
	f.apply(Codex, User, true)
	var doc map[string]map[string]map[string]any
	if _, err := toml.Decode(f.read(pt), &doc); err != nil {
		t.Fatalf("TOML does not parse: %v\n%s", err, f.read(pt))
	}
	if got := doc["mcp_servers"]["ajiya"]["command"]; got != win {
		t.Errorf("TOML command = %v, want %s", got, win)
	}
	// the entry is recognised as Ajiya's by its file name, and a missing file
	// is reported as an entry that needs fixing
	if st, _ := f.env.Status(Codex, User); st != Different {
		t.Errorf("status of a path that does not exist here = %v", st)
	}
	if classify(entry{`C:\x\AJIYA.EXE`, []string{"mcp"}}) != Different || classify(entry{`C:\x\other.exe`, []string{"mcp"}}) != Different {
		t.Error("classify accepted a path that does not exist or another program")
	}
}

const otherTOML = `# my codex settings
model = "gpt-5"   # keep this comment

[mcp_servers.context7]
command = "npx"
args = ["-y", "@upstash/context7-mcp"]

[mcp_servers.context7.env]
KEY = "value"

[projects."/work/app"]
trust_level = "trusted"
`

func TestCodexConfig(t *testing.T) {
	for _, s := range []Scope{User, Project} {
		t.Run(string(s), func(t *testing.T) {
			f := newFixture(t)
			p := f.env.ConfigPath(Codex, s)
			if s == User && p != filepath.Join(f.env.CodexHome, "config.toml") || s == Project && p != filepath.Join(f.env.Dir, ".codex", "config.toml") {
				t.Fatalf("ConfigPath = %s", p)
			}
			f.apply(Codex, s, true)
			en, ok, err := tomlLookup(p, []byte(f.read(p)))
			if err != nil || !ok || en.Command != f.env.Binary || !reflect.DeepEqual(en.Args, []string{"mcp"}) {
				t.Fatalf("entry = %v %v %v\n%s", en, ok, err, f.read(p))
			}
			before := f.read(p)
			if c := f.apply(Codex, s, true); !c.Noop() || f.read(p) != before {
				t.Error("second install changed something")
			}
			f.apply(Codex, s, false)
			if got := strings.TrimSpace(f.read(p)); got != "" {
				t.Errorf("uninstall left %q in a file it created", got)
			}

			// an existing file: everything else, comments included, is kept byte for byte
			f.write(p, otherTOML)
			f.apply(Codex, s, true)
			got := f.read(p)
			if !strings.HasPrefix(got, otherTOML) {
				t.Errorf("existing text changed:\n%s", got)
			}
			if f.read(p+BackupSuffix) != otherTOML {
				t.Error("backup is not the original")
			}
			if !strings.Contains(got, "# keep this comment") || !strings.Contains(got, "[mcp_servers.ajiya]") {
				t.Errorf("missing pieces:\n%s", got)
			}
			f.apply(Codex, s, false)
			if got := f.read(p); got != otherTOML {
				t.Errorf("uninstall did not restore the file:\n%s", got)
			}
		})
	}
}

func TestCodexUninstallTakesSubtablesAndKeepsWhatFollows(t *testing.T) {
	f := newFixture(t)
	p := f.env.ConfigPath(Codex, User)
	f.write(p, `# top
[mcp_servers.ajiya]
command = "/old/ajiya"
args = ["mcp"]

[mcp_servers.ajiya.env]
A = "1"

# about b
[mcp_servers.b]
command = "b"
`)
	f.apply(Codex, User, false)
	want := "# top\n# about b\n[mcp_servers.b]\ncommand = \"b\"\n"
	if got := f.read(p); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestCodexStaleEntryIsReplaced(t *testing.T) {
	f := newFixture(t)
	p := f.env.ConfigPath(Codex, User)
	f.write(p, "[mcp_servers.ajiya]\ncommand = \"/gone/ajiya\"\nargs = [\"mcp\"]\n\n[other]\nx = 1\n")
	if c := f.apply(Codex, User, true); c.Action != "update" {
		t.Fatalf("action = %q", c.Action)
	}
	en, _, _ := tomlLookup(p, []byte(f.read(p)))
	if en.Command != f.env.Binary || !strings.Contains(f.read(p), "[other]\nx = 1") {
		t.Errorf("got:\n%s", f.read(p))
	}
}

func TestCodexCRLFIsKept(t *testing.T) {
	f := newFixture(t)
	p := f.env.ConfigPath(Codex, User)
	f.write(p, "model = \"x\"\r\n")
	f.apply(Codex, User, true)
	if strings.Contains(strings.ReplaceAll(f.read(p), "\r\n", ""), "\n") {
		t.Errorf("mixed line endings: %q", f.read(p))
	}
}

func TestCodexUneditableFormIsRefused(t *testing.T) {
	// dotted keys put the entry where a textual edit cannot find it
	f := newFixture(t)
	p := f.env.ConfigPath(Codex, User)
	const text = "mcp_servers.ajiya.command = \"/x/ajiya\"\nmcp_servers.ajiya.args = [\"mcp\"]\n[other]\nx = 1\n"
	f.write(p, text)
	if _, err := f.env.PlanUninstall(Codex, User); err == nil || !strings.Contains(err.Error(), "Nothing was changed") {
		t.Errorf("err = %v", err)
	}
	if f.read(p) != text {
		t.Error("file changed")
	}
}

func TestMalformedTOMLIsRefused(t *testing.T) {
	for name, text := range map[string]string{
		"syntax":          "[mcp_servers.x\ncommand = 1",
		"bad value":       "model = \n",
		"servers a value": "mcp_servers = 5\n",
	} {
		for _, install := range []bool{true, false} {
			f := newFixture(t)
			p := f.env.ConfigPath(Codex, User)
			f.write(p, text)
			plan := f.env.PlanInstall
			if !install {
				plan = f.env.PlanUninstall
			}
			_, err := plan(Codex, User)
			if err == nil || !strings.Contains(err.Error(), "Nothing was changed") || !strings.Contains(err.Error(), p) {
				t.Errorf("%s install=%v: err = %v", name, install, err)
			}
			if f.read(p) != text {
				t.Errorf("%s: file changed", name)
			}
			if _, err := os.Stat(p + BackupSuffix); err == nil {
				t.Errorf("%s: backup written", name)
			}
		}
	}
}

// Claude Code is configured through its own command when it is on PATH.
func TestClaudeUsesItsOwnCommand(t *testing.T) {
	for _, s := range []Scope{User, Project} {
		t.Run(string(s), func(t *testing.T) {
			f := newFixture(t)
			f.path["claude"] = true
			p := f.env.ConfigPath(Claude, s)
			f.write(p, otherJSON) // Claude's file: kept as the backup source

			c := f.apply(Claude, s, true)
			bin := f.env.Binary
			want := [][]string{{f.env.Dir, "claude", "mcp", "add", "--scope", string(s), "ajiya", "--", bin, "mcp"}}
			if !reflect.DeepEqual(f.calls, want) {
				t.Errorf("calls = %v\nwant    %v", f.calls, want)
			}
			if f.read(p) != otherJSON {
				t.Error("the file was edited directly although the command exists")
			}
			if f.read(p+BackupSuffix) != otherJSON {
				t.Error("no backup before the command ran")
			}
			if len(c.Cmds) != 1 || !strings.HasPrefix(c.Preview[0], "$ claude mcp add") {
				t.Errorf("preview = %v", c.Preview)
			}

			// pretend the command did its work: nothing more to do
			f.write(p, `{"mcpServers":{"ajiya":{"type":"stdio","command":`+string(jsonString(bin))+`,"args":["mcp"]}}}`)
			f.calls = nil
			if c := f.apply(Claude, s, true); !c.Noop() || len(f.calls) != 0 {
				t.Errorf("second install ran %v", f.calls)
			}

			f.apply(Claude, s, false)
			want = [][]string{{f.env.Dir, "claude", "mcp", "remove", "ajiya", "--scope", string(s)}}
			if !reflect.DeepEqual(f.calls, want) {
				t.Errorf("uninstall calls = %v", f.calls)
			}
		})
	}
}

func TestClaudeStaleEntryIsRemovedThenAdded(t *testing.T) {
	f := newFixture(t)
	f.path["claude"] = true
	f.write(f.env.ConfigPath(Claude, User), `{"mcpServers":{"ajiya":{"command":"/gone/ajiya","args":["mcp"]}}}`)
	f.apply(Claude, User, true)
	if len(f.calls) != 2 || f.calls[0][2] != "mcp" || f.calls[0][3] != "remove" || f.calls[1][3] != "add" {
		t.Errorf("calls = %v", f.calls)
	}
}

func TestClaudeCommandFailureIsReported(t *testing.T) {
	f := newFixture(t)
	f.path["claude"] = true
	f.env.Run = func(dir, name string, args ...string) (string, error) { return "boom", errors.New("exit status 1") }
	c, err := f.env.PlanInstall(Claude, User)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Apply(); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("err = %v", err)
	}
}

func TestClaudeMalformedFileIsRefusedEvenWithItsCommand(t *testing.T) {
	f := newFixture(t)
	f.path["claude"] = true
	p := f.env.ConfigPath(Claude, User)
	f.write(p, "{oops")
	if _, err := f.env.PlanInstall(Claude, User); err == nil || !strings.Contains(err.Error(), "Nothing was changed") {
		t.Errorf("err = %v", err)
	}
	if len(f.calls) != 0 {
		t.Errorf("ran %v", f.calls)
	}
}

func TestDetect(t *testing.T) {
	f := newFixture(t)
	if got := f.env.Detect(); len(got) != 0 {
		t.Errorf("found %v on an empty machine", got)
	}
	f.path["codex"] = true
	f.write(filepath.Join(f.env.Home, ".cursor", "mcp.json"), "{}")
	if got := f.env.Detect(); !reflect.DeepEqual(got, []Agent{Codex, Cursor}) {
		t.Errorf("found %v", got)
	}
	f.write(filepath.Join(f.env.Home, ".claude.json"), "{}")
	if got := f.env.Detect(); !reflect.DeepEqual(got, []Agent{Claude, Codex, Cursor}) {
		t.Errorf("found %v", got)
	}
}

func TestResolveBinary(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "ajiya")
	if err := os.WriteFile(real, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	exe := func() (string, error) { return real, nil }
	onPath := func(string) (string, error) { return real, nil }
	off := func(string) (string, error) { return "", errors.New("no") }
	// the running program is the one on PATH: its absolute path
	if got := ResolveBinary(onPath, exe); got != real {
		t.Errorf("on PATH: %s", got)
	}
	// a temporary build that is not on PATH: the plain name
	if got := ResolveBinary(off, exe); got != "ajiya" {
		t.Errorf("temporary: %s", got)
	}
	if got := ResolveBinary(off, func() (string, error) { return "", errors.New("no") }); got != "ajiya" {
		t.Errorf("no executable: %s", got)
	}
	if !isTemp(filepath.Join(os.TempDir(), "x", "ajiya")) || !isTemp("/var/x/go-build123/b001/ajiya") || isTemp("/usr/local/bin/ajiya") {
		t.Error("isTemp is wrong")
	}
}
