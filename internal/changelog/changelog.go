// Package changelog groups commits by the phase and ticket their Ajiya
// trailers name.
package changelog

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/gitx"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// Changelog is the commits of a range grouped by phase and ticket. It is also
// the JSON shape of 'ajiya changelog --json': add fields, never rename them.
type Changelog struct {
	Phases []Phase  `json:"phases"` // in display order, only phases with commits
	Other  []Commit `json:"other"`  // chores, unknown IDs and commits with no trailer
}

// Phase is one phase and its tickets that commits name.
type Phase struct {
	Slug    string   `json:"slug"`
	Title   string   `json:"title"`
	Tickets []Ticket `json:"tickets"` // in ID order
}

// Ticket is one ticket and the commits that name it.
type Ticket struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	State   string   `json:"state"` // pending, in_progress, done, dropped
	Commits []Commit `json:"commits"`
}

// Commit is one commit, oldest first in every list.
type Commit struct {
	Hash    string `json:"hash"`
	Date    string `json:"date"`
	Subject string `json:"subject"`
}

var stateNames = map[plan.State]string{plan.Pending: "pending", plan.InProgress: "in_progress", plan.Done: "done", plan.Dropped: "dropped"}

// Build groups commits (newest first, as gitx.Log returns them) by phase and
// ticket. A commit naming several tickets is listed under each; an old ID
// counts for the ticket that has it as an alias. Merges are skipped, and a
// commit listed under no ticket goes to Other. Phases are in slug order.
func Build(p *plan.Plan, commits []gitx.Commit) *Changelog {
	return BuildIn(p, p.Phases, commits)
}

// BuildIn is Build with the phases listed in the given order, such as
// plan.DisplayOrder.
func BuildIn(p *plan.Plan, phases []*plan.Phase, commits []gitx.Commit) *Changelog {
	byTicket := map[*plan.Ticket][]Commit{}
	cl := &Changelog{Phases: []Phase{}, Other: []Commit{}}
	for _, c := range slices.Backward(commits) {
		if c.Merge {
			continue
		}
		cc := Commit{c.Hash, c.Date, c.Subject}
		var seen []*plan.Ticket
		for _, id := range c.IDs {
			if t := p.Ticket(id); t != nil && !slices.Contains(seen, t) {
				seen = append(seen, t)
				byTicket[t] = append(byTicket[t], cc)
			}
		}
		if len(seen) == 0 {
			cl.Other = append(cl.Other, cc)
		}
	}
	for _, ph := range phases {
		out := Phase{Slug: ph.Slug, Title: ph.Title, Tickets: []Ticket{}}
		for _, t := range p.Tickets() {
			if t.Phase == ph.Slug && len(byTicket[t]) > 0 {
				out.Tickets = append(out.Tickets, Ticket{t.ID, t.Title, stateNames[t.Status.State], byTicket[t]})
			}
		}
		if len(out.Tickets) > 0 {
			cl.Phases = append(cl.Phases, out)
		}
	}
	return cl
}

// WriteMarkdown writes the changelog as markdown.
func (cl *Changelog) WriteMarkdown(w io.Writer) error {
	var b strings.Builder
	b.WriteString("# Changelog\n")
	for _, ph := range cl.Phases {
		fmt.Fprintf(&b, "\n## %s\n\n", ph.Title)
		for _, t := range ph.Tickets {
			fmt.Fprintf(&b, "- **%s** %s — %s\n", t.ID, t.Title, strings.ReplaceAll(t.State, "_", " "))
			for _, c := range t.Commits {
				fmt.Fprintf(&b, "  - %s %s\n", short(c.Hash), c.Subject)
			}
		}
	}
	if len(cl.Other) > 0 {
		b.WriteString("\n## Other changes\n\n")
		for _, c := range cl.Other {
			fmt.Fprintf(&b, "- %s %s\n", short(c.Hash), c.Subject)
		}
	}
	if len(cl.Phases) == 0 && len(cl.Other) == 0 {
		b.WriteString("\nNo changes.\n")
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func short(hash string) string { return gitx.Commit{Hash: hash}.Short() }
