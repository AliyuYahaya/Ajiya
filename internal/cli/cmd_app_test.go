package cli

import (
	"strings"
	"testing"
)

// The registered path is quoted so a path of . does not read as "..".
func TestAppAddQuotesPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	gitIn(t, dir, "init", "-q", "-b", "main")
	ajiya(t, dir, 0, "init", "--name", "Demo", "--prefix", "DE")

	if out := ajiya(t, dir, 0, "app", "add", "cli", "--path", "."); !strings.Contains(string(out), "Registered app cli at `.`.") {
		t.Errorf("app add --path . printed %q", out)
	}
	if out := ajiya(t, dir, 0, "app", "add", "web", "--path", "apps/web"); !strings.Contains(string(out), "Registered app web at `apps/web`.") {
		t.Errorf("app add nested path printed %q", out)
	}
}
