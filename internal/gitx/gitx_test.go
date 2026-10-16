package gitx

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
	run(t, dir, "init", "-q", "-b", "main")
	return dir
}

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

var messages = []struct {
	name string
	msg  string
	want Refs
}{
	{"one", "Add slots\n\nAjiya: AB-0001\n", Refs{IDs: []string{"AB-0001"}}},
	{"two with other trailer", "Add\n\nAjiya: AB-0001, AB-0002\nSigned-off-by: T <t@example.com>\n", Refs{IDs: []string{"AB-0001", "AB-0002"}}},
	{"repeated key", "Add\n\nAjiya: AB-0002\nAjiya: AB-0001 AB-0002\n", Refs{IDs: []string{"AB-0002", "AB-0001"}}},
	{"key case", "Fix\n\najiya: AB-0003\n", Refs{IDs: []string{"AB-0003"}}},
	{"folded value", "Add\n\nAjiya: AB-0001,\n  AB-0002\n", Refs{IDs: []string{"AB-0001", "AB-0002"}}},
	{"chore", "Bump deps\n\nAjiya: chore\n", Refs{Chore: true}},
	{"not the last paragraph", "Add\n\nAjiya: AB-0001\n\nMore text.\n", Refs{}},
	{"mixed with prose", "Add\n\nSee the notes.\nAjiya: AB-0001\n", Refs{}},
	{"subject only", "Ajiya: AB-0001\n", Refs{}},
	{"no trailer", "Just a change\n", Refs{}},
	{"revert", "Revert \"Add slots\"\n\nThis reverts commit 4f2a91c.\n", Refs{Revert: true}},
	{"reapply", "Reapply \"Add slots\"\n\nAjiya: AB-0001\n", Refs{IDs: []string{"AB-0001"}, Revert: true}},
}

// ParseMessage and Log must read every message the same way.
func TestMessageAndHistoryAgree(t *testing.T) {
	dir := newRepo(t)
	for _, m := range messages {
		got, err := ParseMessage(dir, []byte(m.msg))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, m.want) {
			t.Errorf("%s: ParseMessage = %+v, want %+v", m.name, got, m.want)
		}
		f := filepath.Join(t.TempDir(), "msg")
		os.WriteFile(f, []byte(m.msg), 0o644)
		run(t, dir, "commit", "-q", "--allow-empty", "-F", f)
	}
	log, err := Log(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(log) != len(messages) {
		t.Fatalf("log has %d commits, want %d", len(log), len(messages))
	}
	for i, m := range messages {
		c := log[len(log)-1-i] // log is newest first
		if !reflect.DeepEqual(c.Refs, m.want) {
			t.Errorf("%s: Log = %+v, want %+v", m.name, c.Refs, m.want)
		}
		if c.Merge || len(c.Hash) != 40 || len(c.Date) != 10 {
			t.Errorf("%s: commit = %+v", m.name, c)
		}
		if want := m.want.Chore || m.want.Revert; c.Exempt() != want {
			t.Errorf("%s: Exempt = %v, want %v", m.name, c.Exempt(), want)
		}
	}
}

func TestMergeAndReferencing(t *testing.T) {
	dir := newRepo(t)
	if log, err := Log(dir); err != nil || log != nil {
		t.Fatalf("empty repo: %v, %v", log, err)
	}
	run(t, dir, "commit", "-q", "--allow-empty", "-m", "Base", "-m", "Ajiya: AB-0001")
	run(t, dir, "checkout", "-q", "-b", "side")
	run(t, dir, "commit", "-q", "--allow-empty", "-m", "Side", "-m", "Ajiya: AB-0002")
	run(t, dir, "checkout", "-q", "main")
	run(t, dir, "commit", "-q", "--allow-empty", "-m", "Main", "-m", "Ajiya: AB-0001")
	run(t, dir, "merge", "-q", "--no-ff", "--no-edit", "side")

	log, err := Log(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !log[0].Merge || !log[0].Exempt() || log[1].Merge {
		t.Errorf("merge not detected: %+v", log[:2])
	}
	refs, err := Referencing(dir, "AB-0001")
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 2 || refs[0].Subject != "Main" || refs[1].Subject != "Base" {
		t.Errorf("Referencing = %+v", refs)
	}
}

func TestNotARepo(t *testing.T) {
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(t.TempDir()))
	if _, err := Log(t.TempDir()); err != ErrNotRepo {
		t.Errorf("err = %v, want ErrNotRepo", err)
	}
}

func TestReposAndReferencingIn(t *testing.T) {
	root := t.TempDir() // not a repository
	t.Setenv("HOME", root)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"AUTHOR", "COMMITTER"} {
		t.Setenv("GIT_"+k+"_NAME", "T")
		t.Setenv("GIT_"+k+"_EMAIL", "t@example.com")
	}
	a, b, plain := filepath.Join(root, "a"), filepath.Join(root, "b"), filepath.Join(root, "plain")
	for _, d := range []string{filepath.Join(a, "sub"), b, plain} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	run(t, a, "init", "-q", "-b", "main")
	run(t, b, "init", "-q", "-b", "main")
	run(t, a, "commit", "-q", "--allow-empty", "-m", "In a", "-m", "Ajiya: AB-0001")
	run(t, b, "commit", "-q", "--allow-empty", "-m", "In b", "-m", "Ajiya: AB-0001, AB-0002")

	// root and plain are not repositories, a/sub repeats a, "gone" is missing.
	repos := Repos(root, []string{a, filepath.Join(a, "sub"), b, plain, filepath.Join(root, "gone"), b})
	if len(repos) != 2 {
		t.Fatalf("Repos = %v, want a and b once each", repos)
	}
	got, err := ReferencingIn(repos, "AB-0001")
	if err != nil || len(got) != 2 {
		t.Fatalf("ReferencingIn = %+v, %v", got, err)
	}
	if got[0].Repo == got[1].Repo {
		t.Errorf("both commits attributed to %s", got[0].Repo)
	}
	if one, _ := ReferencingIn(repos, "AB-0002"); len(one) != 1 || one[0].Subject != "In b" || filepath.Base(one[0].Repo) != "b" {
		t.Errorf("AB-0002 = %+v", one)
	}
	if none, err := ReferencingIn(repos, "AB-0009"); err != nil || len(none) != 0 {
		t.Errorf("AB-0009 = %+v, %v", none, err)
	}
	if _, err := ReferencingIn(Repos(root, []string{plain}), "AB-0001"); err != ErrNotRepo {
		t.Errorf("no repositories: err = %v, want ErrNotRepo", err)
	}
	// A root that is itself a repository is searched too, once.
	if r := Repos(a, []string{a, filepath.Join(a, "sub")}); len(r) != 1 {
		t.Errorf("Repos(a) = %v", r)
	}
}
