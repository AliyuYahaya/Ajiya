package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/AliyuYahaya/Ajiya/internal/agents"
	"github.com/AliyuYahaya/Ajiya/internal/config"
)

// newAgentEnv is replaced in tests so that none of them can touch the real
// machine.
var newAgentEnv = agents.NewEnv

// agentEnv describes the machine for registering with agents. Tests replace it
// with a temporary home folder; it is never offered over MCP.
func (e *env) agentEnv() (*agents.Env, error) {
	if e.agents != nil {
		return e.agents, nil
	}
	root := e.dir
	if r, err := config.Find(e.dir); err == nil {
		root = r
	}
	ae, err := newAgentEnv(root)
	if err != nil {
		return nil, usageErr("%v", err)
	}
	return ae, nil
}

type installFlags struct {
	targets []agents.Agent
	scope   agents.Scope
	yes     bool
}

func parseInstallFlags(name string, args []string) (*installFlags, error) {
	fs := newFlags(name)
	claude := fs.Bool("claude", false, "Claude Code")
	codex := fs.Bool("codex", false, "Codex")
	cursor := fs.Bool("cursor", false, "Cursor")
	all := fs.Bool("all", false, "every supported agent")
	scope := fs.String("scope", "user", "user (every project) or project (this project only)")
	yes := fs.Bool("yes", false, "make the change without asking")
	if _, err := parse(fs, args, 0); err != nil {
		return nil, err
	}
	f := &installFlags{yes: *yes}
	switch *scope {
	case "user":
		f.scope = agents.User
	case "project":
		f.scope = agents.Project
	default:
		return nil, usageErr("--scope must be user or project, not %q", *scope)
	}
	for _, x := range []struct {
		on bool
		a  agents.Agent
	}{{*claude || *all, agents.Claude}, {*codex || *all, agents.Codex}, {*cursor || *all, agents.Cursor}} {
		if x.on {
			f.targets = append(f.targets, x.a)
		}
	}
	return f, nil
}

func runMCPInstall(e *env, args []string) error { return runMCPChange(e, "mcp install", args, true) }
func runMCPUninstall(e *env, args []string) error {
	return runMCPChange(e, "mcp uninstall", args, false)
}

func runMCPChange(e *env, name string, args []string, install bool) error {
	f, err := parseInstallFlags(name, args)
	if err != nil {
		return err
	}
	ae, err := e.agentEnv()
	if err != nil {
		return err
	}
	if len(f.targets) == 0 { // no agent named: every agent found on this machine
		f.targets = ae.Detect()
		if len(f.targets) == 0 {
			return refused("no supported agent found on this machine (Claude Code, Codex, Cursor); name one with --claude, --codex or --cursor, or use --all")
		}
	}
	return registerWith(e, ae, f.targets, f.scope, install, f.yes)
}

// registerWith plans the change for every agent, shows it, asks unless yes, and
// applies it. It writes nothing if any config cannot be read.
func registerWith(e *env, ae *agents.Env, targets []agents.Agent, scope agents.Scope, install, yes bool) error {
	var changes []*agents.Change
	var problems []string
	for _, a := range targets {
		c, err := ae.PlanInstall(a, scope)
		if !install {
			c, err = ae.PlanUninstall(a, scope)
		}
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", a.Title(), err))
			continue
		}
		changes = append(changes, c)
	}
	if len(problems) > 0 {
		return refused("%s", strings.Join(problems, "\n"))
	}
	pending := 0
	for _, c := range changes {
		fmt.Fprintf(e.stdout, "%s (%s scope): %s\n", c.Agent.Title(), c.Scope, c.Path)
		if c.Noop() {
			fmt.Fprintf(e.stdout, "  %s: nothing to change\n", c.Note)
			continue
		}
		pending++
		for _, l := range c.Preview {
			fmt.Fprintf(e.stdout, "  %s\n", l)
		}
	}
	if pending == 0 {
		return nil
	}
	if !yes {
		if !e.interactive {
			fmt.Fprintf(e.stdout, "Nothing was changed. Run the same command with --yes to make the change above.\n")
			return silent(ExitRefused)
		}
		fmt.Fprintf(e.stdout, "Make this change? A backup of each existing file is kept next to it as <file>%s. [y/N] ", agents.BackupSuffix)
		if a := strings.ToLower(strings.TrimSpace(e.readLine())); a != "y" && a != "yes" {
			fmt.Fprintln(e.stdout, "Nothing was changed.")
			return silent(ExitRefused)
		}
	}
	var registered []agents.Agent
	for _, c := range changes {
		if c.Noop() {
			continue
		}
		if err := c.Apply(); err != nil {
			return refused("%s: %v", c.Agent.Title(), err)
		}
		verb := map[string]string{"add": "Registered Ajiya with", "update": "Updated Ajiya's entry in", "remove": "Removed Ajiya from"}[c.Action]
		fmt.Fprintf(e.stdout, "%s %s.\n", verb, c.Agent.Title())
		if install {
			registered = append(registered, c.Agent)
		}
		if _, err := os.Stat(c.BackupPath()); err == nil {
			fmt.Fprintf(e.stdout, "  Backup: %s\n", c.BackupPath())
		}
	}
	if install {
		fmt.Fprintln(e.stdout, "Restart the agent (or start a new session) so it loads the new server.")
	}
	if len(registered) > 0 {
		fmt.Fprintf(e.stdout, "Undo with: %s\n", undoCommand(registered, scope))
	}
	return nil
}

