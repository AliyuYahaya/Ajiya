package config

import (
	"reflect"
	"strings"
	"testing"
)

const msHead = "[project]\nname = \"X\"\nprefix = \"X\"\n"

// names parses text and returns the milestone names in order.
func names(t *testing.T, text string) []string {
	t.Helper()
	c, err := Parse([]byte(text))
	if err != nil {
		t.Fatalf("%v\n%s", err, text)
	}
	var out []string
	for _, m := range c.MilestoneList() {
		out = append(out, m.Name)
	}
	return out
}

func TestConvertLaunchText(t *testing.T) {
	in := msHead + "\n# when we go live\n[launch]\n# the gate\ntarget = 'go-live'  # the phase\n\n[test]\ncommand = \"go test\"\n"
	want := msHead + "\n# when we go live\n[[milestones]]\nname = \"launch\"\n# the gate\ntargets = [\"go-live\"]  # the phase\n\n[test]\ncommand = \"go test\"\n"
	got, ok := ConvertLaunchText(in)
	if !ok || got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	c, err := Parse([]byte(got))
	if err != nil || c.Launch.Target != "" || len(c.Milestones) != 1 || c.Milestones[0].Targets[0] != "go-live" {
		t.Errorf("parsed %+v, %v", c, err)
	}
	for _, text := range []string{msHead, msHead + "\n# [launch]\n# target = \"x\"\n", msHead + "\n[launch]\ntarget = \"\"\n"} {
		if got, ok := ConvertLaunchText(text); ok || got != text {
			t.Errorf("converted %q to %q", text, got)
		}
	}
}

func TestAddMilestoneText(t *testing.T) {
	launch := msHead + "\n[[milestones]]\nname = \"launch\"\ntargets = [\"go-live\"]\n"
	tests := []struct {
		name, in, add, before, after string
		want                         []string
	}{
		{"first, no apps", msHead, "a", "", "", []string{"a"}},
		{"before launch by default", launch, "a", "", "", []string{"a", "launch"}},
		{"after launch when asked", launch, "a", "", "launch", []string{"launch", "a"}},
		{"old form converted, before launch", msHead + "\n[launch]\ntarget = \"go-live\"\n", "a", "", "", []string{"a", "launch"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddMilestoneText(tt.in, Milestone{Name: tt.add, Targets: []string{"p", "X-0001"}}, tt.before, tt.after)
			if err != nil {
				t.Fatal(err)
			}
			if n := names(t, got); !reflect.DeepEqual(n, tt.want) {
				t.Errorf("order %v, want %v\n%s", n, tt.want, got)
			}
			if !strings.Contains(got, "targets = [\"p\", \"X-0001\"]\n") {
				t.Errorf("targets missing:\n%s", got)
			}
		})
	}

	// Without milestones, the new table goes before the apps.
	apps := msHead + "\n[[apps]]\nname = \"api\"\npath = \"api\"\n"
	got, _ := AddMilestoneText(apps, Milestone{Name: "a", Targets: []string{"p"}}, "", "")
	want := msHead + "\n[[milestones]]\nname = \"a\"\ntargets = [\"p\"]\n\n[[apps]]\nname = \"api\"\npath = \"api\"\n"
	if got != want {
		t.Errorf("before apps: got\n%s\nwant\n%s", got, want)
	}
	// After the last milestone when there is no launch.
	got, _ = AddMilestoneText(got, Milestone{Name: "b", Targets: []string{"p"}}, "", "")
	if !strings.Contains(got, "targets = [\"p\"]\n\n[[milestones]]\nname = \"b\"\ntargets = [\"p\"]\n\n[[apps]]") {
		t.Errorf("after last:\n%s", got)
	}
	if _, err := AddMilestoneText(launch, Milestone{Name: "launch", Targets: []string{"p"}}, "", ""); err == nil {
		t.Error("added a duplicate")
	}
	if _, err := AddMilestoneText(launch, Milestone{Name: "a", Targets: []string{"p"}}, "nope", ""); err == nil {
		t.Error("added before a missing milestone")
	}
}

