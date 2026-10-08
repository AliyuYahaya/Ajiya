package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func mrGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

func mrRun(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	var out, errOut bytes.Buffer
	e := &env{stdout: &out, stderr: &errOut, dir: dir, noAgents: true}
	code := run(e, args)
	return out.String() + errOut.String(), code
}

// mrPlan builds a plan folder that is not a git repository, over apps that
// live in other folders: api and web are separate repositories, shared is a
// folder inside api's repository, docs is not in any repository. It has one
// ticket per app and returns the root.
func mrPlan(t *testing.T, appPaths map[string]string) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"AUTHOR", "COMMITTER"} {
		t.Setenv("GIT_"+k+"_NAME", "T")
		t.Setenv("GIT_"+k+"_EMAIL", "t@example.com")
	}
	for _, d := range []string{"api/shared", "web", "docs"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, d := range []string{"api", "web"} {
		mrGit(t, filepath.Join(root, d), "init", "-q", "-b", "main")
	}
	toml := "[project]\nname = \"Multi\"\nprefix = \"MI\"\n"
	for _, name := range []string{"api", "shared", "web", "docs", "ghost"} {
		if p, ok := appPaths[name]; ok {
			toml += "\n[[apps]]\nname = \"" + name + "\"\npath = \"" + p + "\"\n"
		}
	}
	if err := os.WriteFile(filepath.Join(root, "ajiya.toml"), []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, code := mrRun(t, root, "phase", "add", "build", "Build"); code != 0 {
		t.Fatalf("phase add: %s", out)
	}
	for _, app := range []string{"api", "web", "docs"} {
		if _, ok := appPaths[app]; !ok {
			continue
		}
		if out, code := mrRun(t, root, "ticket", "add", "--phase", "build", "--app", app, "Work on "+app, "--done-when", "it works"); code != 0 {
			t.Fatalf("ticket add: %s", out)
		}
	}
	return root
}

func mrCommit(t *testing.T, repo, trailer string) {
	t.Helper()
	mrGit(t, repo, "commit", "-q", "--allow-empty", "-m", "Work", "-m", trailer)
}

func TestMultiRepoEvidence(t *testing.T) {
	all := map[string]string{"api": "api", "shared": "api/shared", "web": "web", "docs": "docs", "ghost": "missing"}
	tests := []struct {
		name     string
		apps     map[string]string
		commitIn string // repo folder holding the commit "Ajiya: MI-0001"
		id       string
		wantDone bool
		wantOut  string
	}{
		{"commit in the first app repo", all, "api", "MI-0001", true, "evidence: "},
		{"commit in the second app repo", all, "web", "MI-0002", true, "in web"},
		{"commit in an unrelated repo is not found", map[string]string{"api": "api"}, "web", "MI-0001", false, "no commit references MI-0001"},
		{"no commit at all", all, "", "MI-0001", false, "no commit references MI-0001"},
		{"non-git app folder is ignored", all, "api", "MI-0003", false, "no commit references MI-0003"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := mrPlan(t, tt.apps)
			if tt.commitIn != "" {
				// MI-0001 is the api ticket, MI-0002 the web one.
				mrCommit(t, filepath.Join(root, tt.commitIn), "Ajiya: "+map[string]string{"api": "MI-0001", "web": "MI-0002"}[tt.commitIn])
				if tt.name == "commit in an unrelated repo is not found" {
					mrGit(t, filepath.Join(root, "web"), "commit", "-q", "--allow-empty", "-m", "x", "-m", "Ajiya: MI-0001")
				}
			}
			out, code := mrRun(t, root, "ticket", "done", tt.id)
			if (code == 0) != tt.wantDone {
				t.Fatalf("ticket done exit %d, want done=%v:\n%s", code, tt.wantDone, out)
			}
			if !strings.Contains(out, tt.wantOut) {
				t.Errorf("output %q does not contain %q", out, tt.wantOut)
			}
			if tt.wantDone {
				show, _ := mrRun(t, root, "ticket", "show", tt.id)
				if !strings.Contains(show, "["+filepath.Base(tt.commitIn)+"]") {
					t.Errorf("ticket show does not name the repository:\n%s", show)
				}
			}
		})
	}
}

// Two apps in one repository, and a repository named twice, are searched once.
func TestMultiRepoDedupe(t *testing.T) {
	root := mrPlan(t, map[string]string{"api": "api", "shared": "api/shared", "web": "web", "docs": "docs"})
	mrCommit(t, filepath.Join(root, "api"), "Ajiya: MI-0001")
	pr, err := load(&env{dir: root})
	if err != nil {
		t.Fatal(err)
	}
	repos := pr.repos()
	if len(repos) != 2 {
		t.Fatalf("repos = %v, want the api and web repositories once each", repos)
	}
	show, _ := mrRun(t, root, "ticket", "show", "MI-0001")
	if n := strings.Count(show, "Work [api]"); n != 1 {
		t.Errorf("the commit is listed %d times:\n%s", n, show)
	}
}

