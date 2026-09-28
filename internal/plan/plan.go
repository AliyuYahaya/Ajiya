// Package plan reads and writes the phase files in ajiya/.
package plan

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Dir is the folder holding the phase files, relative to the project root.
const Dir = "ajiya"

// Plan is every phase file in ajiya/.
type Plan struct {
	Root   string // project root
	Phases []*Phase
	// Raw holds each phase file as read, with CRLF turned into LF, keyed by slug.
	Raw map[string][]byte
	// Broken lists files that could not be parsed, keyed by slug.
	Broken map[string]*ParseError
}

// Load reads every <slug>.md in root/ajiya. Files that do not parse are
// listed in Broken rather than failing the load.
func Load(root string) (*Plan, error) {
	p := &Plan{Root: root, Raw: map[string][]byte{}, Broken: map[string]*ParseError{}}
	entries, err := os.ReadDir(filepath.Join(root, Dir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return p, nil
		}
		return nil, err
	}
	for _, e := range entries {
		slug, ok := strings.CutSuffix(e.Name(), ".md")
		if e.IsDir() || !ok || !SlugRE.MatchString(slug) {
			continue // PROGRESS.md and other files are not phases
		}
		data, err := os.ReadFile(filepath.Join(root, Dir, e.Name()))
		if err != nil {
			return nil, err
		}
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		p.Raw[slug] = data
		ph, err := ParsePhase(slug, data)
		if err != nil {
			var pe *ParseError
			if !errors.As(err, &pe) {
				return nil, err
			}
			p.Broken[slug] = pe
			continue
		}
		p.Phases = append(p.Phases, ph)
	}
	sort.Slice(p.Phases, func(i, j int) bool { return p.Phases[i].Slug < p.Phases[j].Slug })
	return p, nil
}

// Path returns the file path of a phase, relative to the project root, with / separators.
func Path(slug string) string { return Dir + "/" + slug + ".md" }

// Phase returns the phase with the given slug, or nil.
func (p *Plan) Phase(slug string) *Phase {
	for _, ph := range p.Phases {
		if ph.Slug == slug {
			return ph
		}
	}
	return nil
}

// Tickets returns every ticket, sorted by ID.
func (p *Plan) Tickets() []*Ticket {
	var all []*Ticket
	for _, ph := range p.Phases {
		all = append(all, ph.Tickets...)
	}
	sort.SliceStable(all, func(i, j int) bool { return LessID(all[i].ID, all[j].ID) })
	return all
}

// Ticket returns the first ticket with the given ID, or nil.
func (p *Plan) Ticket(id string) *Ticket {
	for _, ph := range p.Phases {
		for _, t := range ph.Tickets {
			if t.ID == id {
				return t
			}
		}
	}
	return nil
}

// NextID returns the next free ID for prefix: one past the highest in use.
// IDs are never reused because tickets are never deleted.
func (p *Plan) NextID(prefix string) string {
	high := 0
	for _, t := range p.Tickets() {
		if pfx, n, ok := SplitID(t.ID); ok && pfx == prefix && n > high {
			high = n
		}
	}
	return FormatID(prefix, high+1)
}

// Move moves a ticket to another phase. Both phases must then be saved.
func (p *Plan) Move(t *Ticket, to *Phase) {
	if from := p.Phase(t.Phase); from != nil {
		for i, x := range from.Tickets {
			if x == t {
				from.Tickets = append(from.Tickets[:i], from.Tickets[i+1:]...)
				break
			}
		}
	}
	t.Phase = to.Slug
	to.Tickets = append(to.Tickets, t)
}

// Save writes a phase file if its canonical form differs from what is on disk.
func (p *Plan) Save(ph *Phase) error {
	data := ph.Format()
	if bytes.Equal(p.Raw[ph.Slug], data) {
		return nil
	}
	if p.Phase(ph.Slug) == nil {
		p.Phases = append(p.Phases, ph)
		sort.Slice(p.Phases, func(i, j int) bool { return p.Phases[i].Slug < p.Phases[j].Slug })
	}
	if err := writeFile(filepath.Join(p.Root, Dir, ph.Slug+".md"), data); err != nil {
		return err
	}
	p.Raw[ph.Slug] = data
	return nil
}

// writeFile writes through a temporary file so a crash never leaves half a file.
func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".ajiya-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
