//go:build windows

package cli

import (
	"os"

	"golang.org/x/sys/windows"
)

// isTerminal reports whether f is a console. GetConsoleMode fails for pipes,
// files and NUL.
func isTerminal(f *os.File) bool {
	var mode uint32
	return windows.GetConsoleMode(windows.Handle(f.Fd()), &mode) == nil
}
