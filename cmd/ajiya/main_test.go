package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/AliyuYahaya/Ajiya/internal/cli"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"ajiya": func() { os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr)) },
	})
}

// TestScripts runs testdata/script/*.txtar. Each script starts in an empty git
// repository with a fixed author and no user or system git config.
func TestScripts(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
		Setup: func(e *testscript.Env) error {
			for k, v := range map[string]string{
				"HOME":                e.WorkDir,
				"GIT_CONFIG_NOSYSTEM": "1",
				"GIT_AUTHOR_NAME":     "Test Author",
				"GIT_AUTHOR_EMAIL":    "author@example.com",
				"GIT_COMMITTER_NAME":  "Test Author",
				"GIT_COMMITTER_EMAIL": "author@example.com",
			} {
				e.Setenv(k, v)
			}
			cmd := exec.Command("git", "init", "-q", "-b", "main")
			cmd.Dir = e.WorkDir
			cmd.Env = e.Vars
			return cmd.Run()
		},
	})
}
