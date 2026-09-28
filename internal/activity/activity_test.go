package activity

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// newRepo makes an empty repository with a fixed author and no user config.
func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_AUTHOR_NAME", "T")
	t.Setenv("GIT_AUTHOR_EMAIL", "t@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "T")
	t.Setenv("GIT_COMMITTER_EMAIL", "t@example.com")
	run(t, dir, "", "init", "-q", "-b", "main")
	return dir
}

// run runs git in dir; a non-empty date (RFC 3339) sets the author and
// committer dates.
func run(t *testing.T, dir, date string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	if date != "" {
		cmd.Env = append(cmd.Env, "GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// commit writes files and commits them all with msg at date; it returns the hash.
func commit(t *testing.T, dir, date, msg string, files map[string]string) string {
	t.Helper()
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run(t, dir, "", "add", "-A")
	run(t, dir, date, "commit", "-q", "--allow-empty", "-m", msg)
	return run(t, dir, "", "rev-parse", "HEAD")
}

// phase is a phase file holding rows "ID|Status".
func phase(title string, rows ...string) string {
	s := "# " + title + "\n\nGoal: g\n\n| ID | App | Ticket | Done when | Depends | Status |\n|---|---|---|---|---|---|\n"
	for _, r := range rows {
		id, st, _ := strings.Cut(r, "|")
		s += "| " + id + " | web | T | D | - | " + st + " |\n"
	}
	return s
}

// want is the part of an Entry the tests check; Hash, Short and Time are
// checked separately.
type want struct {
	Date, Subject string
	Refs          []string
	Chore, Merge  bool
	Revert        bool
	Agent         string
	Changes       []Change
}

func TestRecent(t *testing.T) {
	dir := newRepo(t)
	// The newest commit is 2026-03-01T12:00:00Z, so a 30-day window starts at
	// 2026-01-30T12:00:00Z whatever today's date is.
	commit(t, dir, "2026-01-30T11:59:59Z", "Too old\n\nAjiya: AB-0009", map[string]string{"README": "x\n"})
	commit(t, dir, "2026-01-30T12:00:00Z", "Plan\n\nAjiya: chore", map[string]string{
		"ajiya/core.md":   phase("Core", "AB-0001|🟥 Pending", "AB-0002|🟥 Pending"),
		"ajiya/broken.md": "not a phase\n",
		"ajiya/README.md": "notes\n",
	})
	commit(t, dir, "2026-02-10T12:00:00Z", "Start slots\n\nAjiya: AB-0001\nCo-Authored-By: Claude Opus <noreply@anthropic.com>", map[string]string{
		"ajiya/core.md": phase("Core", "AB-0001|🟨 In progress", "AB-0002|🟥 Pending"),
	})
	run(t, dir, "", "checkout", "-q", "-b", "feat")
	side := commit(t, dir, "2026-02-15T12:00:00Z", "Side work\n\nAjiya: AB-0002\nCo-authored-by: Jane Doe <jane@example.com>\nCo-authored-by: Ada Amplify <claude@example.com>", map[string]string{"side.txt": "x\n"})
	run(t, dir, "", "checkout", "-q", "main")
	commit(t, dir, "2026-02-20T12:00:00Z", "Finish slots\n\nAjiya: AB-0001, AB-0002\nCo-authored-by: GitHub Copilot <copilot@github.com>", map[string]string{
		"ajiya/core.md": phase("Core", "AB-0001|🟩 Done · abcdef1 · 2026-02-20", "AB-0002|🟥 Pending"),
		"ajiya/next.md": phase("Next", "AB-0003|🟥 Pending"),
	})
	run(t, dir, "2026-02-25T12:00:00Z", "merge", "-q", "--no-ff", "-m", "Merge feat", "feat")
	run(t, dir, "2026-03-01T12:00:00Z", "revert", "--no-edit", side)

	wants := []want{
		{Date: "2026-03-01", Subject: `Revert "Side work"`, Refs: []string{}, Revert: true, Changes: []Change{}},
		{Date: "2026-02-25", Subject: "Merge feat", Refs: []string{}, Merge: true, Changes: []Change{}},
		{Date: "2026-02-20", Subject: "Finish slots", Refs: []string{"AB-0001", "AB-0002"}, Agent: "Copilot", Changes: []Change{
			{ID: "AB-0001", Phase: "core", From: "in_progress", To: "done", Text: "🟩 Done · abcdef1 · 2026-02-20"},
			{ID: "AB-0003", Phase: "next", From: "", To: "pending", Text: "🟥 Pending"},
		}},
		{Date: "2026-02-15", Subject: "Side work", Refs: []string{"AB-0002"}, Changes: []Change{}},
		{Date: "2026-02-10", Subject: "Start slots", Refs: []string{"AB-0001"}, Agent: "Claude", Changes: []Change{
			{ID: "AB-0001", Phase: "core", From: "pending", To: "in_progress", Text: "🟨 In progress"},
		}},
		{Date: "2026-01-30", Subject: "Plan", Refs: []string{}, Chore: true, Changes: []Change{
			{ID: "AB-0001", Phase: "core", To: "pending", Text: "🟥 Pending"},
			{ID: "AB-0002", Phase: "core", To: "pending", Text: "🟥 Pending"},
		}},
	}

	got, err := Recent(dir, 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(wants) {
		t.Fatalf("got %d entries, want %d: %+v", len(got), len(wants), got)
	}
	for i, w := range wants {
		e := got[i]
		g := want{e.Date, e.Subject, e.Refs, e.Chore, e.Merge, e.Revert, e.Agent, e.Changes}
		if !reflect.DeepEqual(g, w) {
			t.Errorf("entry %d:\n got %+v\nwant %+v", i, g, w)
		}
		if len(e.Hash) != 40 || e.Short != e.Hash[:7] || e.Author != "T" {
			t.Errorf("entry %d: hash %q short %q author %q", i, e.Hash, e.Short, e.Author)
		}
	}

	// A shorter window, still measured from the newest commit.
	short, err := Recent(dir, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(short) != 3 || short[2].Subject != "Finish slots" {
		t.Errorf("10-day window: got %d entries, want the 3 since 2026-02-19", len(short))
	}

	// Same repository, same answer.
	again, err := Recent(dir, 30)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, again) {
		t.Error("two runs differ")
	}
}

// Commits at the same second are ordered by hash.
func TestRecentTies(t *testing.T) {
	dir := newRepo(t)
	for _, s := range []string{"a", "b", "c", "d"} {
		commit(t, dir, "2026-05-01T08:00:00Z", s+"\n\nAjiya: chore", nil)
	}
	got, err := Recent(dir, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("got %d entries, want 4", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].Hash > got[i].Hash {
			t.Errorf("entries %d and %d not ordered by hash", i-1, i)
		}
	}
}

func TestRecentEmpty(t *testing.T) {
	got, err := Recent(newRepo(t), 30)
	if err != nil || len(got) != 0 {
		t.Errorf("empty repository: got %v, %v; want no entries", got, err)
	}
}

func TestAgent(t *testing.T) {
	tests := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{"Claude <noreply@anthropic.com>"}, "Claude"},
		{[]string{"claude-code <x@y.z>"}, "Claude"},
		{[]string{"OpenAI CODEX <codex@openai.com>"}, "Codex"},
		{[]string{"Jane Doe <jane@example.com>", "Cursor Agent <cursoragent@cursor.com>"}, "Cursor"},
		{[]string{"Ada Amplify <amp@example.com>"}, ""}, // the email is not the name
		{[]string{"Jane Doe <jane@example.com>"}, ""},
		{[]string{"Sourcegraph Amp <amp@ampcode.com>"}, "Amp"},
		{[]string{"Gemini <g@example.com>", "Claude <c@example.com>"}, "Gemini"},
	}
	for _, tt := range tests {
		if got := Agent(tt.in); got != tt.want {
			t.Errorf("Agent(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
