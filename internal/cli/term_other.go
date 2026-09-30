//go:build !linux && !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !windows

package cli

import "os"

// isTerminal is false where there is no terminal check: not asking is safe,
// the command is printed instead.
func isTerminal(*os.File) bool { return false }
