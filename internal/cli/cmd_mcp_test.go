package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// mcpConnect starts the server over in-memory pipes and returns a client
// session. cwd is the folder the server "started" in; fixed is --dir.
func mcpConnect(t *testing.T, cwd, fixed string, roots ...string) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	srv := newMCPServer(cwd, fixed, func() time.Time { return statusNow })
	st, ct := mcp.NewInMemoryTransports()
	ss, err := srv.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ss.Close() })
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	for _, r := range roots {
		client.AddRoots(&mcp.Root{URI: fileURI(r)})
	}
	// Protocol 2026-07-28 deprecates roots (a server may not ask for them), so a
	// client that sends roots is tested on the version that still has them.
	var opts *mcp.ClientSessionOptions
	if len(roots) > 0 {
		opts = &mcp.ClientSessionOptions{ProtocolVersion: "2025-06-18"}
	}
	cs, err := client.Connect(ctx, ct, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

func mcpCall(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return res
}

func resultText(res *mcp.CallToolResult, i int) string {
	if i >= len(res.Content) {
		return ""
	}
	return res.Content[i].(*mcp.TextContent).Text
}

func TestMCPInitialize(t *testing.T) {
	cs := mcpConnect(t, t.TempDir(), "")
	init := cs.InitializeResult()
	if init.ServerInfo.Name != "ajiya" {
		t.Errorf("server name = %q", init.ServerInfo.Name)
	}
	if init.Capabilities.Tools == nil {
		t.Error("server does not advertise tools")
	}
}

func TestMCPToolsListGolden(t *testing.T) {
	cs := mcpConnect(t, t.TempDir(), "")
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	got := marshalIndent(t, res.Tools) + "\n"
	path := filepath.Join("testdata", "mcp-tools.golden")
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("tools/list differs from %s; run go test ./internal/cli -run MCPToolsList -update and review the diff\n%s", path, got)
	}
}

// The tools that a person decides must never be offered over MCP. Adding one
// to mcpTools fails here; to offer it on purpose, change this test in the
// same commit and say why.
func TestMCPExcludedTools(t *testing.T) {
	want := []string{"ticket drop", "app remove", "milestone remove", "milestone move", "hook install"}
	if strings.Join(mcpExcluded, "|") != strings.Join(want, "|") {
		t.Fatalf("mcpExcluded = %v, want %v", mcpExcluded, want)
	}
	have := map[string]bool{}
	for _, tl := range mcpTools() {
		have[tl.name] = true
	}
	for _, c := range mcpExcluded {
		name := "ajiya_" + strings.ReplaceAll(c, " ", "_")
		if have[name] {
			t.Errorf("%s is excluded from MCP but the tool %s exists", c, name)
		}
	}
	// The complete list of tools, so an addition is a deliberate change.
	allowed := strings.Fields(`ajiya_status ajiya_next ajiya_ticket_show ajiya_ticket_list ajiya_phase_list
		ajiya_milestone_list ajiya_milestone_show ajiya_check ajiya_ticket_add ajiya_ticket_edit
		ajiya_ticket_start ajiya_ticket_done ajiya_ticket_block ajiya_phase_add ajiya_app_add
		ajiya_milestone_add ajiya_launch_set ajiya_init`)
	var got []string
	for n := range have {
		got = append(got, n)
	}
	sort.Strings(got)
	sort.Strings(allowed)
	if strings.Join(got, " ") != strings.Join(allowed, " ") {
		t.Errorf("tools = %v\nwant %v", got, allowed)
	}
	// Each tool runs a real command, and none is an excluded or unrelated one.
	for _, tl := range mcpTools() {
		argv := tl.argv(mcpArgs{})
		c, _ := find(argv)
		if c == nil {
			t.Errorf("%s: %v is not a command", tl.name, argv)
			continue
		}
		for _, x := range mcpExcluded {
			if c.name == x {
				t.Errorf("%s runs excluded command %q", tl.name, x)
			}
		}
		if strings.HasPrefix(c.name, "import ") || c.name == "serve" || c.name == "build" {
			t.Errorf("%s runs %q", tl.name, c.name)
		}
	}
}

