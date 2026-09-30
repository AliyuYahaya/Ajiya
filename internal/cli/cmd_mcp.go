package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// noProjectMsg is what every tool but ajiya_init answers outside a project.
const noProjectMsg = "This folder is not set up for Ajiya. Run ajiya init, or call ajiya_init."

// runMCP serves the MCP tools over stdio until the client disconnects. It
// listens on nothing else.
func runMCP(e *env, args []string) error {
	fs := newFlags("mcp")
	dir := fs.String("dir", "", "project folder (default: the nearest ajiya.toml above the working directory, or the client's roots)")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	fixed := ""
	if *dir != "" {
		abs, err := filepath.Abs(*dir)
		if err != nil {
			return usageErr("--dir %s: %v", *dir, err)
		}
		st, err := os.Stat(abs)
		if err != nil || !st.IsDir() {
			return usageErr("--dir %s is not a folder", *dir)
		}
		fixed = abs
	}
	srv := newMCPServer(e.dir, fixed, e.now)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil && ctx.Err() == nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

// mcpProp is one argument of a tool.
type mcpProp struct {
	name     string
	kind     string // "string" or "boolean"
	desc     string
	required bool
}

// mcpTool is one tool. It only turns its arguments into the command line of
// an existing CLI command and runs that command; it has no behaviour of its own.
type mcpTool struct {
	name     string
	desc     string
	readOnly bool
	json     bool // the command has --json: the tool returns that JSON too
	findings bool // ajiya_check: exit code 1 still carries a result
	props    []mcpProp
	argv     func(a mcpArgs) []string // the command line, without --json
}

type mcpArgs map[string]any

func (a mcpArgs) str(k string) string { s, _ := a[k].(string); return s }
func (a mcpArgs) on(k string) bool    { b, _ := a[k].(bool); return b }

// flag returns "--name=value" when the argument was given. The "=" form means
// a value that starts with "-" is never read as a flag.
func (a mcpArgs) flag(flagName, k string) []string {
	if s := a.str(k); s != "" {
		return []string{"--" + flagName + "=" + s}
	}
	return nil
}

func (a mcpArgs) sw(flagName, k string) []string {
	if a.on(k) {
		return []string{"--" + flagName}
	}
	return nil
}

func cat(head []string, groups ...[]string) []string {
	out := append([]string{}, head...)
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

// pos marks positional values, so one that starts with "-" is not a flag.
func pos(vals ...string) []string { return append([]string{"--"}, vals...) }

func strProp(name, desc string, required bool) mcpProp {
	return mcpProp{name, "string", desc, required}
}
func boolProp(name, desc string) mcpProp { return mcpProp{name, "boolean", desc, false} }

var idProp = strProp("id", "Ticket ID, such as AJ-0043.", true)

// mcpExcluded lists the commands that are never offered over MCP, because a
// person decides them. A test keeps this list and the tools apart.
var mcpExcluded = []string{
	"ticket drop", "app remove", "milestone remove", "milestone move", "hook install",
}

// mcpTools is the whole MCP surface, in the order tools/list returns it. Names
// are an API: do not rename them.
func mcpTools() []mcpTool {
	return []mcpTool{
		{name: "ajiya_status", readOnly: true, json: true,
			desc: "Use at the start of a session to see where the project stands: progress to each milestone, work in progress, what can start, checks and recent activity.",
			argv: func(a mcpArgs) []string { return []string{"status"} }},
		{name: "ajiya_next", readOnly: true, json: true,
			desc:  "Use to pick what to work on: lists the tickets that can start now (open, not blocked, every dependency done), best first.",
			props: []mcpProp{strProp("app", "Only tickets of this app.", false), strProp("milestone", "Only tickets this milestone needs.", false), boolProp("launch", "Only tickets the launch target needs.")},
			argv: func(a mcpArgs) []string {
				return cat([]string{"next"}, a.flag("app", "app"), a.flag("milestone", "milestone"), a.sw("launch", "launch"))
			}},
		{name: "ajiya_ticket_show", readOnly: true, json: true,
			desc:  "Use before starting a ticket, or when you need its done-when, what it waits on, what waits on it, or its commits.",
			props: []mcpProp{idProp},
			argv:  func(a mcpArgs) []string { return cat([]string{"ticket", "show"}, pos(a.str("id"))) }},
		{name: "ajiya_ticket_list", readOnly: true, json: true,
			desc:  "Use to find tickets by phase, app, state or milestone, in plan order.",
			props: []mcpProp{strProp("phase", "Phase slug.", false), strProp("app", "App name.", false), strProp("state", "pending, in_progress, done, dropped, ready or human.", false), strProp("milestone", "Milestone name.", false)},
			argv: func(a mcpArgs) []string {
				return cat([]string{"ticket", "list"}, a.flag("phase", "phase"), a.flag("app", "app"), a.flag("state", "state"), a.flag("milestone", "milestone"))
			}},
		{name: "ajiya_phase_list", readOnly: true, json: true,
			desc: "Use to see each phase with its progress and earliest milestone.",
			argv: func(a mcpArgs) []string { return []string{"phase", "list"} }},
		{name: "ajiya_milestone_list", readOnly: true, json: true,
			desc: "Use to see progress to each milestone gate, in order.",
			argv: func(a mcpArgs) []string { return []string{"milestone", "list"} }},
		{name: "ajiya_milestone_show", readOnly: true, json: true,
			desc:  "Use to see what one milestone still needs and what blocks it.",
			props: []mcpProp{strProp("name", "Milestone name.", true)},
			argv:  func(a mcpArgs) []string { return cat([]string{"milestone", "show"}, pos(a.str("name"))) }},
		{name: "ajiya_check", readOnly: true, json: true, findings: true,
			desc:  "Use before you finish a task to find problems in the plan; fix every error. An empty list means no problems.",
			props: []mcpProp{boolProp("strict", "Treat warnings as failures too."), strProp("commits", "Also check the commit rule on a revision range, such as origin/main..HEAD.", false)},
			argv: func(a mcpArgs) []string {
				return cat([]string{"check"}, a.sw("strict", "strict"), a.flag("commits", "commits"))
			}},
		{name: "ajiya_ticket_add",
			desc:  "Use when you find work that is not in the plan: add a ticket instead of widening the current one.",
			props: []mcpProp{strProp("title", "Ticket title.", true), strProp("phase", "Phase slug.", true), strProp("app", "App name.", true), strProp("done_when", "The check that proves the ticket is done.", true), strProp("depends", "Comma-separated IDs this ticket depends on.", false), strProp("human", "Reason, when only a person can do this ticket.", false)},
			argv: func(a mcpArgs) []string {
				return cat([]string{"ticket", "add"}, a.flag("phase", "phase"), a.flag("app", "app"), a.flag("done-when", "done_when"), a.flag("depends", "depends"), a.flag("human", "human"), pos(a.str("title")))
			}},
		{name: "ajiya_ticket_edit",
			desc:  "Use to correct a ticket's title, done-when, dependencies, app or phase. Give at least one field; depends replaces the whole list.",
			props: []mcpProp{idProp, strProp("title", "New title.", false), strProp("done_when", "New done-when.", false), strProp("depends", "New dependency list, comma separated, or \"-\" for none.", false), strProp("app", "New app.", false), strProp("phase", "New phase slug.", false)},
			argv: func(a mcpArgs) []string {
				return cat([]string{"ticket", "edit"}, a.flag("title", "title"), a.flag("done-when", "done_when"), a.flag("depends", "depends"), a.flag("app", "app"), a.flag("phase", "phase"), pos(a.str("id")))
			}},
		{name: "ajiya_ticket_start",
			desc:  "Use before you begin work on a ticket. It marks the ticket in progress and clears a block.",
			props: []mcpProp{idProp, strProp("note", "What is left, when resuming.", false)},
			argv: func(a mcpArgs) []string {
				return cat([]string{"ticket", "start"}, a.flag("note", "note"), pos(a.str("id")))
			}},
		{name: "ajiya_ticket_done",
			desc:  "Use when the ticket's done-when is met and a commit with its Ajiya trailer exists. It refuses without such a commit; this server never makes commits.",
			props: []mcpProp{idProp, boolProp("test", "Run the project's test command first and refuse if it fails."), strProp("note", "What was done.", false), strProp("by", "Name of the person who did it (for tickets that need a human).", false)},
			argv: func(a mcpArgs) []string {
				return cat([]string{"ticket", "done"}, a.sw("test", "test"), a.flag("note", "note"), a.flag("by", "by"), pos(a.str("id")))
			}},
		{name: "ajiya_ticket_block",
			desc:  "Use when a ticket is stuck on something outside your control, and say what it waits on. ajiya_ticket_start clears the block later.",
			props: []mcpProp{idProp, strProp("reason", "What the ticket waits on.", true)},
			argv: func(a mcpArgs) []string {
				return cat([]string{"ticket", "block"}, pos(a.str("id"), a.str("reason")))
			}},
		{name: "ajiya_phase_add",
			desc:  "Use when planning a project to add a phase, a stage of work with one goal.",
			props: []mcpProp{strProp("slug", "Short lowercase name used in file names.", true), strProp("goal", "What the phase achieves.", true), strProp("title", "Display title.", false)},
			argv: func(a mcpArgs) []string {
				return cat([]string{"phase", "add"}, a.flag("title", "title"), pos(a.str("slug"), a.str("goal")))
			}},
		{name: "ajiya_app_add",
			desc:  "Use in the same task that creates a new app folder, so tickets can name it. Create the folder first.",
			props: []mcpProp{strProp("name", "App name.", true), strProp("path", "Folder of the app, relative to the project.", true), boolProp("library", "The app is a shared package.")},
			argv: func(a mcpArgs) []string {
				return cat([]string{"app", "add"}, a.flag("path", "path"), a.sw("library", "library"), pos(a.str("name")))
			}},
		{name: "ajiya_milestone_add",
			desc:  "Use when planning to add a milestone: a gate that is reached when the phases or tickets it targets are done.",
			props: []mcpProp{strProp("name", "Milestone name.", true), strProp("targets", "Comma-separated phase slugs or ticket IDs it needs.", true), strProp("before", "Place before this milestone.", false), strProp("after", "Place after this milestone.", false)},
			argv: func(a mcpArgs) []string {
				return cat([]string{"milestone", "add"}, a.flag("targets", "targets"), a.flag("before", "before"), a.flag("after", "after"), pos(a.str("name")))
			}},
		{name: "ajiya_launch_set",
			desc:  "Use when planning to set the launch target: the phase or ticket that means the project has launched.",
			props: []mcpProp{strProp("target", "A phase slug or a ticket ID.", true)},
			argv:  func(a mcpArgs) []string { return cat([]string{"launch", "set"}, pos(a.str("target"))) }},
		{name: "ajiya_init",
			desc:  "Use in a folder that has no ajiya.toml to set up Ajiya there, registering the apps it detects. Equivalent to 'ajiya init --yes'.",
			props: []mcpProp{strProp("name", "Project name (default: the folder name).", false), strProp("prefix", "Ticket ID prefix, such as AJ.", false)},
			argv: func(a mcpArgs) []string {
				return cat([]string{"init", "--yes"}, a.flag("name", "name"), a.flag("prefix", "prefix"))
			}},
	}
}

func (t mcpTool) schema() map[string]any {
	props := map[string]any{}
	req := []string{}
	for _, p := range t.props {
		props[p.name] = map[string]any{"type": p.kind, "description": p.desc}
		if p.required {
			req = append(req, p.name)
		}
	}
	s := map[string]any{"type": "object", "properties": props, "additionalProperties": false}
	if len(req) > 0 {
		s["required"] = req
	}
	return s
}

func (t mcpTool) tool() *mcp.Tool {
	destructive := false
	return &mcp.Tool{
		Name:        t.name,
		Description: t.desc,
		InputSchema: t.schema(),
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: t.readOnly, DestructiveHint: &destructive},
	}
}

// parseArgs checks raw arguments against the tool's properties.
func (t mcpTool) parseArgs(raw json.RawMessage) (mcpArgs, error) {
	a := mcpArgs{}
	if s := bytes.TrimSpace(raw); len(s) > 0 && string(s) != "null" {
		if err := json.Unmarshal(s, &a); err != nil {
			return nil, fmt.Errorf("%s: arguments must be a JSON object", t.name)
		}
	}
	known := map[string]mcpProp{}
	for _, p := range t.props {
		known[p.name] = p
	}
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		p, ok := known[k]
		if !ok {
			return nil, fmt.Errorf("%s: unknown argument %q", t.name, k)
		}
		ok = false
		switch a[k].(type) {
		case string:
			ok = p.kind == "string"
		case bool:
			ok = p.kind == "boolean"
		}
		if !ok {
			return nil, fmt.Errorf("%s: argument %q must be a %s", t.name, k, p.kind)
		}
	}
	for _, p := range t.props {
		if p.required && a.str(p.name) == "" {
			return nil, fmt.Errorf("%s: missing required argument %q", t.name, p.name)
		}
	}
	return a, nil
}

