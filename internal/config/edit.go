package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	tableRE  = regexp.MustCompile(`^\s*\[\[?\s*([A-Za-z0-9_.-]+)\s*\]\]?\s*(#.*)?$`)
	targetRE = regexp.MustCompile(`^(\s*target\s*=\s*)("[^"]*"|'[^']*')(.*)$`)
)

// SetLaunchText returns ajiya.toml text with [launch] target set, keeping the
// rest of the file, comments included, as it was.
func SetLaunchText(text, target string) string {
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	value := fmt.Sprintf("%q", target)
	inLaunch, launchAt := false, -1
	for i, l := range lines {
		if m := tableRE.FindStringSubmatch(l); m != nil {
			inLaunch = m[1] == "launch" && !strings.HasPrefix(strings.TrimSpace(l), "[[")
			if inLaunch {
				launchAt = i
			}
			continue
		}
		if inLaunch {
			if m := targetRE.FindStringSubmatch(l); m != nil {
				lines[i] = m[1] + value + m[3]
				return strings.Join(lines, "\n") + "\n"
			}
		}
	}
	if launchAt >= 0 {
		lines = append(lines[:launchAt+1], append([]string{"target = " + value}, lines[launchAt+1:]...)...)
		return strings.Join(lines, "\n") + "\n"
	}
	// No [launch] table: add one before the first [[apps]], or at the end.
	block := []string{"[launch]", "target = " + value, ""}
	for i, l := range lines {
		if m := tableRE.FindStringSubmatch(l); m != nil && strings.HasPrefix(strings.TrimSpace(l), "[[") {
			lines = append(lines[:i], append(block, lines[i:]...)...)
			return strings.Join(lines, "\n") + "\n"
		}
	}
	return strings.Join(append(lines, append([]string{""}, block[:2]...)...), "\n") + "\n"
}

// SetLaunchTarget writes the launch target into root/ajiya.toml.
func SetLaunchTarget(root, target string) error {
	path := filepath.Join(root, FileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := SetLaunchText(strings.ReplaceAll(string(data), "\r\n", "\n"), target)
	c, err := Parse([]byte(text))
	if err != nil {
		return err
	}
	if c.Launch.Target != target {
		return fmt.Errorf("could not set the launch target in %s; set [launch] target = %q by hand", FileName, target)
	}
	return os.WriteFile(path, []byte(text), 0o644)
}
