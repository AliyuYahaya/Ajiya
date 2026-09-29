package importer

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseLegacyConfig(t *testing.T) {
	c, err := ParseLegacyConfig([]byte(`
prefixes = ["API", "WEB"]
rollout = "r.md"
[[sources]]
area = "Web app"
path = "a.md"
[[sources]]
area = "API"
path = "b.md"
phase = "backend"
`))
	if err != nil {
		t.Fatal(err)
	}
	if c.Sources[0].Phase != "web-app" || c.Sources[1].Phase != "backend" || c.RolloutPhase != "rollout" {
		t.Errorf("phases: %q %q %q", c.Sources[0].Phase, c.Sources[1].Phase, c.RolloutPhase)
	}

	src := "[[sources]]\narea = \"API\"\npath = \"a.md\"\n"
	for _, tc := range []struct{ in, want string }{
		{`prefixes = ["API"]` + "\nextra = 1\n" + src, `unknown key "extra"`},
		{src, "prefixes is empty"},
		{`prefixes = ["RO"]` + "\n" + src, `prefix "RO"`},
		{`prefixes = ["api"]` + "\n" + src, `prefix "api"`},
		{`prefixes = ["API"]`, "no [[sources]]"},
		{`prefixes = ["API"]` + "\n[[sources]]\narea = \"API\"\n", "needs both area and path"},
		{`prefixes = ["API"]` + "\n[[sources]]\narea = \"!!\"\npath = \"a.md\"\n", `phase "" of source !!`},
		{`prefixes = ["API"]` + "\n" + src + src, `two sources use phase "api"`},
		{`prefixes = ["API"]` + "\nrollout = \"r.md\"\nrollout_phase = \"api\"\n" + src, "also a source phase"},
		{`prefixes = ["API"]` + "\n[[sources]]\narea = \" API\"\npath = \"a.md\"\n", "must be one line"},
	} {
		_, err := ParseLegacyConfig([]byte(tc.in))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%q: got %v, want %q", tc.in, err, tc.want)
		}
	}
}