// mcpServer holds what the tools share.
type mcpServer struct {
	cwd   string           // where the server was started
	fixed string           // --dir, "" when not given
	now   func() time.Time // the clock; nil means time.Now
	mu    sync.Mutex       // one command at a time, as in the CLI
}

// newMCPServer builds the server. cwd is where it was started and fixed is the
// --dir value ("" when not given).
func newMCPServer(cwd, fixed string, now func() time.Time) *mcp.Server {
	m := &mcpServer{cwd: cwd, fixed: fixed, now: now}
	srv := mcp.NewServer(&mcp.Implementation{Name: "ajiya", Version: Version}, &mcp.ServerOptions{
		Instructions: "Ajiya tracks a project's plan. Start with ajiya_status and ajiya_next, then ajiya_ticket_start. Commit with an 'Ajiya: <ID>' trailer yourself, then call ajiya_ticket_done and ajiya_check.",
	})
	for _, t := range mcpTools() {
		srv.AddTool(t.tool(), m.handler(t))
	}
	return srv
}

func textResult(isErr bool, parts ...string) *mcp.CallToolResult {
	r := &mcp.CallToolResult{IsError: isErr}
	for _, p := range parts {
		if p != "" {
			r.Content = append(r.Content, &mcp.TextContent{Text: p})
		}
	}
	return r
}

