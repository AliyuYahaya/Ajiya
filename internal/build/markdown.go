package build

import (
	"fmt"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/check"
)

// Markdown renders PROGRESS.md.
func Markdown(d *Data) string {
	var b strings.Builder
	w := func(format string, a ...any) { fmt.Fprintf(&b, format, a...) }
	byID := map[string]Ticket{}
	for _, t := range d.Tickets {
		byID[t.ID] = t
	}
	line := func(t Ticket) string {
		return fmt.Sprintf("**%s** %s (%s)", t.ID, esc(t.Title), t.App)
	}

	w("<!-- Written by ajiya build. Do not edit: run 'ajiya build'. -->\n")
	w("# %s: progress\n\nGenerated %s.\n\n", esc(d.Project.Name), d.Generated)

	w("## Milestones\n\n")
	if len(d.Milestones) == 0 {
		w("No milestones. Set one with `ajiya launch set <phase|ticket>`.\n\n")
	} else {
		w("| Milestone | Closed | Required | %% | Can start now |\n|---|---|---|---|---|\n")
		for _, m := range d.Milestones {
			w("| %s | %d | %d | %d%% | %d |\n", m.Name, m.Required.Closed(), m.Required.Total, m.Percent, m.Ready)
		}
		w("\n")
	}

	w("## Phases\n\n")
	if len(d.Phases) == 0 {
		w("No phases yet.\n\n")
	} else {
		w("| Phase | Done | In progress | Pending | Total | %% | Milestone |\n|---|---|---|---|---|---|---|\n")
		for _, p := range d.Phases {
			ms := p.Milestone
			if ms == "" {
				ms = "-"
			}
			w("| %s | %d | %d | %d | %d | %d%% | %s |\n", esc(p.Title), p.Counts.Closed(), p.Counts.InProgress, p.Counts.Pending, p.Counts.Total, p.Percent, ms)
		}
		w("\n")
	}

	w("## Can start now\n\n")
	if len(d.Next) == 0 {
		w("Nothing.\n\n")
	} else {
		for _, id := range d.Next {
			t := byID[id]
			extra := ""
			if t.Milestone != "" {
				extra = " · " + t.Milestone
			}
			if t.Status.State == "in_progress" {
				extra += " · in progress"
			}
			w("- %s%s\n", line(t), extra)
		}
		w("\n")
	}

	w("## Waiting\n\n")
	waiting := 0
	for _, t := range d.Tickets {
		if t.Status.State == "done" || t.Status.State == "dropped" || t.Ready {
			continue
		}
		waiting++
		on := strings.Join(t.WaitingOn, ", ")
		if t.Status.Blocked != "" {
			on = join(on, "blocked: "+esc(t.Status.Blocked))
		}
		w("- %s waits on %s\n", line(t), on)
	}
	if waiting == 0 {
		w("Nothing.\n")
	}
	w("\n")

	w("## Needs a human\n\n")
	humans := 0
	for _, t := range d.Tickets {
		if t.Status.Human != "" && t.Status.State != "done" && t.Status.State != "dropped" {
			humans++
			w("- %s: %s\n", line(t), esc(t.Status.Human))
		}
	}
	if humans == 0 {
		w("Nobody.\n")
	}
	w("\n")

	w("## Recent activity\n\n")
	if len(d.Activity) == 0 {
		w("No commits yet.\n\n")
	} else {
		for i, e := range d.Activity {
			if i == 15 {
				w("- and %d more in data.js\n", len(d.Activity)-15)
				break
			}
			refs := strings.Join(e.Refs, ", ")
			if e.Chore {
				refs = join(refs, "chore")
			}
			if refs == "" {
				refs = "-"
			}
			agent := ""
			if e.Agent != "" {
				agent = " · with " + e.Agent
			}
			w("- %s `%s` %s (%s)%s\n", e.Date, e.Short, esc(e.Subject), refs, agent)
			for _, c := range e.Changes {
				from := c.From
				if from == "" {
					from = "new"
				}
				w("  - %s: %s → %s\n", c.ID, from, c.To)
			}
		}
		w("\n")
	}

	w("## Checks\n\n")
	errs, warns := check.Count(d.Checks)
	if len(d.Checks) == 0 {
		w("No problems found.\n")
	} else {
		w("%d error(s), %d warning(s).\n\n", errs, warns)
		for _, f := range d.Checks {
			w("- %s\n", esc(f.String()))
		}
	}
	return b.String()
}

func join(a, b string) string {
	if a == "" {
		return b
	}
	return a + "; " + b
}

// esc keeps text from breaking markdown tables.
func esc(s string) string { return strings.ReplaceAll(s, "|", `\|`) }
