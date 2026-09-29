// Package kit installs the agent kit that 'ajiya init' writes into a project:
// the two guides, the Claude Code skill, and a marked block in AGENTS.md and
// CLAUDE.md.
//
// The guides and the skill belong to Ajiya and are rewritten whenever they
// differ from this version. AGENTS.md and CLAUDE.md belong to the user: only the
// text between the markers is Ajiya's, and nothing outside it is touched.
package kit

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed files
var files embed.FS

// Markers around Ajiya's block in AGENTS.md and CLAUDE.md.
const (
	Start = "<!-- ajiya:start -->"
	End   = "<!-- ajiya:end -->"
)

// Owned maps each file Ajiya owns, relative to the project, to its source (see
// Source).
var Owned = []struct{ Path, Source string }{
	{".ajiya/guide/setup.md", "guide/setup.md"},
	{".ajiya/guide/daily.md", "guide/daily.md"},
	{".claude/skills/ajiya/SKILL.md", "SKILL.md"},
}

// Blocks are the user's files that carry the marked block.
var Blocks = []string{"AGENTS.md", "CLAUDE.md"}

// Result says what Install did to one file.
type Result struct {
	Path   string
	Action string // "created", "updated" or "unchanged"
}

// Source returns the kit's copy of a file in files/, such as "guide/daily.md",
// "SKILL.md" or "block.md".
func Source(name string) string {
	b, err := files.ReadFile("files/" + name)
	if err != nil {
		panic(err) // embedded at build time
	}
	return string(b)
}

// Block returns the marked block, from Start to End, with no trailing newline.
func Block() string { return strings.TrimRight(Source("block.md"), "\n") }

// Install writes the kit into dir. A file is written only when its content
// changes. A user file whose markers are broken is refused before anything is
// written, so a refusal leaves the project as it was.
func Install(dir string) ([]Result, error) {
	type write struct {
		path, text, action string
	}
	var plan []write
	for _, o := range Owned {
		text := Source(o.Source)
		action, err := compare(filepath.Join(dir, o.Path), text)
		if err != nil {
			return nil, err
		}
		plan = append(plan, write{o.Path, text, action})
	}
	for _, name := range Blocks {
		old, err := os.ReadFile(filepath.Join(dir, name))
		exists := err == nil
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		text, err := WithBlock(string(old))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		action := "created"
		if exists {
			action = "updated"
			if text == string(old) {
				action = "unchanged"
			}
		}
		plan = append(plan, write{name, text, action})
	}
	out := make([]Result, 0, len(plan))
	for _, w := range plan {
		if w.action != "unchanged" {
			p := filepath.Join(dir, w.path)
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				return nil, err
			}
			if err := os.WriteFile(p, []byte(w.text), 0o644); err != nil {
				return nil, err
			}
		}
		out = append(out, Result{w.path, w.action})
	}
	return out, nil
}

// compare says what writing text to path would do.
func compare(path, text string) (string, error) {
	old, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return "created", nil
	case err != nil:
		return "", err
	case string(old) == text:
		return "unchanged", nil
	}
	return "updated", nil
}

// WithBlock returns a user file's text with Ajiya's block in it: the text
// between the markers is replaced, or the block is appended when there are no
// markers. Everything outside the markers is kept byte for byte.
func WithBlock(text string) (string, error) {
	block := Block()
	ns, ne := strings.Count(text, Start), strings.Count(text, End)
	switch {
	case ns == 0 && ne == 0:
		if text == "" {
			return block + "\n", nil
		}
		sep := "\n\n"
		if strings.HasSuffix(text, "\n\n") {
			sep = ""
		} else if strings.HasSuffix(text, "\n") {
			sep = "\n"
		}
		return text + sep + block + "\n", nil
	case ns == 1 && ne == 1:
		i, j := strings.Index(text, Start), strings.Index(text, End)
		if j < i {
			return "", fmt.Errorf("%s comes before %s; put the markers back in order or remove both, then run ajiya init again", End, Start)
		}
		return text[:i] + block + text[j+len(End):], nil
	}
	return "", fmt.Errorf("found %d %s and %d %s markers, want one of each; remove the extra ones, then run ajiya init again", ns, Start, ne, End)
}
