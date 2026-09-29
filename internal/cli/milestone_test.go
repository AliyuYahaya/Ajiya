package cli

import (
	"reflect"
	"testing"
)

func TestMilestoneJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	gitIn(t, dir, "init", "-q", "-b", "main")
	ajiya(t, dir, 0, "init", "--name", "Demo", "--prefix", "DE")

	type stats struct {
		Name     string
		Targets  []string
		Required int
		Closed   int
		Percent  int
		CanStart int `json:"can_start"`
	}
	var list []stats
	decode(t, ajiya(t, dir, 0, "milestone", "list", "--json"), &list)
	if list == nil || len(list) != 0 {
		t.Errorf("milestone list --json with none = %+v", list)
	}

	ajiya(t, dir, 0, "phase", "add", "gate", "Staging")
	ajiya(t, dir, 0, "phase", "add", "live", "Live")
	ajiya(t, dir, 0, "phase", "add", "misc", "Odds")
	ajiya(t, dir, 0, "ticket", "add", "--phase", "gate", "--app", "infra", "--done-when", "x", "Host")
	ajiya(t, dir, 0, "ticket", "add", "--phase", "live", "--app", "infra", "--done-when", "x", "--depends", "DE-0001", "--human", "keys", "Deploy")
	ajiya(t, dir, 0, "ticket", "add", "--phase", "misc", "--app", "infra", "--done-when", "x", "Tidy")
	ajiya(t, dir, 0, "launch", "set", "live")
	ajiya(t, dir, 0, "milestone", "add", "staging", "--targets", "gate")

	decode(t, ajiya(t, dir, 0, "milestone", "list", "--json"), &list)
	want := []stats{
		{Name: "staging", Targets: []string{"gate"}, Required: 1, CanStart: 1},
		{Name: "launch", Targets: []string{"live"}, Required: 2, CanStart: 1},
	}
	if !reflect.DeepEqual(list, want) {
		t.Errorf("milestone list --json = %+v", list)
	}

	var show struct {
		stats
		Position int
		Kinds    []struct{ Target, Kind string } `json:"target_kinds"`
		Open     []struct {
			ticket
			WaitingOn []string `json:"waiting_on"`
			CanStart  bool     `json:"can_start"`
		}
		NeedsHuman []string `json:"needs_human"`
	}
	decode(t, ajiya(t, dir, 0, "milestone", "show", "launch", "--json"), &show)
	if show.Name != "launch" || show.Position != 2 || len(show.Kinds) != 1 || show.Kinds[0].Kind != "phase" ||
		len(show.Open) != 2 || !show.Open[0].CanStart || show.Open[0].WaitingOn == nil ||
		show.Open[1].CanStart || !reflect.DeepEqual(show.Open[1].WaitingOn, []string{"DE-0001"}) ||
		!reflect.DeepEqual(show.NeedsHuman, []string{"DE-0002"}) {
		t.Errorf("milestone show --json = %+v", show)
	}

	var next []struct {
		ticket
		Launch    bool
		Milestone string
		Unblocks  int
	}
	decode(t, ajiya(t, dir, 0, "next", "--json"), &next)
	if len(next) != 2 || next[0].ID != "DE-0001" || next[0].Milestone != "staging" || !next[0].Launch ||
		next[1].ID != "DE-0003" || next[1].Milestone != "" || next[1].Launch {
		t.Errorf("next --json = %+v", next)
	}
}