const three = `[project]
name = "X"
prefix = "X"

# first gate
[[milestones]]
name = "a"   # staging
targets = ["p"]

[[milestones]]
name = 'b'
targets = [
  "q",   # the q phase
  "r",
]

# the big day
[[milestones]]
name = "launch"
targets = ["go-live"]

[test]
command = "go test"
`

func TestMoveMilestoneText(t *testing.T) {
	tests := []struct {
		name, before, after string
		want                []string
	}{
		{"a", "", "launch", []string{"b", "launch", "a"}},
		{"launch", "a", "", []string{"launch", "a", "b"}},
		{"b", "a", "", []string{"b", "a", "launch"}},
		{"a", "", "b", []string{"b", "a", "launch"}},
	}
	for _, tt := range tests {
		got, err := MoveMilestoneText(three, tt.name, tt.before, tt.after)
		if err != nil {
			t.Fatal(err)
		}
		if n := names(t, got); !reflect.DeepEqual(n, tt.want) {
			t.Errorf("move %s: %v, want %v\n%s", tt.name, n, tt.want, got)
		}
		// Comments travel with their table; other tables stay.
		for _, keep := range []string{"# first gate\n[[milestones]]\nname = \"a\"   # staging\n", "# the big day\n[[milestones]]\nname = \"launch\"", "\"q\",   # the q phase", "[test]\ncommand = \"go test\"\n"} {
			if !strings.Contains(got, keep) {
				t.Errorf("move %s lost %q:\n%s", tt.name, keep, got)
			}
		}
		if strings.Contains(got, "\n\n\n") {
			t.Errorf("move %s left a double blank line:\n%s", tt.name, got)
		}
	}
	for _, bad := range [][3]string{{"a", "", ""}, {"a", "b", "b"}, {"a", "a", ""}, {"nope", "a", ""}, {"a", "nope", ""}} {
		if _, err := MoveMilestoneText(three, bad[0], bad[1], bad[2]); err == nil {
			t.Errorf("move %v: no error", bad)
		}
	}
}

func TestRemoveMilestoneText(t *testing.T) {
	got, err := RemoveMilestoneText(three, "b")
	if err != nil || !reflect.DeepEqual(names(t, got), []string{"a", "launch"}) || strings.Contains(got, "\"q\"") {
		t.Errorf("remove b: %v\n%s", err, got)
	}
	got, _ = RemoveMilestoneText(three, "a")
	if strings.Contains(got, "first gate") || !strings.HasPrefix(got, msHead+"\n[[milestones]]\nname = 'b'") {
		t.Errorf("remove a:\n%s", got)
	}
	if _, err := RemoveMilestoneText(three, "nope"); err == nil {
		t.Error("removed a missing milestone")
	}
	// The old form is converted, then removed.
	got, err = RemoveMilestoneText(msHead+"\n[launch]\ntarget = \"x\"\n", "launch")
	if err != nil || got != msHead {
		t.Errorf("remove old launch: %v\n%q", err, got)
	}
}

func TestSetMilestoneTargetsText(t *testing.T) {
	got, err := SetMilestoneTargetsText(three, "b", []string{"s"})
	if err != nil {
		t.Fatal(err)
	}
	c, err := Parse([]byte(got))
	if err != nil || !reflect.DeepEqual(c.Milestones[1].Targets, []string{"s"}) || strings.Contains(got, "\"r\"") ||
		!strings.Contains(got, "name = 'b'\ntargets = [\"s\"]\n\n# the big day") {
		t.Errorf("set b: %v\n%s", err, got)
	}
	got, _ = SetMilestoneTargetsText(three, "launch", []string{"X-0001"})
	if !strings.Contains(got, "name = \"launch\"\ntargets = [\"X-0001\"]\n") {
		t.Errorf("set launch:\n%s", got)
	}
	// A missing launch milestone is added last.
	got, _ = SetMilestoneTargetsText(msHead+"\n[[milestones]]\nname = \"a\"\ntargets = [\"p\"]\n", "launch", []string{"q"})
	if n := names(t, got); !reflect.DeepEqual(n, []string{"a", "launch"}) {
		t.Errorf("add launch: %v\n%s", n, got)
	}
}