func TestMCPNoProject(t *testing.T) {
	dir := t.TempDir() // no ajiya.toml here or above (temp dirs are outside the repository)
	cs := mcpConnect(t, dir, "")
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tl := range res.Tools {
		if tl.Name == "ajiya_init" {
			continue
		}
		args := map[string]any{}
		for _, p := range findTool(t, tl.Name).props {
			if p.required {
				args[p.name] = "x"
			}
		}
		r := mcpCall(t, cs, tl.Name, args)
		if !r.IsError || resultText(r, 0) != noProjectMsg {
			t.Errorf("%s: got error=%v %q, want the no-project message", tl.Name, r.IsError, resultText(r, 0))
		}
	}
	if noProjectMsg != "This folder is not set up for Ajiya. Run ajiya init, or call ajiya_init." {
		t.Errorf("message changed: %q", noProjectMsg)
	}
}

func findTool(t *testing.T, name string) mcpTool {
	t.Helper()
	for _, tl := range mcpTools() {
		if tl.name == name {
			return tl
		}
	}
	t.Fatalf("no tool %s", name)
	return mcpTool{}
}

func TestMCPInitThenPlan(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cs := mcpConnect(t, dir, "")
	r := mcpCall(t, cs, "ajiya_init", map[string]any{"name": "Demo", "prefix": "DE"})
	if r.IsError {
		t.Fatalf("init: %s", resultText(r, 0))
	}
	if _, err := os.Stat(filepath.Join(dir, "ajiya.toml")); err != nil {
		t.Fatal(err)
	}
	for _, step := range []struct {
		tool string
		args map[string]any
	}{
		{"ajiya_phase_add", map[string]any{"slug": "core", "goal": "Core works"}},
		{"ajiya_ticket_add", map[string]any{"phase": "core", "app": "infra", "title": "-leading dash title", "done_when": "it works"}},
		{"ajiya_ticket_edit", map[string]any{"id": "DE-0001", "done_when": "it really works"}},
		{"ajiya_milestone_add", map[string]any{"name": "alpha", "targets": "core"}},
		{"ajiya_ticket_start", map[string]any{"id": "DE-0001"}},
		{"ajiya_ticket_block", map[string]any{"id": "DE-0001", "reason": "waiting for a key"}},
		{"ajiya_launch_set", map[string]any{"target": "core"}},
	} {
		if r := mcpCall(t, cs, step.tool, step.args); r.IsError {
			t.Fatalf("%s: %s", step.tool, resultText(r, 0))
		}
	}
	r = mcpCall(t, cs, "ajiya_ticket_show", map[string]any{"id": "DE-0001"})
	if !strings.Contains(resultText(r, 1), "-leading dash title") {
		t.Errorf("ticket show: %v", resultText(r, 1))
	}
}

func TestMCPErrorsAreTheCLIMessage(t *testing.T) {
	dir := statusFixture(t)
	cs := mcpConnect(t, dir, "")
	r := mcpCall(t, cs, "ajiya_ticket_show", map[string]any{"id": "DE-9999"})
	if !r.IsError {
		t.Fatal("want a tool error")
	}
	var out strings.Builder
	e := &env{stdout: &out, stderr: &out, dir: dir}
	run(e, []string{"ticket", "show", "DE-9999"})
	if got, want := resultText(r, 0), strings.TrimSpace(out.String()); got != want || strings.Contains(got, "\n") {
		t.Errorf("error = %q, CLI prints %q", got, want)
	}

	// Argument mistakes are tool errors too, not protocol failures.
	r = mcpCall(t, cs, "ajiya_ticket_show", map[string]any{})
	if !r.IsError || !strings.Contains(resultText(r, 0), "missing required argument") {
		t.Errorf("missing arg: %v %q", r.IsError, resultText(r, 0))
	}
}

// ticket done keeps the CLI rule: no referencing commit, no done. The server
// never commits.
func TestMCPTicketDoneNeedsCommit(t *testing.T) {
	dir := statusFixture(t)
	cs := mcpConnect(t, dir, "")
	mcpCall(t, cs, "ajiya_ticket_start", map[string]any{"id": "DE-0002"})
	r := mcpCall(t, cs, "ajiya_ticket_done", map[string]any{"id": "DE-0001"})
	if !r.IsError || !strings.Contains(resultText(r, 0), "commit") {
		t.Errorf("done without a commit: %v %q", r.IsError, resultText(r, 0))
	}
	head := gitHead(t, dir)
	mcpCall(t, cs, "ajiya_ticket_done", map[string]any{"id": "DE-0001"})
	if gitHead(t, dir) != head {
		t.Error("the server made a commit")
	}
}

