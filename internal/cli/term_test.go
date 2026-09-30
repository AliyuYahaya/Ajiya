package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIsTerminal: only a real terminal is interactive. A pipe, the null device
// and a regular file are not, on every OS.
func TestIsTerminal(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	null, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer null.Close()
	reg, err := os.Create(filepath.Join(t.TempDir(), "in.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer reg.Close()
	for name, f := range map[string]*os.File{"pipe read end": r, "pipe write end": w, "null device": null, "regular file": reg} {
		if isTerminal(f) {
			t.Errorf("%s counted as a terminal", name)
		}
	}
}