// A project in one repository behaves as before: no repository note anywhere.
func TestSingleRepoEvidenceUnchanged(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"AUTHOR", "COMMITTER"} {
		t.Setenv("GIT_"+k+"_NAME", "T")
		t.Setenv("GIT_"+k+"_EMAIL", "t@example.com")
	}
	mrGit(t, root, "init", "-q", "-b", "main")
	os.MkdirAll(filepath.Join(root, "app"), 0o755)
	os.WriteFile(filepath.Join(root, "ajiya.toml"), []byte("[project]\nname = \"One\"\nprefix = \"ON\"\n\n[[apps]]\nname = \"app\"\npath = \"app\"\n"), 0o644)
	mrRun(t, root, "phase", "add", "build", "Build")
	mrRun(t, root, "ticket", "add", "--phase", "build", "--app", "app", "Work", "--done-when", "ok")
	mrCommit(t, root, "Ajiya: ON-0001")
	out, code := mrRun(t, root, "ticket", "done", "ON-0001")
	if code != 0 || strings.Contains(out, "evidence:") || strings.Count(out, "\n") != 1 {
		t.Errorf("done = %d %q, want one plain line", code, out)
	}
	if show, _ := mrRun(t, root, "ticket", "show", "ON-0001"); strings.Contains(show, "[") && strings.Contains(show, "Work [") {
		t.Errorf("single-repo show names a repository:\n%s", show)
	}
}

// Hooks in an app repository find the plan in the folder above it, and the
// commit rule applies there.
func TestHookInstallAppsAndRun(t *testing.T) {
	root := mrPlan(t, map[string]string{"api": "api", "web": "web", "docs": "docs"})
	out, code := mrRun(t, root, "hook", "install", "--apps")
	if code != 0 {
		t.Fatalf("hook install --apps: %s", out)
	}
	for _, repo := range []string{"api", "web"} {
		for _, h := range hookNames {
			if _, err := os.Stat(filepath.Join(root, repo, ".git", "hooks", h)); err != nil {
				t.Errorf("%s hook missing in %s: %v", h, repo, err)
			}
		}
	}
	if !strings.Contains(out, "api/.git/hooks/commit-msg") {
		t.Errorf("output does not name the repositories:\n%s", out)
	}

	msg := func(text string) string {
		f := filepath.Join(t.TempDir(), "MSG")
		os.WriteFile(f, []byte(text), 0o644)
		return f
	}
	api := filepath.Join(root, "api")
	for _, c := range []struct {
		name, text string
		ok         bool
	}{
		{"no trailer", "Add\n", false},
		{"ticket", "Add\n\nAjiya: MI-0001\n", true},
		{"unknown ticket", "Add\n\nAjiya: MI-0099\n", false},
		{"chore", "Add\n\nAjiya: chore\n", true},
		{"revert", "Revert \"Add\"\n", true},
	} {
		out, code := mrRun(t, api, "hook", "run", "commit-msg", msg(c.text))
		if (code == 0) != c.ok {
			t.Errorf("%s: exit %d, want ok=%v\n%s", c.name, code, c.ok, out)
		}
	}
}

// Without --apps, hook install keeps its old behaviour: the root only.
func TestHookInstallDefaultIsRootOnly(t *testing.T) {
	root := mrPlan(t, map[string]string{"api": "api"})
	if out, code := mrRun(t, root, "hook", "install"); code == 0 {
		t.Errorf("a non-git root should still be refused without --apps:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(root, "api", ".git", "hooks", "commit-msg")); err == nil {
		t.Error("hook install without --apps wrote into an app repository")
	}
}

// W003 finds an evidence commit that lives in an app repository.
func TestCheckFindsEvidenceInAppRepo(t *testing.T) {
	root := mrPlan(t, map[string]string{"api": "api", "web": "web"})
	web := filepath.Join(root, "web")
	os.WriteFile(filepath.Join(web, "x.txt"), []byte("x"), 0o644)
	mrGit(t, web, "add", "x.txt")
	mrCommit(t, web, "Ajiya: MI-0001") // the api ticket, done in web
	if out, code := mrRun(t, root, "ticket", "done", "MI-0001"); code != 0 {
		t.Fatalf("done: %s", out)
	}
	out, _ := mrRun(t, root, "check")
	if !strings.Contains(out, "W003") || !strings.Contains(out, "changed nothing under api") {
		t.Errorf("check did not look in the web repository:\n%s", out)
	}
}
