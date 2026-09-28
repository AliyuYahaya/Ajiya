// Package importer reads work items kept outside Ajiya so the CLI can add
// them to the plan. Importers never reword items.
package importer

import (
	"regexp"
	"strings"
)

// TodoItem is one list item of a markdown to-do list.
type TodoItem struct {
	Line   int    // 1-based line number in the file
	Text   string // the item text as written, without marker and checkbox
	Ticked bool   // the item had a ticked checkbox: [x] or [X]
}

var (
	// todoItemRE matches a list item: -, *, + or a number followed by . or ),
	// then whitespace and the rest of the line (or nothing).
	todoItemRE = regexp.MustCompile(`^[ \t]*(?:[-*+]|[0-9]{1,9}[.)])(?:[ \t]+(.*))?$`)
	// todoBoxRE matches a checkbox at the start of an item's text.
	todoBoxRE = regexp.MustCompile(`^\[([ xX])\](?:[ \t]+|$)`)
	// todoBreakRE matches a thematic break such as "* * *" or "- - -".
	todoBreakRE = regexp.MustCompile(`^[ \t]*(?:(?:\*[ \t]*){3,}|(?:-[ \t]*){3,}|(?:_[ \t]*){3,})$`)
)

// ParseTodo returns the list items of a markdown file in order. Nested items
// are items of their own. Lines inside ``` or ~~~ fenced code blocks, headings
// and paragraphs are ignored, and so are items with no text.
func ParseTodo(src string) []TodoItem {
	var items []TodoItem
	fence := "" // the opening fence while inside a code block
	for i, line := range strings.Split(src, "\n") {
		line = strings.TrimSuffix(line, "\r")
		trimmed := strings.TrimLeft(line, " \t")
		if fence != "" {
			if strings.HasPrefix(trimmed, fence) && strings.Trim(trimmed, fence[:1]+" \t") == "" {
				fence = ""
			}
			continue
		}
		if f := todoFence(trimmed); f != "" {
			fence = f
			continue
		}
		if todoBreakRE.MatchString(line) {
			continue
		}
		m := todoItemRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		text, ticked := m[1], false
		if b := todoBoxRE.FindStringSubmatch(text); b != nil {
			ticked = b[1] != " "
			text = text[len(b[0]):]
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		items = append(items, TodoItem{Line: i + 1, Text: text, Ticked: ticked})
	}
	return items
}

// todoFence returns the fence that opens a code block on this line: three or
// more backticks or tildes. It returns "" when the line opens none.
func todoFence(trimmed string) string {
	for _, c := range []string{"`", "~"} {
		n := len(trimmed) - len(strings.TrimLeft(trimmed, c))
		if n >= 3 {
			return strings.Repeat(c, n)
		}
	}
	return ""
}
