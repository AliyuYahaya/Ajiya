// Package agents registers Ajiya's MCP server (`ajiya mcp`) with AI coding
// agents: Claude Code, Codex and Cursor.
//
// Config locations and formats, checked against each agent's documentation
// (2026-09):
//
//   - Claude Code, https://code.claude.com/docs/en/mcp
//     `claude mcp add --scope user|project <name> -- <command> [args...]` and
//     `claude mcp remove <name> --scope user|project`. User scope is stored in
//     ~/.claude.json (top-level "mcpServers"). When CLAUDE_CONFIG_DIR is set it
//     moves with the rest of the configuration: $CLAUDE_CONFIG_DIR/.claude.json
//     (https://code.claude.com/docs/en/env-vars: "Override the directory where
//     Claude Code stores configuration"; https://code.claude.com/docs/en/claude-directory:
//     "If you set CLAUDE_CONFIG_DIR, every ~/.claude path on this page lives under
//     that directory instead". The docs never spell out the .claude.json case,
//     so the location is Claude Code's known behaviour, not a quoted sentence).
//     `claude` reads the variable itself, so it is passed to it. Project scope is in .mcp.json at the
//     project root ({"mcpServers": {"<name>": {"type": "stdio", "command": ...,
//     "args": [...]}}}). The agent's own command is preferred; when `claude` is
//     not on PATH the same two files are edited directly.
//   - Codex, https://developers.openai.com/codex/mcp (redirects to
//     https://learn.chatgpt.com/docs/extend/mcp?surface=cli)
//     ~/.codex/config.toml (CODEX_HOME overrides ~/.codex) or, for trusted
//     projects, .codex/config.toml, with a [mcp_servers.<name>] table holding
//     command and args. `codex mcp add` exists, but it would rewrite the whole
//     file and can drop comments, so the table is added and removed textually.
//   - Cursor, https://cursor.com/docs/context/mcp
//     ~/.cursor/mcp.json (global) or .cursor/mcp.json (project):
//     {"mcpServers": {"<name>": {"type": "stdio", "command": ..., "args": [...]}}}.
package agents

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// ServerName is the name of the entry in every agent's config.
const ServerName = "ajiya"

// BackupSuffix is added to a file's name for the copy kept before an edit.
const BackupSuffix = ".ajiya-backup"

// Agent is one supported AI coding agent.
type Agent string

const (
	Claude Agent = "claude"
	Codex  Agent = "codex"
	Cursor Agent = "cursor"
)

// All lists every supported agent.
var All = []Agent{Claude, Codex, Cursor}

// Title is the name shown to people.
func (a Agent) Title() string {
	switch a {
	case Claude:
		return "Claude Code"
	case Codex:
		return "Codex"
	}
	return "Cursor"
}

// Scope says whose config an entry goes in.
type Scope string

const (
	User    Scope = "user"    // every project of this person
	Project Scope = "project" // this project only
)

// Runner runs a command in dir and returns its combined output.
type Runner func(dir, name string, args ...string) (string, error)

// Env is everything the code needs from the machine, so tests can replace it.
type Env struct {
	Home      string // the person's home folder
	CodexHome string // Codex's folder ($CODEX_HOME, else Home/.codex)
	// ClaudeConfigDir is $CLAUDE_CONFIG_DIR; empty means Claude Code's
	// defaults (~/.claude and ~/.claude.json).
	ClaudeConfigDir string
	Dir             string // the project folder, for project scope
	Binary          string // what the entry runs: an absolute path, or "ajiya"
	LookPath        func(string) (string, error)
	Run             Runner
}

// NewEnv describes this machine and the running program.
func NewEnv(projectDir string) (*Env, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot find your home folder: %v", err)
	}
	codexHome := os.Getenv("CODEX_HOME")
	if codexHome == "" {
		codexHome = filepath.Join(home, ".codex")
	}
	claudeDir := os.Getenv("CLAUDE_CONFIG_DIR")
	return &Env{Home: home, CodexHome: codexHome, ClaudeConfigDir: claudeDir, Dir: projectDir, Binary: ResolveBinary(exec.LookPath, os.Executable), LookPath: exec.LookPath, Run: runner(claudeDir)}, nil
}