func TestLegacySlug(t *testing.T) {
	for in, want := range map[string]string{"API": "api", "Web app": "web-app", " Priority  gaps! ": "priority-gaps", "Ops/2": "ops-2"} {
		if got := LegacySlug(in); got != want {
			t.Errorf("LegacySlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLegacyStatusOf(t *testing.T) {
	for _, tc := range []struct {
		row  string
		want LegacyStatus
	}{
		{"| A-1 | x | 🟩 Done |", LegacyDone},
		{"| A-1 | x | 🟨 |", LegacyPartial},
		{"| A-1 | x | 🟥 |", LegacyPending},
		{"| A-1 | x | |", LegacyUnknown},
		{"| A-1 | x | 🟩 once | 🟥 again |", LegacyPending}, // the last marked cell wins
		{"| A-1 | x | 🟥 then 🟩 |", LegacyDone},            // within a cell, 🟩 is checked first
		{"| A-1 | x 🟨 | done | |", LegacyPartial},         // an empty last cell is passed over
	} {
		if got := legacyStatusOf(legacyCells(tc.row)); got != tc.want {
			t.Errorf("%s: got %s, want %s", tc.row, got, tc.want)
		}
	}
}

func TestLegacyTyped(t *testing.T) {
	for _, tc := range []struct {
		in     string
		want   LegacyStatus
		marked bool
	}{
		{"🟩 Done", LegacyDone, true},
		{"🟨 Half", LegacyPartial, true},
		{"Completed in May", LegacyDone, false},
		{"In progress", LegacyPartial, false},
		{"not started", LegacyPending, false},
		{"Blocked on legal", LegacyPending, false},
		{"Maybe", LegacyUnknown, false},
		{"", LegacyUnknown, false},
	} {
		got, marked := legacyTyped(tc.in)
		if got != tc.want || marked != tc.marked {
			t.Errorf("%q: got %s %v, want %s %v", tc.in, got, marked, tc.want, tc.marked)
		}
	}
}

const legacyWBS = `# Title

## EPIC A: Accounts

| ID | Ticket | Status |
|---|---|---|
| API-001 | Sign up \| in | 🟩 |
| API-002 | Log in, tracked as WEB-010 | 🟨 |
| API-003 | Same as API-003 is not an alias | 🟥 |
### EPIC B
| API-004 | Pay | cross-reference only: API-001 |
| API-004 | Pay again | 🟥 |
| RO-1 | not a ticket | |
| XYZ-9 | unknown | 🟥 |
| WEB-010 | Log in page | 🟩 |
`

func TestParseLegacySource(t *testing.T) {
	ts, notes := ParseLegacySource(legacyWBS, "wbs.md", 2, []string{"API", "WEB"})
	type row struct {
		ID, Title string
		Status    LegacyStatus
		Epic      string
		Line      int
		AliasOf   string
	}
	var got []row
	for _, x := range ts {
		if x.Source != 2 {
			t.Errorf("%s: source %d", x.ID, x.Source)
		}
		got = append(got, row{x.ID, x.Title, x.Status, x.Epic, x.Line, x.AliasOf})
	}
	want := []row{
		{"API-001", "Sign up | in", LegacyDone, "EPIC A: Accounts", 7, ""},
		{"API-002", "Log in, tracked as WEB-010", LegacyPartial, "EPIC A: Accounts", 8, "WEB-010"},
		{"API-003", "Same as API-003 is not an alias", LegacyPending, "EPIC A: Accounts", 9, ""},
		{"API-004", "Pay", LegacyUnknown, "EPIC B", 11, "API-001"},
		{"API-004", "Pay again", LegacyPending, "EPIC B", 12, ""},
		{"WEB-010", "Log in page", LegacyDone, "EPIC B", 15, ""},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%v\nwant\n%v", got, want)
	}
	wantNotes := []string{
		"wbs.md: 1 row(s) skipped because the prefix XYZ- is not in prefixes: XYZ-9",
		"wbs.md: 1 ticket(s) with no 🟩/🟨/🟥 mark, imported as pending: API-004",
	}
	if !reflect.DeepEqual(notes, wantNotes) {
		t.Errorf("notes:\n%q\nwant\n%q", notes, wantNotes)
	}

	if _, notes := ParseLegacySource("no table here", "x.md", 0, []string{"API"}); len(notes) != 1 || !strings.Contains(notes[0], "no ticket rows found") {
		t.Errorf("empty file notes: %q", notes)
	}
}

func TestResolveLegacyScope(t *testing.T) {
	prefixes := []string{"API", "WEB"}
	ts, _ := ParseLegacySource(legacyWBS, "wbs.md", 0, prefixes)
	re := newLegacyRE(prefixes)
	ids := func(idx []int) string {
		var s []string
		for _, i := range idx {
			s = append(s, ts[i].ID+"/"+ts[i].Title)
		}
		return strings.Join(s, "; ")
	}
	for _, tc := range []struct {
		scope                                string
		counted, aliased, missing, ambiguous string
	}{
		{"API-001", "API-001/Sign up | in", "", "", ""},
		{"API-001..API-003", "API-001/Sign up | in; API-002/Log in, tracked as WEB-010; API-003/Same as API-003 is not an alias", "", "", ""},
		{"API[epic b]", "API-004/Pay; API-004/Pay again", "", "", ""},
		{"WEB[EPIC B]", "WEB-010/Log in page", "", "", ""},
		{"API-004", "API-004/Pay; API-004/Pay again", "", "", "API-004"},
		// a cross-reference counts once only when its target is in scope too
		{"API-002, WEB-010", "WEB-010/Log in page", "API-002/Log in, tracked as WEB-010", "", ""},
		{"API-002", "API-002/Log in, tracked as WEB-010", "", "", ""},
		{"API-003..API-005", "API-003/Same as API-003 is not an alias; API-004/Pay; API-004/Pay again", "", "API-005", "API-004"},
		{"API-001, API-001..API-001", "API-001/Sign up | in", "", "", ""},
		{"API-003..API-001, API-001..WEB-002, RO-1, API[EPIC Z], nonsense", "", "", "API-003..API-001, API-001..WEB-002, RO-1, API[EPIC Z], nonsense", ""},
	} {
		c, a, m, amb := resolveLegacyScope(re, tc.scope, ts)
		if ids(c) != tc.counted || ids(a) != tc.aliased || strings.Join(m, ", ") != tc.missing || strings.Join(amb, ", ") != tc.ambiguous {
			t.Errorf("%s:\n counted %q\n aliased %q\n missing %v\n ambiguous %v", tc.scope, ids(c), ids(a), m, amb)
		}
	}
}

func TestParseLegacyRollout(t *testing.T) {
	prefixes := []string{"API", "WEB"}
	ts, _ := ParseLegacySource(legacyWBS, "wbs.md", 0, prefixes)
	pkgs, notes := ParseLegacyRollout(`## Stage
| ID | Name | Scope | Depends | Launch | Status |
|---|---|---|---|---|---|
| RO-1 | All done | API-001, WEB-010 | - | Required | |
| RO-2 | Mixed | API-001..API-003 | RO-1 and RO-9 | Required | |
| RO-3 | Typed | none | RO-3 | Optional | 🟩 Done |
| RO-4 | Words | - | RO-1, RO-2 | Optional | pending |
| RO-5 | Nothing | API-099 | | | |
| RO-6 | Short |
`, "r.md", ts, prefixes)
	var got []string
	for _, p := range pkgs {
		got = append(got, p.ID+" "+string(p.Status)+" "+strings.Join(p.Depends, ","))
	}
	want := []string{"RO-1 done ", "RO-2 partial RO-1,RO-9", "RO-3 done RO-3", "RO-4 pending RO-1,RO-2", "RO-5 unknown "}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("packages %q, want %q", got, want)
	}
	if pkgs[2].Scope != "" || pkgs[3].Scope != "" {
		t.Errorf("scopes %q %q should be empty", pkgs[2].Scope, pkgs[3].Scope)
	}
	wantNotes := []string{
		`RO-4: typed status "pending" has no mark, read as pending`,
		"RO-5: scope entries not found: API-099",
		"RO-5: scope matched no tickets, imported as pending",
		"r.md:9: RO-6 has 2 columns, needs 6 (ID, name, scope, depends, launch, status); skipped",
		"RO-2: depends on RO-9, which does not exist; not linked",
		"RO-3: depends on itself; not linked",
	}
	if !reflect.DeepEqual(notes, wantNotes) {
		t.Errorf("notes:\n%q\nwant\n%q", notes, wantNotes)
	}
}

func TestParseLegacyRolloutHeadingsAndApp(t *testing.T) {
	c, err := ParseLegacyConfig([]byte("prefixes = [\"API\"]\n[[sources]]\narea = \"API\"\npath = \"a.md\"\napp = \"backend\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Sources[0].App != "backend" {
		t.Errorf("app = %q", c.Sources[0].App)
	}
	rollout := `# Rollout

| RO-0 | Before any heading | - | - | Required | Done |

## Gate 0: Staging proven

| RO-1 | One | - | - | Required | Done |

### A sub-heading is not a milestone

| RO-2 | Two | - | RO-1 | Required | Done |

## Phase A: Safety

| RO-3 | Three | - | RO-2 | Required | Done |
`
	pkgs, _ := ParseLegacyRollout(rollout, "r.md", nil, c.Prefixes)
	var got []string
	for _, p := range pkgs {
		got = append(got, p.ID+"="+p.Heading)
	}
	want := []string{"RO-0=", "RO-1=Gate 0: Staging proven", "RO-2=Gate 0: Staging proven", "RO-3=Phase A: Safety"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("headings = %q, want %q", got, want)
	}
}
