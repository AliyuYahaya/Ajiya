package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/gitx"
)

// hookNames are the git hooks ajiya installs.
var hookNames = []string{"commit-msg", "prepare-commit-msg"}

const (
	hookStart = "# ajiya:start (managed by 'ajiya hook install'; do not edit)"
	hookEnd   = "# ajiya:end"
)

// hookBlock is the part of a hook script that ajiya owns. It warns rather than
// fails when ajiya is missing, because the CI check backs the hook up.
func hookBlock(name string) string {
	return hookStart + `
if command -v ajiya >/dev/null 2>&1; then
  ajiya hook run ` + name + ` "$@" || exit $?
else
  echo "ajiya: not on PATH, so this commit was not checked against the ticket rule. Install it: https://github.com/AliyuYahaya/Ajiya" >&2
fi
` + hookEnd + "\n"
}

func runHookInstall(e *env, args []string) error {
	if _, err := parse(newFlags("hook install"), args, 0); err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	dir, err := gitx.GitPath(pr.cfg.Root, "hooks") // honours core.hooksPath
	if err != nil {
		return refused("%v", err)
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(pr.cfg.Root, dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, name := range hookNames {
		path := filepath.Join(dir, name)
		result, err := installHook(path, name)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(pr.cfg.Root, path)
		fmt.Fprintf(e.stdout, "%s %s\n", result, filepath.ToSlash(rel))
	}
	return nil
}

// installHook writes, updates or appends ajiya's block in one hook script.
func installHook(path, name string) (string, error) {
	block := hookBlock(name)
	old, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "Installed", writeHook(path, []byte("#!/bin/sh\n"+block))
	}
	if err != nil {
		return "", err
	}
	text := strings.ReplaceAll(string(old), "\r\n", "\n")
	if i := strings.Index(text, hookStart); i >= 0 {
		j := strings.Index(text[i:], hookEnd+"\n")
		if j < 0 {
			return "", refused("%s has '# ajiya:start' without '# ajiya:end'; fix or delete the block and run again", path)
		}
		updated := text[:i] + block + text[i+j+len(hookEnd)+1:]
		if updated == text {
			return "Up to date:", nil
		}
		return "Updated", writeHook(path, []byte(updated))
	}
	first, _, _ := strings.Cut(text, "\n")
	if !strings.HasPrefix(first, "#!") || !(strings.HasSuffix(first, "sh") || strings.Contains(first, "/sh ") || strings.Contains(first, "bash")) {
		return "", refused("%s is not a shell script (%q); add the line 'ajiya hook run %s \"$@\" || exit $?' to it yourself", path, first, name)
	}
	for line := range strings.SplitSeq(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "exec ") {
			return "", refused("%s ends in 'exec', so an appended block would never run; add the line 'ajiya hook run %s \"$@\" || exit $?' before it yourself", path, name)
		}
	}
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return "Added ajiya to", writeHook(path, []byte(text+"\n"+block))
}

func writeHook(path string, data []byte) error {
	if old, err := os.ReadFile(path); err == nil && bytes.Equal(old, data) {
		return nil
	}
	if err := os.WriteFile(path, data, 0o755); err != nil {
		return err
	}
	return os.Chmod(path, 0o755) // WriteFile keeps the mode of an existing file
}
