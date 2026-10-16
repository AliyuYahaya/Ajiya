package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// nestedProjects builds an outer project (prefix OU, ticket OU-0001) in a git
// repository, with an inner project (prefix IN, ticket IN-0001) in its own git
// repository below it, and a plain folder "sub" in the outer repository that
// holds an ajiya.toml of its own (prefix SU). It returns the three folders.
func nestedProjects(t *testing.T) (outer, inner, sub string) {
	t.Helper()
	outer = t.TempDir()
	t.Setenv("HOME", outer)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	inner = filepath.Join(outer, "inner")
	sub = filepath.Join(outer, "sub")
	for _, d := range []string{inner, sub} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mrGit(t, outer, "init", "-q", "-b", "main")
	mrGit(t, inner, "init", "-q", "-b", "main")
	for dir, prefix := range map[string]string{outer: "OU", inner: "IN", sub: "SU"} {
		toml := "[project]\nname = \"" + prefix + "\"\nprefix = \"" + prefix + "\"\n"
		if err := os.WriteFile(filepath.Join(dir, "ajiya.toml"), []byte(toml), 0o644); err != nil {
			t.Fatal(err)
		}
		if out, code := mrRun(t, dir, "phase", "add", "build", "Build"); code != 0 {
			t.Fatalf("phase add in %s: %s", dir, out)
		}
		if out, code := mrRun(t, dir, "ticket", "add", "--phase", "build", "--app", "infra", "Work "+prefix, "--done-when", "it works"); code != 0 {
			t.Fatalf("ticket add in %s: %s", dir, out)
		}
	}
	return outer, inner, sub
}

func hookMsg(t *testing.T, text string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "MSG")
	if err := os.WriteFile(f, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return f
}

// A nested project commits from both levels: each repository's commit-msg hook
// checks trailers against the plan that owns that repository.
func TestNestedProjectHooksCheckOwningPlan(t *testing.T) {
	outer, inner, _ := nestedProjects(t)
	for _, c := range []struct {
		name, dir, trailer string
		ok                 bool
	}{
		{"outer commit names an outer ticket", outer, "OU-0001", true},
		{"outer commit names an inner ticket", outer, "IN-0001", false},
		{"inner commit names an inner ticket", inner, "IN-0001", true},
		{"inner commit names an outer ticket", inner, "OU-0001", false},
	} {
		out, code := mrRun(t, c.dir, "hook", "run", "commit-msg", hookMsg(t, "Work\n\nAjiya: "+c.trailer+"\n"))
		if (code == 0) != c.ok {
			t.Errorf("%s: exit %d, want ok=%v\n%s", c.name, code, c.ok, out)
		}
	}
}

// Run from a folder that has its own ajiya.toml but is not a repository, the
// hook still checks against the plan of the repository being committed to.
func TestNestedProjectHookFromSubfolderUsesRepositoryPlan(t *testing.T) {
	_, _, sub := nestedProjects(t)
	if out, code := mrRun(t, sub, "hook", "run", "commit-msg", hookMsg(t, "Work\n\nAjiya: OU-0001\n")); code != 0 {
		t.Errorf("outer ticket from sub should pass: exit %d\n%s", code, out)
	}
	out, code := mrRun(t, sub, "hook", "run", "commit-msg", hookMsg(t, "Work\n\nAjiya: SU-0001\n"))
	if code == 0 || !strings.Contains(out, "SU-0001") {
		t.Errorf("sub's own ticket must not satisfy the outer repository: exit %d\n%s", code, out)
	}
}

// GIT_DIR and GIT_WORK_TREE of the commit decide the owning plan, even when the
// hook's working folder is somewhere else.
func TestNestedProjectHookHonoursGitDir(t *testing.T) {
	outer, inner, _ := nestedProjects(t)
	t.Setenv("GIT_DIR", filepath.Join(outer, ".git"))
	t.Setenv("GIT_WORK_TREE", outer)
	if out, code := mrRun(t, inner, "hook", "run", "commit-msg", hookMsg(t, "Work\n\nAjiya: OU-0001\n")); code != 0 {
		t.Errorf("commit to the outer repository should use the outer plan: exit %d\n%s", code, out)
	}
	if _, code := mrRun(t, inner, "hook", "run", "commit-msg", hookMsg(t, "Work\n\nAjiya: IN-0001\n")); code == 0 {
		t.Error("an inner ticket must not pass for a commit to the outer repository")
	}
}

// prepare-commit-msg pre-fills from the owning plan's in-progress tickets.
func TestNestedProjectPrepareUsesOwningPlan(t *testing.T) {
	outer, inner, _ := nestedProjects(t)
	for dir, id := range map[string]string{outer: "OU-0001", inner: "IN-0001"} {
		if out, code := mrRun(t, dir, "ticket", "start", id); code != 0 {
			t.Fatalf("start %s: %s", id, out)
		}
		f := hookMsg(t, "")
		if out, code := mrRun(t, dir, "hook", "run", "prepare-commit-msg", f); code != 0 {
			t.Fatalf("prepare in %s: %s", dir, out)
		}
		got, _ := os.ReadFile(f)
		if !strings.Contains(string(got), "Ajiya: "+id) {
			t.Errorf("message in %s is %q, want trailer for %s", dir, got, id)
		}
	}
}

func TestInitWarnsAboutParentProject(t *testing.T) {
	outer, _, _ := nestedProjects(t)
	child := filepath.Join(outer, "fresh")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	out, code := mrRun(t, child, "init", "--name", "Fresh", "--prefix", "FR")
	if code != 0 {
		t.Fatalf("init failed: %s", out)
	}
	if !strings.Contains(out, "warning: "+outer) || !strings.Contains(out, "nested") {
		t.Errorf("init did not warn about the parent project in %s:\n%s", outer, out)
	}
	if _, err := os.Stat(filepath.Join(child, "ajiya.toml")); err != nil {
		t.Errorf("init should still create the nested project: %v", err)
	}
}

func TestInitWithoutParentProjectDoesNotWarn(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	out, code := mrRun(t, dir, "init", "--name", "Solo", "--prefix", "SO")
	if code != 0 {
		t.Fatalf("init failed: %s", out)
	}
	if strings.Contains(out, "warning") {
		t.Errorf("unexpected warning:\n%s", out)
	}
}
