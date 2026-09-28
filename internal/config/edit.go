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

// appBlock finds the [[apps]] table for name: the header line and the line
// after its last line (the next table header, or the end).
func appBlock(lines []string, name string) (start, end int, ok bool) {
	nameRE := regexp.MustCompile(`^\s*name\s*=\s*("` + regexp.QuoteMeta(name) + `"|'` + regexp.QuoteMeta(name) + `')\s*(#.*)?$`)
	start = -1
	for i := 0; i <= len(lines); i++ {
		header := i == len(lines)
		if !header {
			header = tableRE.MatchString(lines[i])
		}
		if header {
			if start >= 0 && ok {
				return start, i, true
			}
			start, ok = -1, false
			if i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "[[") && tableRE.FindStringSubmatch(lines[i])[1] == "apps" {
				start = i
			}
			continue
		}
		if start >= 0 && nameRE.MatchString(lines[i]) {
			ok = true
		}
	}
	return -1, -1, false
}

func splitLines(text string) []string {
	return strings.Split(strings.TrimSuffix(strings.ReplaceAll(text, "\r\n", "\n"), "\n"), "\n")
}

// AddAppText appends an [[apps]] table.
func AddAppText(text string, a App) string {
	block := fmt.Sprintf("\n[[apps]]\nname = %q\npath = %q\n", a.Name, a.Path)
	if a.Kind != "" {
		block += fmt.Sprintf("kind = %q\n", a.Kind)
	}
	return strings.TrimRight(text, "\n") + "\n" + block
}

// RemoveAppText deletes the [[apps]] table for name, with the blank lines before it.
func RemoveAppText(text, name string) (string, bool) {
	lines := splitLines(text)
	start, end, ok := appBlock(lines, name)
	if !ok {
		return text, false
	}
	// The table owns the blank lines after it; at the end of the file it takes
	// the blank lines before it instead.
	for end == len(lines) && start > 0 && strings.TrimSpace(lines[start-1]) == "" {
		start--
	}
	out := append(append([]string{}, lines[:start]...), lines[end:]...)
	if start == 0 && len(out) > 0 && strings.TrimSpace(out[0]) == "" {
		out = out[1:]
	}
	return strings.Join(out, "\n") + "\n", true
}

// SetAppPathText changes the path of the app called name.
func SetAppPathText(text, name, path string) (string, bool) {
	lines := splitLines(text)
	start, end, ok := appBlock(lines, name)
	if !ok {
		return text, false
	}
	pathRE := regexp.MustCompile(`^(\s*path\s*=\s*)("[^"]*"|'[^']*')(.*)$`)
	for i := start; i < end; i++ {
		if m := pathRE.FindStringSubmatch(lines[i]); m != nil {
			lines[i] = m[1] + fmt.Sprintf("%q", path) + m[3]
			return strings.Join(lines, "\n") + "\n", true
		}
	}
	lines = append(lines[:end], append([]string{fmt.Sprintf("path = %q", path)}, lines[end:]...)...)
	return strings.Join(lines, "\n") + "\n", true
}

// WriteText validates new ajiya.toml text and writes it to root.
func WriteText(root, text string) (*Config, error) {
	c, err := Parse([]byte(text))
	if err != nil {
		return nil, err
	}
	c.Root = root
	return c, os.WriteFile(filepath.Join(root, FileName), []byte(text), 0o644)
}

// ReadText returns root/ajiya.toml as text with LF line endings.
func ReadText(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, FileName))
	return strings.ReplaceAll(string(data), "\r\n", "\n"), err
}
