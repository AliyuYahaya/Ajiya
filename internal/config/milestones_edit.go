package config

import (
	"fmt"
	"regexp"
	"strings"
)

// Text edits of the [[milestones]] tables in ajiya.toml. Each keeps the other
// tables and the comments as they were. A milestone table owns the comment
// lines directly above its header (no blank line between), so moving or
// removing it takes them along.

var (
	nameKeyRE    = regexp.MustCompile(`^\s*name\s*=\s*("([^"]*)"|'([^']*)')\s*(#.*)?$`)
	targetsKeyRE = regexp.MustCompile(`^\s*targets\s*=`)
)

// CheckMilestoneName reports whether name can name a milestone.
func CheckMilestoneName(name string) error {
	if !milestoneRE.MatchString(name) {
		return fmt.Errorf("milestone name %q must be lower case letters and digits separated by single hyphens, like staging-proven", name)
	}
	return nil
}

// mblock is one [[milestones]] table: lines[start:end], header at head.
type mblock struct {
	name             string
	start, head, end int
}

func isComment(l string) bool { return strings.HasPrefix(strings.TrimSpace(l), "#") }
func isBlank(l string) bool   { return strings.TrimSpace(l) == "" }

// isArrayTable reports whether line is the header [[name]].
func isArrayTable(line, name string) bool {
	m := tableRE.FindStringSubmatch(line)
	return m != nil && m[1] == name && strings.HasPrefix(strings.TrimSpace(line), "[[")
}

// isPlainTable reports whether line is the header [name].
func isPlainTable(line, name string) bool {
	m := tableRE.FindStringSubmatch(line)
	return m != nil && m[1] == name && !strings.HasPrefix(strings.TrimSpace(line), "[[")
}

// leadStart returns where the comments directly above the header at i begin.
func leadStart(lines []string, i int) int {
	for i > 0 && isComment(lines[i-1]) {
		i--
	}
	return i
}

// tableEnd returns the end of the table whose header is at head: the start of
// the next table (with its lead comments), less trailing blank lines.
func tableEnd(lines []string, head int) int {
	end := len(lines)
	for i := head + 1; i < len(lines); i++ {
		if tableRE.MatchString(lines[i]) {
			end = leadStart(lines, i)
			break
		}
	}
	for end > head+1 && isBlank(lines[end-1]) {
		end--
	}
	return end
}

func milestoneBlocks(lines []string) []mblock {
	var out []mblock
	for i, l := range lines {
		if !isArrayTable(l, "milestones") {
			continue
		}
		b := mblock{start: leadStart(lines, i), head: i, end: tableEnd(lines, i)}
		for _, kl := range lines[i+1 : b.end] {
			if m := nameKeyRE.FindStringSubmatch(kl); m != nil {
				b.name = m[2] + m[3]
				break
			}
		}
		out = append(out, b)
	}
	return out
}

func findBlock(bs []mblock, name string) (mblock, bool) {
	for _, b := range bs {
		if b.name == name {
			return b, true
		}
	}
	return mblock{}, false
}