func gitHead(t *testing.T, dir string) string {
	t.Helper()
	var out strings.Builder
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

// Every read tool returns exactly the JSON of the CLI's --json.
func TestMCPReadToolParity(t *testing.T) {
	dir := statusFixture(t)
	cs := mcpConnect(t, dir, "")
	tests := []struct {
		tool string
		args map[string]any
		cli  []string
	}{
		{"ajiya_status", nil, []string{"status"}},
		{"ajiya_next", nil, []string{"next"}},
		{"ajiya_next", map[string]any{"milestone": "staging"}, []string{"next", "--milestone", "staging"}},
		{"ajiya_ticket_show", map[string]any{"id": "DE-0003"}, []string{"ticket", "show", "DE-0003"}},
		{"ajiya_ticket_list", nil, []string{"ticket", "list"}},
		{"ajiya_ticket_list", map[string]any{"phase": "api", "state": "in_progress"}, []string{"ticket", "list", "--phase", "api", "--state", "in_progress"}},
		{"ajiya_phase_list", nil, []string{"phase", "list"}},
		{"ajiya_milestone_list", nil, []string{"milestone", "list"}},
		{"ajiya_check", nil, []string{"check"}},
	}
	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			r := mcpCall(t, cs, tt.tool, tt.args)
			if r.IsError {
				t.Fatalf("error: %s", resultText(r, 0))
			}
			want := cliJSON(t, dir, tt.cli)
			if got := resultText(r, 1); got != want {
				t.Errorf("JSON differs from the CLI\ngot:\n%s\nwant:\n%s", got, want)
			}
			if s := resultText(r, 0); s == "" || strings.Contains(s, "\n") {
				t.Errorf("summary should be one line, got %q", s)
			}
		})
	}
}

func TestMCPMilestoneShowParity(t *testing.T) {
	dir := statusFixture(t)
	cs := mcpConnect(t, dir, "")
	list := cliJSON(t, dir, []string{"milestone", "list"})
	name := firstJSONName(t, list)
	r := mcpCall(t, cs, "ajiya_milestone_show", map[string]any{"name": name})
	if r.IsError {
		t.Fatalf("error: %s", resultText(r, 0))
	}
	if got, want := resultText(r, 1), cliJSON(t, dir, []string{"milestone", "show", name}); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

// --dir and client roots both find a project the working directory does not.
func TestMCPProjectSelection(t *testing.T) {
	proj := statusFixture(t)
	elsewhere := t.TempDir()
	want := cliJSON(t, proj, []string{"phase", "list"})

	for name, cs := range map[string]*mcp.ClientSession{
		"--dir":                 mcpConnect(t, elsewhere, proj),
		"roots":                 mcpConnect(t, elsewhere, "", proj),
		"cwd":                   mcpConnect(t, proj, ""),
		"cwd below the project": mcpConnect(t, mkSub(t, proj), ""),
	} {
		r := mcpCall(t, cs, "ajiya_phase_list", nil)
		if r.IsError || resultText(r, 1) != want {
			t.Errorf("%s: error=%v %q", name, r.IsError, resultText(r, 0))
		}
	}
}

func mkSub(t *testing.T, dir string) string {
	t.Helper()
	sub := filepath.Join(dir, "deep", "er")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	return sub
}

func TestFileURIPath(t *testing.T) {
	for _, tt := range []struct{ uri, want string }{
		{"file:///home/me/proj", "/home/me/proj"},
		{"file://localhost/home/me", "/home/me"},
		{"file:///C:/Users/me/proj", "C:/Users/me/proj"},
		{"file://C:/Users/me/proj", "C:/Users/me/proj"},
		{"file:///home/me/my%20proj", "/home/me/my proj"},
	} {
		got, ok := fileURIPath(tt.uri)
		if !ok || got != filepath.FromSlash(tt.want) {
			t.Errorf("fileURIPath(%q) = %q, %v; want %q", tt.uri, got, ok, filepath.FromSlash(tt.want))
		}
	}
	for _, bad := range []string{"https://x/y", "file://server/share/x", "::"} {
		if _, ok := fileURIPath(bad); ok {
			t.Errorf("fileURIPath(%q) accepted", bad)
		}
	}
	dir := t.TempDir()
	if got, ok := fileURIPath(fileURI(dir)); !ok || got != dir {
		t.Errorf("round trip of %q gave %q", dir, got)
	}
}
