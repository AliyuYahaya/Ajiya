package cli

import (
	"bytes"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/AliyuYahaya/Ajiya/internal/gitx"
)

var update = flag.Bool("update", false, "rewrite the golden files")

var statusNow = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

func gitAt(t *testing.T, dir, date string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// statusRun runs ajiya in dir with the clock fixed at statusNow.
func statusRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	e := &env{stdout: &out, stderr: &out, dir: dir, now: func() time.Time { return statusNow }}
	if code := run(e, args); code != 0 {
		t.Fatalf("ajiya %v: exit %d\n%s", args, code, out.String())
	}
	return out.String()
}

// statusFixture builds a clean repository with fixed commit dates: two
// milestones and four tickets in progress or waiting on a person. DE-0002 has a
// start record (5 hours ago), DE-0003 only a commit that started it (2 days
// ago), and DE-0004 neither: its commit is older than the activity window.
func statusFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"AUTHOR", "COMMITTER"} {
		t.Setenv("GIT_"+k+"_NAME", "Test Author")
		t.Setenv("GIT_"+k+"_EMAIL", "author@example.com")
	}
	steps := func(list ...[]string) {
		for _, args := range list {
			statusRun(t, dir, args...)
		}
	}
	gitAt(t, dir, "2026-01-01T10:00:00Z", "init", "-q", "-b", "main")
	steps(
		[]string{"init", "--name", "Demo", "--prefix", "DE"},
		[]string{"phase", "add", "gate", "Staging proven"},
		[]string{"phase", "add", "api", "API works"},
		[]string{"phase", "add", "go-live", "Live"},
		[]string{"ticket", "add", "--phase", "gate", "--app", "infra", "--done-when", "x", "Staging host"},
		[]string{"ticket", "add", "--phase", "api", "--app", "infra", "--done-when", "x", "Schema"},
		[]string{"ticket", "add", "--phase", "api", "--app", "infra", "--done-when", "x", "Endpoint", "--depends", "DE-0002"},
		[]string{"ticket", "add", "--phase", "api", "--app", "infra", "--done-when", "x", "Auth"},
		[]string{"ticket", "start", "DE-0004"},
	)
	gitAt(t, dir, "2026-01-01T10:00:00Z", "add", "-A")
	gitAt(t, dir, "2026-01-01T10:00:00Z", "commit", "-q", "-m", "Plan the project\n\nAjiya: chore")

	steps(
		[]string{"ticket", "add", "--phase", "go-live", "--app", "infra", "--done-when", "x", "Deploy", "--depends", "DE-0001,DE-0003", "--human", "needs the production keys"},
		[]string{"ticket", "add", "--phase", "go-live", "--app", "infra", "--done-when", "x", "Docs"},
		[]string{"milestone", "add", "staging", "--targets", "gate"},
		[]string{"milestone", "add", "launch", "--targets", "go-live,api"},
		[]string{"ticket", "start", "DE-0003"},
	)
	gitAt(t, dir, "2026-03-08T12:00:00Z", "add", "-A")
	gitAt(t, dir, "2026-03-08T12:00:00Z", "commit", "-q", "-m", "Start the endpoint work\n\nAjiya: DE-0003")

	steps([]string{"ticket", "start", "DE-0002"})
	gitAt(t, dir, "2026-03-09T09:00:00Z", "commit", "-q", "-am", "Start the schema work\n\nAjiya: DE-0002")

	// ticket start recorded the real time; replace it with fixed records.
	for _, id := range []string{"DE-0002", "DE-0003", "DE-0004"} {
		if err := gitx.ClearStart(dir, id); err != nil {
			t.Fatal(err)
		}
	}
	if err := gitx.RecordStart(dir, "DE-0002", statusNow.Add(-5*time.Hour).Unix()); err != nil {
		t.Fatal(err)
	}
	return dir
}

var shortHash = regexp.MustCompile(`"short": "[0-9a-f]+"`)

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("%s differs (run go test ./internal/cli -run Status -update):\n--- got\n%s\n--- want\n%s", name, got, want)
	}
}

func TestStatusGolden(t *testing.T) {
	dir := statusFixture(t)
	golden(t, "status.golden", statusRun(t, dir, "status"))
	golden(t, "status-brief.golden", statusRun(t, dir, "status", "--brief"))
	js := shortHash.ReplaceAllString(statusRun(t, dir, "status", "--json"), `"short": "HASH"`)
	golden(t, "status-json.golden", js)
}

func TestStatusBriefIsThreeLines(t *testing.T) {
	dir := statusFixture(t)
	if n := strings.Count(statusRun(t, dir, "status", "--brief"), "\n"); n > 3 {
		t.Errorf("--brief printed %d lines", n)
	}
}

func TestStatusFlagsConflict(t *testing.T) {
	var out bytes.Buffer
	e := &env{stdout: &out, stderr: &out, dir: t.TempDir()}
	if code := run(e, []string{"status", "--brief", "--json"}); code != ExitUsage {
		t.Errorf("exit = %d, want %d", code, ExitUsage)
	}
}

func TestAgo(t *testing.T) {
	for secs, want := range map[int64]string{0: "just now", 59: "just now", 60: "1m ago", 3599: "59m ago", 3600: "1h ago", 47 * 3600: "47h ago", 48 * 3600: "2d ago", 9 * 86400: "9d ago"} {
		if got := ago(secs); got != want {
			t.Errorf("ago(%d) = %q, want %q", secs, got, want)
		}
	}
}
