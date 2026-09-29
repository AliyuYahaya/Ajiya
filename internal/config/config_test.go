package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const valid = `
[project]
name = "Clinic Booking"
prefix = "CB"

[launch]
target = "go-live"

[test]
command = "go test ./..."

[[apps]]
name = "api"
path = "apps/api"

[[apps]]
name = "ui"
path = "packages/ui"
kind = "library"
`

func TestParseValid(t *testing.T) {
	c, err := Parse([]byte(valid))
	if err != nil {
		t.Fatal(err)
	}
	if c.Project.Name != "Clinic Booking" || c.Project.Prefix != "CB" || c.Launch.Target != "go-live" ||
		c.Test.Command != "go test ./..." || len(c.Apps) != 2 || c.Apps[1].Kind != "library" {
		t.Errorf("unexpected config: %+v", c)
	}
	for _, name := range []string{"api", "ui", "infra"} {
		if !c.HasApp(name) {
			t.Errorf("HasApp(%q) = false", name)
		}
	}
	if c.HasApp("web") {
		t.Error("HasApp(web) = true")
	}
	if got := strings.Join(c.AppNames(), ","); got != "api,infra,ui" {
		t.Errorf("AppNames = %s", got)
	}
}

func TestParseErrors(t *testing.T) {
	const project = "[project]\nname = \"X\"\nprefix = \"CB\"\n"
	app := func(body string) string { return project + "[[apps]]\n" + body }
	tests := []struct {
		name, in, want string
	}{
		{"bad toml", "[project\n", "ajiya.toml: "},
		{"unknown key", "[project]\nname = \"X\"\nprefix = \"CB\"\nnmae = \"Y\"\n", `unknown key "project.nmae"`},
		{"no name", "[project]\nprefix = \"CB\"\n", "name is missing"},
		{"lower prefix", "[project]\nname = \"X\"\nprefix = \"cb\"\n", `prefix "cb" must be`},
		{"digit prefix", "[project]\nname = \"X\"\nprefix = \"1B\"\n", `prefix "1B" must be`},
		{"bad app name", app("name = \"Api\"\npath = \"a\"\n"), `name "Api" must be`},
		{"infra reserved", app("name = \"infra\"\npath = \"a\"\n"), "is reserved"},
		{"duplicate app", app("name = \"a\"\npath = \"a\"\n[[apps]]\nname = \"a\"\npath = \"b\"\n"), "registered twice"},
		{"bad kind", app("name = \"a\"\npath = \"a\"\nkind = \"lib\"\n"), `kind "lib" is not known`},
		{"no path", app("name = \"a\"\n"), "path is missing"},
		{"absolute path", app("name = \"a\"\npath = \"/srv/a\"\n"), "must be relative"},
		{"backslash path", app("name = \"a\"\npath = 'apps\\a'\n"), "must be relative"},
		{"unclean path", app("name = \"a\"\npath = \"apps/a/\"\n"), `write it as "apps/a"`},
		{"outside root", app("name = \"a\"\npath = \"../a\"\n"), "outside the project root"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.in))
			if err == nil {
				t.Fatalf("no error, want %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want it to contain %q", err, tt.want)
			}
			if strings.Contains(err.Error(), "\n") {
				t.Errorf("error spans several lines: %q", err)
			}
		})
	}
}

func TestFindAndLoad(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Find(sub); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Find without config: err = %v, want ErrNotFound", err)
	}
	if err := os.WriteFile(filepath.Join(root, FileName), []byte(Template("Demo", "DE")), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Find(sub)
	if err != nil {
		t.Fatal(err)
	}
	c, err := Load(got)
	if err != nil {
		t.Fatalf("template does not load: %v", err)
	}
	if c.Project.Prefix != "DE" || c.Root != got {
		t.Errorf("unexpected config: %+v", c)
	}
}

func TestSuggestPrefix(t *testing.T) {
	tests := map[string]string{
		"Clinic Booking":         "CB",
		"ajiya":                  "AJ",
		"x":                      "X",
		"my-big-new-shiny-thing": "MBNS",
		"2048 game":              "T2G",
		"":                       "T",
	}
	for in, want := range tests {
		if got := SuggestPrefix(in); got != want {
			t.Errorf("SuggestPrefix(%q) = %q, want %q", in, got, want)
		}
		if _, err := Parse([]byte(Template("n", SuggestPrefix(in)))); err != nil {
			t.Errorf("suggested prefix for %q is invalid: %v", in, err)
		}
	}
}

func TestMilestones(t *testing.T) {
	const head = "[project]\nname = \"X\"\nprefix = \"X\"\n"
	c, err := Parse([]byte(head + `
[[milestones]]
name = "staging-proven"
targets = ["gate-0"]

[[milestones]]
name = "launch"
targets = ["go-live", "X-0142"]
`))
	if err != nil {
		t.Fatal(err)
	}
	ms := c.MilestoneList()
	if len(ms) != 2 || ms[0].Name != "staging-proven" || ms[1].Targets[1] != "X-0142" {
		t.Errorf("MilestoneList = %+v", ms)
	}
	old, err := Parse([]byte(head + "[launch]\ntarget = \"go-live\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if ms := old.MilestoneList(); len(ms) != 1 || ms[0].Name != LaunchMilestone || ms[0].Targets[0] != "go-live" {
		t.Errorf("old [launch] form = %+v", ms)
	}
	for name, tt := range map[string]struct{ in, want string }{
		"both forms":   {"[launch]\ntarget = \"a\"\n[[milestones]]\nname = \"b\"\ntargets = [\"b\"]\n", "move the launch target into the list"},
		"bad name":     {"[[milestones]]\nname = \"Gate 0\"\ntargets = [\"a\"]\n", `name "Gate 0" must be lower case`},
		"duplicate":    {"[[milestones]]\nname = \"a\"\ntargets = [\"a\"]\n[[milestones]]\nname = \"a\"\ntargets = [\"b\"]\n", "listed twice"},
		"no targets":   {"[[milestones]]\nname = \"a\"\ntargets = []\n", "has no targets"},
		"empty target": {"[[milestones]]\nname = \"a\"\ntargets = [\" \"]\n", "empty target"},
	} {
		if _, err := Parse([]byte(head + tt.in)); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: err = %v, want %q", name, err, tt.want)
		}
	}
}