func joinLines(lines []string) string {
	for len(lines) > 0 && isBlank(lines[0]) {
		lines = lines[1:]
	}
	for len(lines) > 0 && isBlank(lines[len(lines)-1]) {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n") + "\n"
}

func quoteList(items []string) string {
	q := make([]string, len(items))
	for i, s := range items {
		q[i] = fmt.Sprintf("%q", s)
	}
	return "[" + strings.Join(q, ", ") + "]"
}

func milestoneTable(m Milestone) []string {
	return []string{"[[milestones]]", fmt.Sprintf("name = %q", m.Name), "targets = " + quoteList(m.Targets)}
}

// ConvertLaunchText rewrites the older [launch] target = "x" as a
// [[milestones]] table named launch with targets = ["x"], in the same place.
// Comments in the table stay. It reports false when there is no [launch]
// target to convert.
func ConvertLaunchText(text string) (string, bool) {
	lines := splitLines(text)
	for i, l := range lines {
		if !isPlainTable(l, "launch") {
			continue
		}
		end := tableEnd(lines, i)
		for j := i + 1; j < end; j++ {
			m := targetRE.FindStringSubmatch(lines[j])
			if m == nil {
				continue
			}
			target := strings.Trim(m[2], `"'`)
			if target == "" {
				return text, false
			}
			out := append([]string{}, lines[:i]...)
			out = append(out, "[[milestones]]", fmt.Sprintf("name = %q", LaunchMilestone))
			out = append(out, lines[i+1:j]...)
			out = append(out, "targets = "+quoteList([]string{target})+m[3])
			out = append(out, lines[j+1:]...)
			return joinLines(out), true
		}
		return text, false
	}
	return text, false
}

// insertAt puts a table at line i (a table start), with a blank line after it.
func insertAt(lines []string, i int, table []string) []string {
	out := append([]string{}, lines[:i]...)
	out = append(out, table...)
	out = append(out, "")
	return append(out, lines[i:]...)
}

// insertAfter puts a table after line i-1 (a table end), with a blank line before it.
func insertAfter(lines []string, i int, table []string) []string {
	out := append([]string{}, lines[:i]...)
	out = append(out, "")
	out = append(out, table...)
	if i < len(lines) && !isBlank(lines[i]) {
		out = append(out, "")
	}
	return append(out, lines[i:]...)
}

// placeTable inserts table among the milestones: before or after the one
// named (at most one of the two is set), or else before launch when there is
// a launch milestone and this is not it, or else after the last milestone.
// With no milestones yet it goes before the first [[apps]] table, or at the
// end.
func placeTable(lines []string, table []string, name, before, after string) ([]string, error) {
	bs := milestoneBlocks(lines)
	switch {
	case before != "":
		b, ok := findBlock(bs, before)
		if !ok {
			return nil, fmt.Errorf("no milestone %q", before)
		}
		return insertAt(lines, b.start, table), nil
	case after != "":
		b, ok := findBlock(bs, after)
		if !ok {
			return nil, fmt.Errorf("no milestone %q", after)
		}
		return insertAfter(lines, b.end, table), nil
	}
	if b, ok := findBlock(bs, LaunchMilestone); ok && name != LaunchMilestone {
		return insertAt(lines, b.start, table), nil
	}
	if len(bs) > 0 {
		return insertAfter(lines, bs[len(bs)-1].end, table), nil
	}
	for i, l := range lines {
		if isArrayTable(l, "apps") {
			return insertAt(lines, leadStart(lines, i), table), nil
		}
	}
	end := len(lines)
	for end > 0 && isBlank(lines[end-1]) {
		end--
	}
	return insertAfter(lines[:end], end, table), nil
}

// AddMilestoneText adds a [[milestones]] table. See placeTable for where it
// goes. An older [launch] target is converted first.
func AddMilestoneText(text string, m Milestone, before, after string) (string, error) {
	text, _ = ConvertLaunchText(text)
	lines := splitLines(text)
	if _, ok := findBlock(milestoneBlocks(lines), m.Name); ok {
		return "", fmt.Errorf("milestone %q already exists", m.Name)
	}
	lines, err := placeTable(lines, milestoneTable(m), m.Name, before, after)
	if err != nil {
		return "", err
	}
	return joinLines(lines), nil
}

// cut removes a table and the blank lines after it; at the end of the file
// it takes the blank lines before it instead.
func cut(lines []string, b mblock) (rest, table []string) {
	table = append([]string{}, lines[b.start:b.end]...)
	start, end := b.start, b.end
	for end < len(lines) && isBlank(lines[end]) {
		end++
	}
	if end == len(lines) {
		for start > 0 && isBlank(lines[start-1]) {
			start--
		}
	}
	rest = append(append([]string{}, lines[:start]...), lines[end:]...)
	return rest, table
}

// RemoveMilestoneText deletes the milestone called name. An older [launch]
// target is converted first.
func RemoveMilestoneText(text, name string) (string, error) {
	text, _ = ConvertLaunchText(text)
	lines := splitLines(text)
	b, ok := findBlock(milestoneBlocks(lines), name)
	if !ok {
		return "", fmt.Errorf("no milestone %q", name)
	}
	rest, _ := cut(lines, b)
	return joinLines(rest), nil
}

// MoveMilestoneText moves the milestone called name before or after another.
// Exactly one of before and after is set.
func MoveMilestoneText(text, name, before, after string) (string, error) {
	if (before == "") == (after == "") {
		return "", fmt.Errorf("give one of before and after")
	}
	if before == name || after == name {
		return "", fmt.Errorf("cannot move milestone %q relative to itself", name)
	}
	text, _ = ConvertLaunchText(text)
	lines := splitLines(text)
	bs := milestoneBlocks(lines)
	b, ok := findBlock(bs, name)
	if !ok {
		return "", fmt.Errorf("no milestone %q", name)
	}
	for _, other := range []string{before, after} {
		if _, ok := findBlock(bs, other); other != "" && !ok {
			return "", fmt.Errorf("no milestone %q", other)
		}
	}
	rest, table := cut(lines, b)
	lines, err := placeTable(rest, table, name, before, after)
	if err != nil {
		return "", err
	}
	return joinLines(lines), nil
}

// SetMilestoneTargetsText sets the targets of the milestone called name,
// adding it (see placeTable) when there is none.
func SetMilestoneTargetsText(text, name string, targets []string) (string, error) {
	lines := splitLines(text)
	b, ok := findBlock(milestoneBlocks(lines), name)
	if !ok {
		return AddMilestoneText(text, Milestone{Name: name, Targets: targets}, "", "")
	}
	value := "targets = " + quoteList(targets)
	for i := b.head + 1; i < b.end; i++ {
		if !targetsKeyRE.MatchString(lines[i]) {
			continue
		}
		// The array may span lines: find where its brackets close.
		j, depth := i, 0
		for ; j < b.end; j++ {
			depth += bracketDepth(lines[j])
			if depth <= 0 {
				break
			}
		}
		if j == b.end {
			return "", fmt.Errorf("could not read the targets of milestone %q", name)
		}
		comment := ""
		if k := commentAt(lines[j]); k >= 0 {
			comment = "  " + strings.TrimSpace(lines[j][k:])
		}
		out := append([]string{}, lines[:i]...)
		out = append(out, value+comment)
		return joinLines(append(out, lines[j+1:]...)), nil
	}
	lines = append(lines[:b.head+1], append([]string{value}, lines[b.head+1:]...)...)
	return joinLines(lines), nil
}

// commentAt returns where a comment starts in l, outside quoted strings, or -1.
func commentAt(l string) int {
	var quote rune
	for i, r := range l {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
		case r == '#':
			return i
		}
	}
	return -1
}

// bracketDepth counts [ less ] outside quoted strings and comments.
func bracketDepth(l string) int {
	if k := commentAt(l); k >= 0 {
		l = l[:k]
	}
	n := 0
	var quote rune
	for _, r := range l {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
		case r == '[':
			n++
		case r == ']':
			n--
		}
	}
	return n
}
