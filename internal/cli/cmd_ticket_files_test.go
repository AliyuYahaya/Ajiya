package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tfWrite(t *testing.T, repo, rel string) {
	t.Helper()
	p := filepath.Join(repo, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(rel+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// tfCommit writes the files and commits them with the given trailer.
func tfCommit(t *testing.T, repo, trailer string, files ...string) {
	t.Helper()
	for _, f := range files {
		if f != "ajiya" { // the plan folder already exists
			tfWrite(t, repo, f)
		}
	}
	mrGit(t, repo, append([]string{"add", "--"}, files...)...)
	mrGit(t, repo, "commit", "-q", "--allow-empty", "-m", "Work", "-m", trailer)
}

// A project in one repository with five tickets (TF-0001 to TF-0005).
func tfPlan(t *testing.T) string {
	t.Helper()
	root := mrPlan(t, map[string]string{"docs": "docs"})
	// mrPlan leaves the root outside git; make this one a repository itself.
	mrGit(t, root, "init", "-q", "-b", "main")
	for i := 0; i < 4; i++ {
		if out, code := mrRun(t, root, "ticket", "add", "--phase", "build", "--app", "docs", "More", "--done-when", "ok"); code != 0 {
			t.Fatalf("ticket add: %s", out)
		}
	}
	return root
}

// tfPlanChange edits the plan, so the next commit touches files under ajiya/.
func tfPlanChange(t *testing.T, root string) {
	t.Helper()
	if out, code := mrRun(t, root, "ticket", "add", "--phase", "build", "--app", "docs", "Extra", "--done-when", "ok"); code != 0 {
		t.Fatalf("ticket add: %s", out)
	}
}

func TestTicketFiles(t *testing.T) {
	root := tfPlan(t) // MI-0001..MI-0005; the root repository holds the plan and the api folder
	tfCommit(t, root, "Ajiya: MI-0001", "README.md", "docs/b.go", "docs/a.go")
	tfPlanChange(t, root)
	tfCommit(t, root, "Ajiya: MI-0001", "docs/a.go", "docs/c.go", "ajiya")
	tfCommit(t, root, "Ajiya: MI-0001, MI-0003", "docs/shared/x.go")
	tfPlanChange(t, root)
	tfCommit(t, root, "Ajiya: MI-0004", "ajiya") // only plan files
	// A merge commit: its own diff is not listed, the branch commit's is.
	mrGit(t, root, "checkout", "-q", "-b", "side")
	tfCommit(t, root, "Ajiya: MI-0002", "docs/side.go")
	mrGit(t, root, "checkout", "-q", "main")
	tfCommit(t, root, "Ajiya: chore", "unrelated.txt")
	mrGit(t, root, "merge", "-q", "--no-ff", "-m", "Merge side", "-m", "Ajiya: MI-0001", "side")

	out, code := mrRun(t, root, "ticket", "files", "MI-0001")
	if code != 0 {
		t.Fatalf("exit %d:\n%s", code, out)
	}
	want := "README.md\ndocs/a.go\ndocs/b.go\ndocs/c.go\ndocs/shared/x.go\n"
	if out != want {
		t.Errorf("MI-0001 files:\n%s\nwant:\n%s", out, want)
	}
	// The merge commit names MI-0001 but adds no files (docs/side.go is MI-0002's).
	if strings.Contains(out, "side.go") || strings.Contains(out, "ajiya/") {
		t.Errorf("merge diff or plan file listed:\n%s", out)
	}

	// A commit naming two tickets counts for both.
	if out, _ := mrRun(t, root, "ticket", "files", "MI-0003"); out != "docs/shared/x.go\n" {
		t.Errorf("MI-0003 files: %q", out)
	}
	if out, _ := mrRun(t, root, "ticket", "files", "MI-0002"); out != "docs/side.go\n" {
		t.Errorf("MI-0002 files: %q", out)
	}
	// Only plan files changed.
	if out, _ := mrRun(t, root, "ticket", "files", "MI-0004"); !strings.Contains(out, "none changed a file outside the plan folder") {
		t.Errorf("MI-0004: %q", out)
	}
	// No commits at all.
	if out, _ := mrRun(t, root, "ticket", "files", "MI-0005"); !strings.Contains(out, "No commit references MI-0005") {
		t.Errorf("MI-0005: %q", out)
	}
	// Unknown ticket.
	if _, code := mrRun(t, root, "ticket", "files", "MI-0099"); code == 0 {
		t.Error("unknown ticket should fail")
	}

	out, _ = mrRun(t, root, "ticket", "files", "MI-0001", "--json")
	var got struct {
		ID      string
		Commits int
		Files   []struct{ Path, Repo string }
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if got.ID != "MI-0001" || got.Commits != 4 || len(got.Files) != 5 || got.Files[0].Path != "README.md" {
		t.Errorf("json = %+v", got)
	}
	out, _ = mrRun(t, root, "ticket", "files", "MI-0005", "--json")
	if !strings.Contains(out, `"files": []`) || !strings.Contains(out, `"commits": 0`) {
		t.Errorf("empty json = %s", out)
	}
}

// With app repositories, files are named from the project root and the search
// agrees with ticket show.
func TestTicketFilesMultiRepo(t *testing.T) {
	root := mrPlan(t, map[string]string{"api": "api", "web": "web"})
	tfCommit(t, filepath.Join(root, "api"), "Ajiya: MI-0001", "z.go", "a.go")
	tfCommit(t, filepath.Join(root, "web"), "Ajiya: MI-0001, MI-0002", "src/app.ts")
	out, code := mrRun(t, root, "ticket", "files", "MI-0001")
	if want := "api/a.go\napi/z.go\nweb/src/app.ts\n"; code != 0 || out != want {
		t.Errorf("files (exit %d):\n%s\nwant:\n%s", code, out, want)
	}
	if out, _ := mrRun(t, root, "ticket", "files", "MI-0002"); out != "web/src/app.ts\n" {
		t.Errorf("MI-0002: %q", out)
	}
	show, _ := mrRun(t, root, "ticket", "show", "MI-0001")
	if strings.Count(show, "Work [") != 2 {
		t.Errorf("ticket show disagrees:\n%s", show)
	}
	js, _ := mrRun(t, root, "ticket", "files", "MI-0001", "--json")
	if !strings.Contains(js, `"repo": "web"`) || !strings.Contains(js, `"path": "web/src/app.ts"`) {
		t.Errorf("json:\n%s", js)
	}
}

// The function change impact calls: one pass, every ticket.
func TestFilesByTicketAllTickets(t *testing.T) {
	root := mrPlan(t, map[string]string{"api": "api", "web": "web"})
	tfCommit(t, filepath.Join(root, "api"), "Ajiya: MI-0001", "a.go")
	tfCommit(t, filepath.Join(root, "web"), "Ajiya: MI-0002", "w.ts")
	pr, err := load(&env{dir: root})
	if err != nil {
		t.Fatal(err)
	}
	m, err := pr.filesByTicket()
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 2 || len(m["MI-0001"].Files) != 1 || pr.displayPath(m["MI-0002"].Files[0]) != "web/w.ts" || m["MI-0003"] != nil {
		t.Errorf("filesByTicket = %+v", m)
	}
}