// firstLine is the first non-empty line of s.
func firstLine(s string) string {
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			return l
		}
	}
	return ""
}

func (m *mcpServer) handler(t mcpTool) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		a, err := t.parseArgs(req.Params.Arguments)
		if err != nil {
			return textResult(true, err.Error()), nil
		}
		dir, found := m.projectDir(ctx, req.Session, t.name == "ajiya_init")
		if !found && t.name != "ajiya_init" {
			return textResult(true, noProjectMsg), nil
		}
		m.mu.Lock()
		defer m.mu.Unlock()
		return m.call(t, a, dir), nil
	}
}

// call runs the CLI command behind a tool and shapes its output.
func (m *mcpServer) call(t mcpTool, a mcpArgs, dir string) *mcp.CallToolResult {
	argv := t.argv(a)
	out, errOut, code := m.exec(dir, argv)
	if code != 0 && !(t.findings && code == ExitRefused && errOut == "") {
		msg := strings.TrimSpace(errOut)
		if msg == "" {
			msg = fmt.Sprintf("ajiya %s: exited with code %d", strings.Join(argv[:min(2, len(argv))], " "), code)
		}
		return textResult(true, msg)
	}
	if !t.json {
		return textResult(false, firstLine(out), restLines(out), strings.TrimSpace(errOut))
	}
	// The one-line summary is the first line of the command's own text output.
	// Read commands do not change anything, so running twice is safe.
	jout, _, _ := m.exec(dir, append(append([]string{}, argv...), "--json"))
	plain := out
	if t.findings { // check prints the findings, then a count line
		lines := strings.Split(strings.TrimSpace(out), "\n")
		plain = lines[len(lines)-1]
	}
	return textResult(false, firstLine(plain), strings.TrimSpace(jout))
}

