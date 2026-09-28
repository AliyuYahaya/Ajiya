package changelog

import (
	"strings"
	"testing"

	"github.com/AliyuYahaya/Ajiya/internal/gitx"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

func TestBuild(t *testing.T) {
	p := &plan.Plan{Phases: []*plan.Phase{
		{Slug: "a-core", Title: "Core", Tickets: []*plan.Ticket{
			{ID: "DE-0002", Title: "Two", Phase: "a-core", Status: plan.Status{State: plan.InProgress}},
			{ID: "DE-0001", Title: "One", Phase: "a-core", Status: plan.Status{State: plan.Done}},
		}},
		{Slug: "b-empty", Title: "Empty", Tickets: []*plan.Ticket{
			{ID: "DE-0003", Title: "Three", Phase: "b-empty"},
		}},
		{Slug: "c-late", Title: "Late", Tickets: []*plan.Ticket{
			{ID: "DE-0004", Title: "Four", Phase: "c-late", Status: plan.Status{State: plan.Dropped, Aliases: []string{"OLD-1"}}},
		}},
	}}
	c := func(hash, subject string, merge, chore bool, ids ...string) gitx.Commit {
		return gitx.Commit{Hash: hash, Subject: subject, Merge: merge, Refs: gitx.Refs{IDs: ids, Chore: chore}}
	}
	// Newest first, as git log gives them.
	log := []gitx.Commit{
		c("9999999aaa", "Merge side", true, false),
		c("8888888aaa", "Old ID and new", false, false, "OLD-1", "DE-0004"),
		c("7777777aaa", "Unknown", false, false, "DE-0099"),
		c("6666666aaa", "Upkeep", false, true),
		c("5555555aaa", "Both", false, false, "DE-0002", "DE-0001"),
		c("4444444aaa", "First", false, false, "DE-0001"),
	}
	var b strings.Builder
	if err := Build(p, log).WriteMarkdown(&b); err != nil {
		t.Fatal(err)
	}
	want := `# Changelog

## Core

- **DE-0001** One — done
  - 4444444 First
  - 5555555 Both
- **DE-0002** Two — in progress
  - 5555555 Both

## Late

- **DE-0004** Four — dropped
  - 8888888 Old ID and new

## Other changes

- 6666666 Upkeep
- 7777777 Unknown
`
	if b.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", b.String(), want)
	}
}

func TestEmpty(t *testing.T) {
	cl := Build(&plan.Plan{}, nil)
	if cl.Phases == nil || cl.Other == nil {
		t.Errorf("empty lists should not be nil: %+v", cl)
	}
	var b strings.Builder
	cl.WriteMarkdown(&b)
	if b.String() != "# Changelog\n\nNo changes.\n" {
		t.Errorf("got %q", b.String())
	}
}
