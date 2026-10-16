package build

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

const histPhase = `<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Platform

Goal: Run the platform

| ID | App | Ticket | Done when | Depends | Tags | Status |
|---|---|---|---|---|---|---|
| AJ-0101 | infra | Provision RDS | Database reachable | - | tech:aws-rds dec:DEC-0003 | 🟩 Done · by Aliyu · 2026-09-20 |
| AJ-0102 | infra | Back up RDS | Daily backup runs | AJ-0101 | tech:aws-rds | 🟥 Pending |
`

const histDec3 = `<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# DEC-0003: Use PostgreSQL on AWS RDS

Status: Accepted · 2026-09-12 · decided by Aliyu
Tags: tech:aws-rds

## Context
Bookings need transactions.

## Decision
Use PostgreSQL on AWS RDS.

## Consequences
Ties us to AWS.
`

const histDec4 = `<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# DEC-0004: Use Supabase

Status: Accepted · 2026-10-01 · decided by Sam
Tags: tech:supabase

## Context
RDS is heavy.

## Decision
Use Supabase.

## Consequences
Rules out RDS.
`

const histChg1 = `<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# CHG-0001: Move the database to Supabase

Status: Open · 2026-10-16
Replaces: tech:aws-rds
With: tech:supabase
Decisions: DEC-0003
Paths: -
Snapshot: before-CHG-0001

## Candidates
| ID | Ticket | Reasons | Verdict | Verdict reason | New title | New done when |
|---|---|---|---|---|---|---|
| AJ-0101 | Provision RDS | tag:tech:aws-rds | replace | RDS is gone | Provision Supabase | Supabase reachable |
| AJ-0102 | Back up RDS | tag:tech:aws-rds | modify | Backups move | Back up Supabase | - |

## Results
-
`

const histChg2 = `<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# CHG-0002: Move to a monorepo

Status: Rejected · 2026-10-20 · by Aliyu · The team keeps two repos
Replaces: -
With: -
Decisions: -
Paths: services/**
Snapshot: before-CHG-0002

## Candidates
| ID | Ticket | Reasons | Verdict | Verdict reason | New title | New done when |
|---|---|---|---|---|---|---|

## Results
-
`

// A plan with one approved change: the Results ApplyChange writes become
// lists of IDs, and the decisions and the rejected change get their events in
// date order.
func TestHistoryApprovedChange(t *testing.T) {
	root := t.TempDir()
	write := func(path, text string) {
		t.Helper()
		full := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("ajiya/platform.md", histPhase)
	write("ajiya/decisions/DEC-0003.md", histDec3)
	write("ajiya/decisions/DEC-0004.md", histDec4)
	write("ajiya/changes/CHG-0001.md", histChg1)
	write("ajiya/changes/CHG-0002.md", histChg2)
	p, err := plan.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Broken)+len(p.BrokenDecisions)+len(p.BrokenChanges) > 0 {
		t.Fatalf("fixtures do not parse: %v %v %v", p.Broken, p.BrokenDecisions, p.BrokenChanges)
	}
	if _, err := p.ApplyChange(p.Changes[0], plan.ApplyOptions{Prefix: "AJ", NewDecision: "DEC-0004", By: "Aliyu", Date: "2026-10-18"}); err != nil {
		t.Fatal(err)
	}

	h := History(p)
	var kinds []string
	for _, e := range h {
		kinds = append(kinds, e.Date+" "+e.Kind+" "+e.ID)
	}
	want := []string{
		"2026-09-12 decision_accepted DEC-0003",
		"2026-10-01 decision_accepted DEC-0004",
		"2026-10-18 change_approved CHG-0001",
		"2026-10-18 decision_superseded DEC-0003",
		"2026-10-20 change_rejected CHG-0002",
	}
	if strings.Join(kinds, "\n") != strings.Join(want, "\n") {
		t.Fatalf("history =\n%s\nwant\n%s", strings.Join(kinds, "\n"), strings.Join(want, "\n"))
	}

	if got := h[1].Replaces; len(got) != 1 || got[0] != "DEC-0003" {
		t.Errorf("DEC-0004 replaces = %v, want [DEC-0003]", got)
	}
	if e := h[3]; e.Change != "CHG-0001" || e.SupersededBy != "DEC-0004" {
		t.Errorf("DEC-0003 superseded = change %q by %q, want CHG-0001 by DEC-0004", e.Change, e.SupersededBy)
	}
	if e := h[4]; e.Reason != "The team keeps two repos" || e.By != "Aliyu" {
		t.Errorf("rejected change = reason %q by %q", e.Reason, e.By)
	}

	o := h[2].Outcome
	if o == nil {
		t.Fatal("approved change has no outcome")
	}
	if len(o.Created) != 1 || o.Created[0].ID != "AJ-0103" || o.Created[0].Replaces != "AJ-0101" {
		t.Errorf("created = %+v, want AJ-0103 replacing AJ-0101", o.Created)
	}
	if len(o.Superseded) != 1 || o.Superseded[0].ID != "AJ-0101" || o.Superseded[0].By != "AJ-0103" || o.Superseded[0].WasState != "done" || !strings.Contains(o.Superseded[0].Was, "Done") {
		t.Errorf("superseded = %+v, want AJ-0101 by AJ-0103, was done", o.Superseded)
	}
	if len(o.Edited) != 1 || o.Edited[0].ID != "AJ-0102" || o.Edited[0].Fields[0] != "title" {
		t.Errorf("edited = %+v, want AJ-0102 title", o.Edited)
	}
	if len(o.Moved) != 1 || o.Moved[0].ID != "AJ-0102" || o.Moved[0].From[0] != "AJ-0101" || o.Moved[0].To[0] != "AJ-0103" {
		t.Errorf("moved = %+v, want AJ-0102 AJ-0101 -> AJ-0103", o.Moved)
	}
	if len(o.Decisions) != 1 || o.Decisions[0].ID != "DEC-0003" || o.Decisions[0].By != "DEC-0004" {
		t.Errorf("decisions = %+v, want DEC-0003 by DEC-0004", o.Decisions)
	}

	// data.js carries the list, empty rather than null for a plan without any.
	js, err := json.Marshal(&Data{History: []HistoryEvent{}})
	if err != nil || !strings.Contains(string(js), `"history":[]`) {
		t.Errorf("empty history marshals as %s, want \"history\":[]", js)
	}
}

func TestParseResultsOtherLines(t *testing.T) {
	o := ParseResults("- Created AJ-0105: Add form (follow-up to AJ-0100)\n- Superseded AJ-0100 with no replacement (was: In progress: x)\n- Superseded decision DEC-0001 with no replacement\n- milestone alpha: target AJ-0100 → AJ-0105\n- No ticket or decision needed changing")
	if len(o.Created) != 1 || o.Created[0].FollowUp != "AJ-0100" {
		t.Errorf("created = %+v", o.Created)
	}
	if len(o.Superseded) != 1 || o.Superseded[0].By != "" || o.Superseded[0].WasState != "in_progress" {
		t.Errorf("superseded = %+v", o.Superseded)
	}
	if len(o.Decisions) != 1 || o.Decisions[0].By != "" {
		t.Errorf("decisions = %+v", o.Decisions)
	}
	if len(o.Milestones) != 1 {
		t.Errorf("milestones = %v", o.Milestones)
	}
}
