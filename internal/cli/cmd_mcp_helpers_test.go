package cli

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func marshalIndent(t *testing.T, v any) string {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// cliJSON runs a read command with --json, as the CLI does, and trims it the
// way the tool result is trimmed.
func cliJSON(t *testing.T, dir string, args []string) string {
	t.Helper()
	var out, errOut strings.Builder
	e := &env{stdout: &out, stderr: &errOut, dir: dir, now: func() time.Time { return statusNow }}
	run(e, append(append([]string{}, args...), "--json"))
	return strings.TrimSpace(out.String())
}

func firstJSONName(t *testing.T, list string) string {
	t.Helper()
	var rows []map[string]any
	if err := json.Unmarshal([]byte(list), &rows); err != nil || len(rows) == 0 {
		t.Fatalf("milestone list: %v %s", err, list)
	}
	name, _ := rows[0]["name"].(string)
	if name == "" {
		t.Fatalf("no name in %s", list)
	}
	return name
}