func restLines(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "\n"); i >= 0 {
		return strings.TrimSpace(s[i+1:])
	}
	return ""
}

// exec runs one CLI command in dir with captured output.
func (m *mcpServer) exec(dir string, argv []string) (stdout, stderr string, code int) {
	var o, e bytes.Buffer
	code = run(&env{stdin: strings.NewReader(""), stdout: &o, stderr: &e, dir: dir, now: m.now}, argv)
	return o.String(), e.String(), code
}

// projectDir picks the folder to work in and reports whether it holds an
// Ajiya project. Order: --dir, the working directory (nearest ajiya.toml
// above it), then the roots the client sent. For ajiya_init with no project,
// the folder to set up is --dir, the first root, or the working directory.
func (m *mcpServer) projectDir(ctx context.Context, sess *mcp.ServerSession, forInit bool) (string, bool) {
	if m.fixed != "" {
		_, err := config.Find(m.fixed)
		return m.fixed, err == nil
	}
	if _, err := config.Find(m.cwd); err == nil {
		return m.cwd, true
	}
	roots := clientRoots(ctx, sess)
	for _, r := range roots {
		if _, err := config.Find(r); err == nil {
			return r, true
		}
	}
	if forInit && len(roots) > 0 {
		return roots[0], false
	}
	return m.cwd, false
}

// clientRoots returns the local folders the client sent as MCP roots, if it
// sends any.
func clientRoots(ctx context.Context, sess *mcp.ServerSession) []string {
	if sess == nil {
		return nil
	}
	res, err := sess.ListRoots(ctx, nil)
	if err != nil || res == nil {
		return nil
	}
	var dirs []string
	for _, r := range res.Roots {
		p, ok := fileURIPath(r.URI)
		if !ok {
			continue
		}
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			dirs = append(dirs, p)
		}
	}
	return dirs
}

// fileURIPath turns a file:// URI into a local path. On Windows a URI such as
// file:///C:/Users/x has the path /C:/Users/x, so the slash before the drive
// letter is dropped; file://C:/Users/x (drive in the host part) is accepted too.
func fileURIPath(uri string) (string, bool) {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "file" {
		return "", false
	}
	p := u.Path
	if len(u.Host) == 2 && u.Host[1] == ':' { // file://C:/x
		p = u.Host + p
	} else if u.Host != "" && u.Host != "localhost" {
		return "", false // a network share is not a local folder
	}
	if len(p) >= 3 && p[0] == '/' && p[2] == ':' { // /C:/x
		p = p[1:]
	}
	return filepath.FromSlash(p), p != ""
}

// fileURI is the file:// URI for a local absolute path (file:///C:/x on Windows).
func fileURI(path string) string {
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return (&url.URL{Scheme: "file", Path: p}).String()
}
