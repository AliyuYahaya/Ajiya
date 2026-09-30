package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/AliyuYahaya/Ajiya/internal/agents"
)

// TestMain makes sure no test can read or change the real ~/.claude, ~/.codex
// or ~/.cursor (for example through 'init --yes', which registers Ajiya with the
// agents it finds): the home folder is a temporary one and no agent command is
// on the (fake) PATH.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "ajiya-cli-home")
	if err != nil {
		panic(err)
	}
	for _, k := range []string{"HOME", "USERPROFILE"} {
		os.Setenv(k, home)
	}
	os.Setenv("CODEX_HOME", filepath.Join(home, ".codex"))
	newAgentEnv = func(projectDir string) (*agents.Env, error) {
		return &agents.Env{
			Home: home, CodexHome: filepath.Join(home, ".codex"), Dir: projectDir, Binary: "ajiya",
			LookPath: func(string) (string, error) { return "", errors.New("not found") },
			Run: func(dir, name string, args ...string) (string, error) {
				return "", errors.New("tests must not run agent commands")
			},
		}, nil
	}
	code := m.Run()
	os.RemoveAll(home)
	os.Exit(code)
}
