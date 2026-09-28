package config

import (
	"strings"
	"testing"
)

func TestSetLaunchText(t *testing.T) {
	const head = "[project]\nname = \"X\"\nprefix = \"X\"\n"
	tests := []struct{ name, in, want string }{
		{"replace, keeping comment",
			head + "\n[launch]\ntarget = 'old'  # the go-live phase\n",
			head + "\n[launch]\ntarget = \"v0-1\"  # the go-live phase\n"},
		{"table without target",
			head + "\n[launch]\n\n[test]\ncommand = \"go test\"\n",
			head + "\n[launch]\ntarget = \"v0-1\"\n\n[test]\ncommand = \"go test\"\n"},
		{"no table, apps follow",
			head + "\n[[apps]]\nname = \"a\"\npath = \"a\"\n",
			head + "\n[launch]\ntarget = \"v0-1\"\n\n[[apps]]\nname = \"a\"\npath = \"a\"\n"},
		{"no table, nothing follows",
			head,
			head + "\n[launch]\ntarget = \"v0-1\"\n"},
		{"commented template is left alone",
			head + "\n# [launch]\n# target = \"go-live\"\n",
			head + "\n# [launch]\n# target = \"go-live\"\n\n[launch]\ntarget = \"v0-1\"\n"},
		{"target in another table is not touched",
			head + "\n[test]\ntarget = \"x\"\n",
			head + "\n[test]\ntarget = \"x\"\n\n[launch]\ntarget = \"v0-1\"\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SetLaunchText(tt.in, "v0-1")
			if got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestAppText(t *testing.T) {
	const in = `[project]
name = "X"
prefix = "X"

[[apps]]
name = "api"   # the backend
path = "apps/api"

[[apps]]
name = 'web'
path = "apps/web"

[test]
command = "go test"
`
	got, ok := SetAppPathText(in, "web", "web")
	if !ok || !strings.Contains(got, "name = 'web'\npath = \"web\"\n") || !strings.Contains(got, `path = "apps/api"`) {
		t.Errorf("SetAppPathText:\n%s", got)
	}
	got, ok = RemoveAppText(in, "api")
	want := strings.Replace(in, "\n[[apps]]\nname = \"api\"   # the backend\npath = \"apps/api\"\n", "", 1)
	if !ok || got != want {
		t.Errorf("RemoveAppText:\n%s\nwant\n%s", got, want)
	}
	got, _ = RemoveAppText(in, "web")
	if strings.Contains(got, "web") || !strings.Contains(got, "[test]") {
		t.Errorf("RemoveAppText web:\n%s", got)
	}
	if _, ok := RemoveAppText(in, "nope"); ok {
		t.Error("removed an app that is not there")
	}
	got = AddAppText(in, App{Name: "ui", Path: "packages/ui", Kind: "library"})
	c, err := Parse([]byte(got))
	if err != nil || len(c.Apps) != 3 || c.Apps[2].Kind != "library" {
		t.Errorf("AddAppText: %v\n%s", err, got)
	}
}
