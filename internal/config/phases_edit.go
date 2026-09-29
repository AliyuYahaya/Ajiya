package config

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var orderKeyRE = regexp.MustCompile(`^(\s*order\s*=\s*)`)

// SetPhaseOrderText returns ajiya.toml text with [phases] order set, keeping
// the rest of the file, comments included, as it was. A long list is written
// one slug per line.
func SetPhaseOrderText(text string, order []string) string {
	lines := splitLines(text)
	value := formatOrder(order)
	inPhases, phasesAt := false, -1
	for i, l := range lines {
		if m := tableRE.FindStringSubmatch(l); m != nil {
			inPhases = m[1] == "phases" && !strings.HasPrefix(strings.TrimSpace(l), "[[")
			if inPhases {
				phasesAt = i
			}
			continue
		}
		if !inPhases {
			continue
		}
		m := orderKeyRE.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		endLine, endCol, ok := arrayEnd(lines, i, len(m[1]))
		if !ok {
			break // not an array we can read; fall back to adding nothing
		}
		rest := lines[endLine][endCol:]
		repl := strings.Split(m[1]+value+rest, "\n")
		out := append(append(append([]string{}, lines[:i]...), repl...), lines[endLine+1:]...)
		return strings.Join(out, "\n") + "\n"
	}
	if phasesAt >= 0 {
		add := strings.Split("order = "+value, "\n")
		lines = slices.Insert(lines, phasesAt+1, add...)
		return strings.Join(lines, "\n") + "\n"
	}
	// No [phases] table: add one before the first array table, or at the end.
	block := append([]string{"[phases]"}, strings.Split("order = "+value, "\n")...)
	for i, l := range lines {
		if tableRE.MatchString(l) && strings.HasPrefix(strings.TrimSpace(l), "[[") {
			lines = slices.Insert(lines, i, append(block, "")...)
			return strings.Join(lines, "\n") + "\n"
		}
	}
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
		lines = append(lines, "")
	}
	return strings.Join(append(lines, block...), "\n") + "\n"
}

// formatOrder writes a TOML array of strings: on one line when it is short.
func formatOrder(order []string) string {
	quoted := make([]string, len(order))
	for i, s := range order {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	one := "[" + strings.Join(quoted, ", ") + "]"
	if len(one) <= 70 {
		return one
	}
	return "[\n  " + strings.Join(quoted, ",\n  ") + ",\n]"
}

// arrayEnd finds the "]" closing the array that starts at column col of
// lines[start], skipping strings and comments. It returns the line and the
// column just past the bracket.
func arrayEnd(lines []string, start, col int) (line, end int, ok bool) {
	if col >= len(lines[start]) || lines[start][col] != '[' {
		return 0, 0, false
	}
	depth := 0
	for i := start; i < len(lines); i++ {
		l := lines[i]
		j := 0
		if i == start {
			j = col
		}
		for ; j < len(l); j++ {
			switch c := l[j]; c {
			case '#':
				j = len(l)
			case '"', '\'':
				k := j + 1
				for k < len(l) && l[k] != c {
					if c == '"' && l[k] == '\\' {
						k++
					}
					k++
				}
				j = k
			case '[':
				depth++
			case ']':
				depth--
				if depth == 0 {
					return i, j + 1, true
				}
			}
		}
	}
	return 0, 0, false
}

// SetPhaseOrder writes [phases] order into root/ajiya.toml, checking the
// result before writing it.
func SetPhaseOrder(root string, order []string) (*Config, error) {
	text, err := ReadText(root)
	if err != nil {
		return nil, err
	}
	text = SetPhaseOrderText(text, order)
	c, err := Parse([]byte(text))
	if err != nil {
		return nil, err
	}
	if !slices.Equal(c.Phases.Order, order) && !(len(order) == 0 && len(c.Phases.Order) == 0) {
		return nil, fmt.Errorf("could not set [phases] order in %s; set it by hand", FileName)
	}
	return WriteText(root, text)
}
