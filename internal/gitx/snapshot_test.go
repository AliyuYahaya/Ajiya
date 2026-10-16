package gitx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateSnapshot(t *testing.T) {
	dir := newRepo(t)
	if _, err := CreateSnapshot(dir, "ajiya", "x"); err == nil || !strings.Contains(err.Error(), "no commits") {
		t.Errorf("empty repo: %v", err)
	}
	os.MkdirAll(filepath.Join(dir, "ajiya"), 0o755)
	os.WriteFile(filepath.Join(dir, "ajiya", "core.md"), []byte("a\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "other.txt"), []byte("a\n"), 0o644)
	run(t, dir, "add", "-A")
	run(t, dir, "commit", "-q", "-m", "Plan")

	s, err := CreateSnapshot(dir, "ajiya", "before-CHG-0001")
	if err != nil {
		t.Fatal(err)
	}
	head := run(t, dir, "rev-parse", "--short", "HEAD")
	if s.Tag != "ajiya/plan/before-CHG-0001" || s.Commit != head || s.Message != "Ajiya plan snapshot before-CHG-0001" {
		t.Errorf("snapshot = %+v, head %s", s, head)
	}
	if got := run(t, dir, "cat-file", "-t", s.Tag); got != "tag" {
		t.Errorf("tag type = %s, want an annotated tag", got)
	}

	for _, bad := range []string{"", "list", "a b", "a..b", "-x", "x/", "x.lock", "x~1"} {
		if _, err := CreateSnapshot(dir, "ajiya", bad); err == nil {
			t.Errorf("name %q accepted", bad)
		}
	}
	if _, err := CreateSnapshot(dir, "ajiya", "before-CHG-0001"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("duplicate: %v", err)
	}

	// A change outside the plan folder is fine; inside it, even untracked, is not.
	os.WriteFile(filepath.Join(dir, "other.txt"), []byte("b\n"), 0o644)
	if _, err := CreateSnapshot(dir, "ajiya", "outside"); err != nil {
		t.Errorf("change outside ajiya/: %v", err)
	}
	os.WriteFile(filepath.Join(dir, "ajiya", "new.md"), []byte("n\n"), 0o644)
	_, err = CreateSnapshot(dir, "ajiya", "inside")
	if err == nil || !strings.Contains(err.Error(), "ajiya/new.md") {
		t.Errorf("untracked file in ajiya/: %v", err)
	}

	snaps, err := Snapshots(dir)
	if err != nil || len(snaps) != 2 {
		t.Fatalf("Snapshots = %+v, %v", snaps, err)
	}
}