// undoCommand is the one command that removes Ajiya again from the agents
// just registered.
func undoCommand(registered []agents.Agent, scope agents.Scope) string {
	cmd := "ajiya mcp uninstall"
	for _, a := range registered {
		cmd += " --" + string(a)
	}
	if scope == agents.Project {
		cmd += " --scope project"
	}
	return cmd
}

func runMCPStatus(e *env, args []string) error {
	fs := newFlags("mcp status")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	ae, err := e.agentEnv()
	if err != nil {
		return err
	}
	found := ae.Detect()
	if len(found) == 0 {
		fmt.Fprintln(e.stdout, "No supported agent found on this machine (Claude Code, Codex, Cursor).")
		return nil
	}
	w := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
	missing := 0
	for _, a := range found {
		for _, s := range []agents.Scope{agents.User, agents.Project} {
			p := ae.ConfigPath(a, s)
			if s == agents.Project {
				if _, err := os.Stat(p); err != nil {
					continue // no project config: nothing to report
				}
			}
			st, err := ae.Status(a, s)
			msg := st.String()
			switch {
			case err != nil:
				msg = "cannot read: " + firstLine(err.Error())
			case s == agents.User && st != agents.Registered:
				missing++
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", a.Title(), s, msg, p)
		}
	}
	w.Flush()
	if missing > 0 {
		fmt.Fprintf(e.stdout, "\nRegister with: %s\n", installHint(ae, found))
	}
	return nil
}

// installHint is the one command that registers Ajiya with the agents in
// found that do not have it (user scope).
func installHint(ae *agents.Env, found []agents.Agent) string {
	var flags []string
	for _, a := range found {
		if st, err := ae.Status(a, agents.User); err == nil && st != agents.Registered {
			flags = append(flags, "--"+string(a))
		}
	}
	if len(flags) == 0 {
		return ""
	}
	return "ajiya mcp install " + strings.Join(flags, " ")
}

// offerAgents is the last step of 'ajiya init': if an agent on this machine
// does not have Ajiya registered, register it with --yes, ask at a terminal,
// and otherwise print the one command that does it.
func offerAgents(e *env, yes bool) error {
	if e.noAgents {
		return nil
	}
	ae, err := e.agentEnv()
	if err != nil {
		return nil // no home folder: nothing to offer
	}
	var todo []agents.Agent
	var names []string
	for _, a := range ae.Detect() {
		if st, err := ae.Status(a, agents.User); err == nil && st != agents.Registered {
			todo = append(todo, a)
			names = append(names, a.Title())
		}
	}
	if len(todo) == 0 {
		return nil
	}
	fmt.Fprintf(e.stdout, "Ajiya is not registered with %s, so it cannot use Ajiya's tools yet.\n", strings.Join(names, ", "))
	if !yes && !e.interactive {
		fmt.Fprintf(e.stdout, "To register it, run: %s\n", installHint(ae, todo))
		return nil
	}
	if err := registerWith(e, ae, todo, agents.User, true, yes); err != nil {
		var s silent
		if errors.As(err, &s) {
			return nil // the person said no
		}
		fmt.Fprintf(e.stdout, "Not registered: %v\nTry again with: %s\n", err, installHint(ae, todo))
	}
	return nil
}
