// Package cli dispatches ajiya subcommands.
package cli

import (
	"fmt"
	"io"
)

// Exit codes shared by every command.
const (
	ExitOK      = 0 // success
	ExitRefused = 1 // a check failed or the command was refused
	ExitUsage   = 2 // usage or config error
)

// Version is set at build time with -ldflags "-X github.com/AliyuYahaya/Ajiya/internal/cli.Version=...".
var Version = "dev"

const usage = `usage: ajiya <command> [arguments]

Commands:
  version   print the version
`

// Run executes the command in args and returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return ExitUsage
	}
	switch args[0] {
	case "version", "--version":
		fmt.Fprintf(stdout, "ajiya %s\n", Version)
		return ExitOK
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return ExitOK
	default:
		fmt.Fprintf(stderr, "ajiya: unknown command %q; run 'ajiya help' for the list\n", args[0])
		return ExitUsage
	}
}