// runner runs commands with the process environment, and with CLAUDE_CONFIG_DIR
// set to claudeDir when there is one, so `claude` edits the same file
// ConfigPath reports.
func runner(claudeDir string) Runner {
	return func(dir, name string, args ...string) (string, error) {
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		if claudeDir != "" {
			cmd.Env = append(os.Environ(), "CLAUDE_CONFIG_DIR="+claudeDir)
		}
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
}

func execRun(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// ResolveBinary decides what the config entry runs.
//
// When the running program is the `ajiya` found on PATH, the absolute path of
// that PATH entry is used, not the resolved one: agents started from a desktop
// launcher often have a short PATH, and a Homebrew-style symlink survives
// upgrades where the version folder it points to does not. Otherwise the
// running program's own absolute path is used, unless it lives in a temporary
// folder (go run, a test build), where the plain name `ajiya` is the only
// thing that can work later.
func ResolveBinary(lookPath func(string) (string, error), executable func() (string, error)) string {
	exe, err := executable()
	if err != nil || exe == "" {
		return "ajiya"
	}
	if p, err := lookPath("ajiya"); err == nil {
		if abs, err := filepath.Abs(p); err == nil && sameFile(abs, exe) {
			return abs
		}
	}
	if isTemp(exe) {
		return "ajiya"
	}
	return exe
}

func sameFile(a, b string) bool {
	sa, err1 := os.Stat(a)
	sb, err2 := os.Stat(b)
	return err1 == nil && err2 == nil && os.SameFile(sa, sb)
}

func isTemp(p string) bool {
	p = filepath.ToSlash(p)
	if strings.Contains(p, "/go-build") {
		return true
	}
	tmp := filepath.ToSlash(os.TempDir())
	return tmp != "" && tmp != "/" && strings.HasPrefix(p, strings.TrimSuffix(tmp, "/")+"/")
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func (e *Env) has(cmd string) bool {
	if e.LookPath == nil {
		return false
	}
	_, err := e.LookPath(cmd)
	return err == nil
}

// Detect lists the agents found on this machine: their command is on PATH or
// their config folder exists.
func (e *Env) Detect() []Agent {
	var out []Agent
	for _, a := range All {
		if e.found(a) {
			out = append(out, a)
		}
	}
	return out
}

func (e *Env) found(a Agent) bool {
	switch a {
	case Claude:
		if e.ClaudeConfigDir != "" {
			return e.has("claude") || exists(e.ClaudeConfigDir)
		}
		return e.has("claude") || exists(filepath.Join(e.Home, ".claude")) || exists(filepath.Join(e.Home, ".claude.json"))
	case Codex:
		return e.has("codex") || exists(e.CodexHome)
	}
	return e.has("cursor") || exists(filepath.Join(e.Home, ".cursor"))
}

// ConfigPath is the file that holds an agent's entry for a scope.
func (e *Env) ConfigPath(a Agent, s Scope) string {
	switch a {
	case Claude:
		if s == Project {
			return filepath.Join(e.Dir, ".mcp.json")
		}
		if e.ClaudeConfigDir != "" {
			return filepath.Join(e.ClaudeConfigDir, ".claude.json")
		}
		return filepath.Join(e.Home, ".claude.json")
	case Codex:
		if s == Project {
			return filepath.Join(e.Dir, ".codex", "config.toml")
		}
		return filepath.Join(e.CodexHome, "config.toml")
	}
	if s == Project {
		return filepath.Join(e.Dir, ".cursor", "mcp.json")
	}
	return filepath.Join(e.Home, ".cursor", "mcp.json")
}

// State is whether Ajiya is registered in one config.
type State int

const (
	NotRegistered State = iota
	Registered          // an entry that runs `ajiya mcp`
	Different           // an entry named ajiya that does not run `ajiya mcp` (or points at a missing file)
)

func (s State) String() string {
	switch s {
	case Registered:
		return "registered"
	case Different:
		return "an entry named ajiya exists but does not run 'ajiya mcp'"
	}
	return "not registered"
}

// entry is one server as the agents store it.
type entry struct {
	Command string
	Args    []string
}

// classify says whether an existing entry already runs `ajiya mcp`. Any
// ajiya binary counts, so an entry someone wrote by hand is left alone; a path
// to a file that no longer exists does not.
func classify(en entry) State {
	if len(en.Args) == 0 || en.Args[0] != "mcp" {
		return Different
	}
	cmd := strings.ReplaceAll(en.Command, `\`, "/")
	base := strings.TrimSuffix(strings.ToLower(path.Base(cmd)), ".exe")
	if base != "ajiya" {
		return Different
	}
	if !strings.ContainsAny(en.Command, `/\`) { // bare "ajiya": found on PATH
		return Registered
	}
	if exists(en.Command) {
		return Registered
	}
	return Different
}

// Status reads one config. A missing file is NotRegistered; a file that cannot
// be read as its format is an error.
func (e *Env) Status(a Agent, s Scope) (State, error) {
	p := e.ConfigPath(a, s)
	data, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return NotRegistered, nil
	}
	if err != nil {
		return NotRegistered, err
	}
	en, ok, err := lookup(a, p, data)
	if err != nil {
		return NotRegistered, err
	}
	if !ok {
		return NotRegistered, nil
	}
	return classify(en), nil
}

func lookup(a Agent, p string, data []byte) (entry, bool, error) {
	if a == Codex {
		return tomlLookup(p, data)
	}
	return jsonLookup(p, data)
}

// Change is one planned edit. Nothing has been written until Apply.
type Change struct {
	Agent   Agent
	Scope   Scope
	Path    string     // the config file
	Action  string     // "add", "update", "remove" or "" for nothing to do
	Note    string     // why there is nothing to do
	Preview []string   // diff lines to show
	Cmds    [][]string // the agent's own commands to run (in the project folder), or nil to write the file
	env     *Env
	old     []byte // the file before, nil when it did not exist
	updated []byte // the file after, when Cmds is nil
}

// Noop is true when there is nothing to change.
func (c *Change) Noop() bool { return c.Action == "" }

func (e *Env) entryFor() entry { return entry{Command: e.Binary, Args: []string{"mcp"}} }

// PlanInstall works out what registering Ajiya would change.
func (e *Env) PlanInstall(a Agent, s Scope) (*Change, error) { return e.plan(a, s, true) }

// PlanUninstall works out what removing Ajiya's entry would change.
func (e *Env) PlanUninstall(a Agent, s Scope) (*Change, error) { return e.plan(a, s, false) }

func (e *Env) plan(a Agent, s Scope, install bool) (*Change, error) {
	p := e.ConfigPath(a, s)
	c := &Change{Agent: a, Scope: s, Path: p, env: e}
	old, err := os.ReadFile(p)
	switch {
	case errors.Is(err, os.ErrNotExist):
		old = nil
	case err != nil:
		return nil, fmt.Errorf("%s: %v", p, err)
	default:
		c.old = old
	}
	cur, ok, err := lookup(a, p, old)
	if err != nil {
		return nil, err
	}
	state := NotRegistered
	if ok {
		state = classify(cur)
	}
	if install {
		if state == Registered {
			c.Note = "already registered"
			return c, nil
		}
		c.Action = "add"
		if state == Different {
			c.Action = "update"
		}
	} else {
		if !ok {
			c.Note = "not registered"
			return c, nil
		}
		c.Action = "remove"
	}
	if a == Claude && e.has("claude") {
		c.planClaudeCLI(s)
		return c, nil
	}
	switch a {
	case Codex:
		c.updated, c.Preview, err = tomlEdit(p, old, e.entryFor(), install, ok)
	default:
		c.updated, c.Preview, err = jsonEdit(p, old, e.entryFor(), install, ok)
	}
	return c, err
}

// planClaudeCLI uses `claude mcp add` and `claude mcp remove`, the documented way.
func (c *Change) planClaudeCLI(s Scope) {
	scope := string(s)
	remove := []string{"claude", "mcp", "remove", ServerName, "--scope", scope}
	add := []string{"claude", "mcp", "add", "--scope", scope, ServerName, "--", c.env.Binary, "mcp"}
	switch c.Action {
	case "remove":
		c.Cmds = [][]string{remove}
	case "update":
		c.Cmds = [][]string{remove, add}
	default:
		c.Cmds = [][]string{add}
	}
	for _, cmd := range c.Cmds {
		c.Preview = append(c.Preview, "$ "+shellJoin(cmd))
	}
	c.Preview = append(c.Preview, "(Claude Code edits "+c.Path+" itself)")
}

func shellJoin(args []string) string {
	q := make([]string, len(args))
	for i, a := range args {
		if a == "" || strings.ContainsAny(a, " \t\"'\\$") {
			a = `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(a) + `"`
		}
		q[i] = a
	}
	return strings.Join(q, " ")
}

// BackupPath is where the copy of the file is kept.
func (c *Change) BackupPath() string { return c.Path + BackupSuffix }

// Apply keeps a backup of the file, then makes the change.
func (c *Change) Apply() error {
	if c.Noop() {
		return nil
	}
	if c.old != nil {
		mode := os.FileMode(0o600)
		if st, err := os.Stat(c.Path); err == nil {
			mode = st.Mode().Perm()
		}
		if err := os.WriteFile(c.BackupPath(), c.old, mode); err != nil {
			return fmt.Errorf("backup %s: %v", c.BackupPath(), err)
		}
	}
	if c.Cmds != nil {
		run := c.env.Run
		if run == nil {
			run = execRun
		}
		for _, cmd := range c.Cmds {
			if out, err := run(c.env.Dir, cmd[0], cmd[1:]...); err != nil {
				return fmt.Errorf("%s: %v\n%s", shellJoin(cmd), err, strings.TrimSpace(out))
			}
		}
		return nil
	}
	return writeFile(c.Path, c.updated, c.old != nil)
}

// writeFile replaces path atomically, keeping an existing file's permissions.
func writeFile(p string, data []byte, existed bool) error {
	mode := os.FileMode(0o644)
	if existed {
		if st, err := os.Stat(p); err == nil {
			mode = st.Mode().Perm()
		}
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(p), filepath.Base(p)+".tmp*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	_, werr := tmp.Write(data)
	cerr := tmp.Close()
	if werr != nil || cerr != nil {
		os.Remove(name)
		return errors.Join(werr, cerr)
	}
	if err := os.Chmod(name, mode); err != nil {
		os.Remove(name)
		return err
	}
	if err := os.Rename(name, p); err != nil {
		os.Remove(name)
		return err
	}
	return nil
}

func lines(b []byte) []string {
	s := strings.ReplaceAll(string(bytes.TrimRight(b, "\r\n")), "\r\n", "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
