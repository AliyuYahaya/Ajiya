package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// ajiya runs a command in dir and returns stdout; it fails the test on a
// non-zero exit unless want says otherwise.
func ajiya(t *testing.T, dir string, want int, args ...string) []byte {
	t.Helper()
	var out, errOut bytes.Buffer
	if code := run(&env{stdout: &out, stderr: &errOut, dir: dir}, args); code != want {
		t.Fatalf("ajiya %v: exit %d, want %d\n%s%s", args, code, want, out.String(), errOut.String())
	}
	return out.Bytes()
}

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func decode(t *testing.T, data []byte, v any) {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		t.Fatalf("decode: %v\n%s", err, data)
	}
}

type status struct {
	State, Text, Note, Human, Blocked, Commit, By, Date, Reason string
	TestsPassed                                              bool `json:"tests_passed"`
}

type ticket struct {
	ID, Phase, App, Title string
	DoneWhen              string `json:"done_when"`
	Depends               []string
	Status                status
}

func TestJSON(t *testing.T) {
	dir := t.TempDir()
	for k, v := range map[string]string{"HOME": dir, "GIT_CONFIG_NOSYSTEM": "1", "GIT_AUTHOR_NAME": "T",
		"GIT_AUTHOR_EMAIL": "t@example.com", "GIT_COMMITTER_NAME": "T", "GIT_COMMITTER_EMAIL": "t@example.com"} {
		t.Setenv(k, v)
	}
	gitIn(t, dir, "init", "-q", "-b", "main")
	ajiya(t, dir, 0, "init", "--name", "Demo", "--prefix", "DE")

	// Empty results are empty lists, not null.
	if got := string(ajiya(t, dir, 0, "check", "--json")); got != "[]\n" {
		t.Errorf("check --json on a clean plan = %q", got)
	}
	if got := string(ajiya(t, dir, 0, "next", "--json")); got != "[]\n" {
		t.Errorf("next --json with no tickets = %q", got)
	}
	var none struct {
		Target      *string
		Kind        string
		Required    int
		Closed      int
		Percent     int
		AfterLaunch int `json:"after_launch"`
		Open        []ticket
	}
	decode(t, ajiya(t, dir, 0, "launch", "show", "--json"), &none)
	if none.Target != nil || none.Open == nil {
		t.Errorf("launch show --json without a target = %+v", none)
	}

	ajiya(t, dir, 0, "phase", "add", "core", "Core works")
	ajiya(t, dir, 0, "ticket", "add", "--phase", "core", "--app", "infra", "--done-when", "x", "Deploy")
	ajiya(t, dir, 0, "ticket", "add", "--phase", "core", "--app", "infra", "--done-when", "y", "--depends", "DE-0001", "--human", "keys", "Go live")
	ajiya(t, dir, 0, "launch", "set", "DE-0002")
	os.WriteFile(filepath.Join(dir, "x.txt"), []byte("x"), 0o644)
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "Deploy", "-m", "Ajiya: DE-0001")
	ajiya(t, dir, 0, "ticket", "done", "DE-0001", "--note", "first try")

	var next []struct {
		ticket
		Launch   bool
		Unblocks int
	}
	decode(t, ajiya(t, dir, 0, "next", "--json"), &next)
	if len(next) != 1 || next[0].ID != "DE-0002" || !next[0].Launch || next[0].Status.Human != "keys" ||
		next[0].Status.State != "pending" || next[0].Depends[0] != "DE-0001" {
		t.Errorf("next --json = %+v", next)
	}

	var show struct {
		ticket
		PhaseTitle   string `json:"phase_title"`
		Dependencies []struct{ ID, Title, State string }
		Dependants   []struct{ ID, Title, State string }
		Commits      []struct{ Hash, Date, Subject string }
	}
	decode(t, ajiya(t, dir, 0, "ticket", "show", "DE-0001", "--json"), &show)
	if show.Status.State != "done" || show.Status.Note != "first try" || len(show.Status.Commit) != 7 ||
		show.PhaseTitle != "Core" || len(show.Dependencies) != 0 || show.Dependencies == nil ||
		len(show.Dependants) != 1 || show.Dependants[0].State != "pending" ||
		len(show.Commits) != 1 || show.Commits[0].Subject != "Deploy" || len(show.Commits[0].Hash) != 40 {
		t.Errorf("ticket show --json = %+v", show)
	}

	var launch = none
	decode(t, ajiya(t, dir, 0, "launch", "show", "--json"), &launch)
	if launch.Target == nil || *launch.Target != "DE-0002" || launch.Kind != "ticket" || launch.Required != 2 ||
		launch.Closed != 1 || launch.Percent != 50 || len(launch.Open) != 1 || launch.AfterLaunch != 0 {
		t.Errorf("launch show --json = %+v", launch)
	}

	// Findings, with the exit code unchanged.
	ajiya(t, dir, 0, "phase", "add", "empty", "Nothing yet")
	var findings []struct{ Code, Level, Location, Message, Fix string }
	decode(t, ajiya(t, dir, 1, "check", "--strict", "--json"), &findings)
	if len(findings) != 1 || findings[0].Code != "W001" || findings[0].Level != "warning" || findings[0].Location != "ajiya/empty.md:1" {
		t.Errorf("check --json = %+v", findings)
	}
}
