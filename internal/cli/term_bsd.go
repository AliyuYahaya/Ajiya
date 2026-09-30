//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package cli

import (
	"os"

	"golang.org/x/sys/unix"
)

// isTerminal reports whether f is a real terminal. A character device is not
// enough (/dev/null is one): the terminal attributes must be readable.
func isTerminal(f *os.File) bool {
	_, err := unix.IoctlGetTermios(int(f.Fd()), unix.TIOCGETA)
	return err == nil
}
