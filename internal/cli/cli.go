// Package cli dispatches ajiya subcommands.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Exit codes shared by every command.
const (
	ExitOK      = 0 // success
	ExitRefused = 1 // a check failed or the command was refused
	ExitUsage   = 2 // usage or config error
)

// Version is set at build time with -ldflags "-X github.com/AliyuYahaya/Ajiya/internal/cli.Version=...".
var Version = "dev"

// exitError carries a message and the exit code it should produce.
type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string { return e.msg }

// refused is a command that ran but would not do what was asked (exit 1).
func refused(format string, a ...any) error {
	return &exitError{ExitRefused, fmt.Sprintf(format, a...)}
}

// usageErr is a mistake in how the command was called, or in the config (exit 2).
func usageErr(format string, a ...any) error {
	return &exitError{ExitUsage, fmt.Sprintf(format, a...)}
}

// silent exits with a code without printing anything more.
type silent int

func (s silent) Error() string { return "" }

type env struct {
	stdout, stderr io.Writer
	dir            string // working directory
}

type command struct {
	name    string // "ticket add"
	args    string // what follows the name in the usage line
	summary string
	run     func(e *env, args []string) error
}

var commands []command

func init() {
	commands = []command{
		{"init", "[--name <name>] [--prefix <PREFIX>]", "Set up ajiya in this directory", runInit},
		{"app add", "<name> --path <dir> [--library]", "Register an app", runAppAdd},
		{"app move", "<name> --path <dir>", "Change an app's folder", runAppMove},
		{"app remove", "<name> [--to <app>]", "Unregister an app, moving its open tickets", runAppRemove},
		{"phase add", `<slug> "<goal>" [--title "<title>"]`, "Add a phase", runPhaseAdd},
		{"ticket add", `--phase <slug> --app <app> "<title>" --done-when "<check>" [--depends <IDs>] [--human "<reason>"]`, "Add a ticket", runTicketAdd},
		{"ticket start", `<ID> [--note "<what is left>"]`, "Mark a ticket in progress", runTicketStart},
		{"ticket done", `<ID> [--note "<note>"] [--by "<name>"]`, "Mark a ticket done, with evidence", runTicketDone},
		{"ticket block", `<ID> "<reason>"`, "Mark a ticket blocked, with what it waits on", runTicketBlock},
		{"ticket drop", `<ID> --reason "<why>" --by "<name>"`, "Close a ticket that is not needed (a person's decision)", runTicketDrop},
		{"ticket show", "<ID>", "Show a ticket, its dependencies and dependants", runTicketShow},
		{"ticket edit", `<ID> [--title] [--done-when] [--depends] [--app] [--phase]`, "Change a ticket", runTicketEdit},
		{"next", "[--app <app>] [--launch]", "List tickets that can start now", runNext},
		{"launch set", "<phase|ticket>", "Set the launch target", runLaunchSet},
		{"launch show", "", "Show the launch target and what it still needs", runLaunchShow},
		{"check", "[--strict] [--commits <range>]", "Check the plan for problems", runCheck},
		{"hook install", "", "Install the commit-msg and prepare-commit-msg git hooks", runHookInstall},
		{"hook run", "<hook> <git hook arguments>", "Run a git hook (called by the installed hook scripts)", runHookRun},
		{"version", "", "Print the version", runVersion},
	}
}

func usageText() string {
	var b strings.Builder
	b.WriteString("usage: ajiya <command> [arguments]\n\nCommands:\n")
	for _, c := range commands {
		fmt.Fprintf(&b, "  %-13s %s\n", c.name, c.summary)
	}
	b.WriteString("\nRun 'ajiya <command> --help' for the arguments of one command.\n")
	return b.String()
}

// Run executes the command in args and returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "ajiya: %v\n", err)
		return ExitUsage
	}
	return run(&env{stdout: stdout, stderr: stderr, dir: dir}, args)
}

func run(e *env, args []string) int {
	if len(args) == 0 {
		fmt.Fprint(e.stderr, usageText())
		return ExitUsage
	}
	switch args[0] {
	case "help", "-h", "--help":
		fmt.Fprint(e.stdout, usageText())
		return ExitOK
	case "--version":
		args = []string{"version"}
	}
	c, rest := find(args)
	if c == nil {
		fmt.Fprintf(e.stderr, "ajiya: unknown command %q; run 'ajiya help' for the list\n", strings.Join(args[:min(2, len(args))], " "))
		return ExitUsage
	}
	err := c.run(e, rest)
	var ee *exitError
	var s silent
	switch {
	case err == nil:
		return ExitOK
	case errors.Is(err, flag.ErrHelp):
		fmt.Fprintf(e.stdout, "usage: ajiya %s %s\n\n%s.\n", c.name, c.args, c.summary)
		return ExitOK
	case errors.As(err, &s):
		return int(s)
	case errors.As(err, &ee):
		fmt.Fprintf(e.stderr, "ajiya %s: %s\n", c.name, ee.msg)
		return ee.code
	default:
		fmt.Fprintf(e.stderr, "ajiya %s: %v\n", c.name, err)
		return ExitRefused
	}
}

// find matches the longest command name at the start of args.
func find(args []string) (*command, []string) {
	for _, n := range []int{2, 1} {
		if len(args) < n {
			continue
		}
		name := strings.Join(args[:n], " ")
		for i := range commands {
			if commands[i].name == name {
				return &commands[i], args[n:]
			}
		}
	}
	return nil, nil
}

func runVersion(e *env, args []string) error {
	if len(args) > 0 {
		return usageErr("takes no arguments")
	}
	fmt.Fprintf(e.stdout, "ajiya %s\n", Version)
	return nil
}
