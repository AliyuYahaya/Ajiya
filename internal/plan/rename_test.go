package plan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRename(t *testing.T) {
	root := t.TempDir()
	p := &Plan{Root: root, Raw: map[string][]byte{}, Broken: map[string]*ParseError{}}
	ph := &Phase{Slug: "api", Title: "API", Goal: "It works"}
	ph.Tickets = []*Ticket{{ID: "DE-0001", App: "infra", Title: "Schema", DoneWhen: "x", Phase: "api"}}
	if err := p.Save(ph); err != nil {
		t.Fatal(err)
	}
	if err := p.Save(&Phase{Slug: "web", Title: "Web", Goal: "Pages"}); err != nil {
		t.Fatal(err)
	}
	if err := p.Rename(ph, "web"); err == nil {
		t.Fatal("renaming onto an existing phase should fail")
	}
	if ph.Slug != "api" {
		t.Fatalf("failed rename changed the slug to %q", ph.Slug)
	}
	if err := p.Rename(ph, "backend"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, Dir, "api.md")); !os.IsNotExist(err) {
		t.Errorf("old file still there: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, Dir, "backend.md"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParsePhase("backend", data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "API" || len(got.Tickets) != 1 || got.Tickets[0].Phase != "backend" {
		t.Errorf("renamed phase = %+v", got)
	}
	if p.Phase("backend") != ph || p.Phase("api") != nil || p.Ticket("DE-0001").Phase != "backend" {
		t.Error("plan not updated after rename")
	}
	if _, ok := p.Raw["api"]; ok {
		t.Error("Raw still holds the old slug")
	}
}
