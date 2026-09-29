package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/AliyuYahaya/Ajiya/internal/activity"
	"github.com/AliyuYahaya/Ajiya/internal/build"
	"github.com/AliyuYahaya/Ajiya/internal/check"
	"github.com/AliyuYahaya/Ajiya/internal/gitx"
)

// clock returns the current time; tests set e.now to make durations exact.
func (e *env) clock() time.Time {
	if e.now != nil {
		return e.now()
	}
	return time.Now()
}

// Sources of a start time.
const (
	startLocal   = "local"   // the start record kept in this clone
	startHistory = "history" // the commit that made the ticket in progress
)

const (
	statusFirst    = 3 // tickets listed from 'next'
	statusActivity = 5 // activity entries listed
)

type statusMilestone struct {
	Name     string `json:"name"`
	Closed   int    `json:"closed"`
	Required int    `json:"required"`
	CanStart int    `json:"can_start"`
}

type statusTicket struct {
	ID    string `json:"id"`
	App   string `json:"app"`
	Title string `json:"title"`
}

type statusActive struct {
	statusTicket
	// SinceSeconds is how long the ticket has been in progress; nil when unknown.
	SinceSeconds *int64 `json:"since_seconds"`
	SinceSource  string `json:"since_source"` // local, history, or "" when unknown
}

type statusHuman struct {
	statusTicket
	Reason string `json:"reason"`
}

type statusActivityEntry struct {
	Short   string   `json:"short"`
	Date    string   `json:"date"`
	Subject string   `json:"subject"`
	Refs    []string `json:"refs"`
}

// statusData is what 'ajiya status' shows, and its --json output.
type statusData struct {
	Project    string            `json:"project"`
	Milestones []statusMilestone `json:"milestones"`
	InProgress []statusActive    `json:"in_progress"`
	CanStart   struct {
		Count int            `json:"count"`
		First []statusTicket `json:"first"`
	} `json:"can_start"`
	NeedsHuman []statusHuman `json:"needs_human"`
	Checks     struct {
		Errors   int `json:"errors"`
		Warnings int `json:"warnings"`
	} `json:"checks"`
	Activity []statusActivityEntry `json:"activity"`
}

func runStatus(e *env, args []string) error {
	fs := newFlags("status")
	brief := fs.Bool("brief", false, "print at most three lines")
	asJSON := fs.Bool("json", false, "print JSON")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	if *brief && *asJSON {
		return usageErr("give --brief or --json, not both")
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	d, err := build.Collect(pr.cfg, pr.plan, nil)
	if err != nil {
		return err
	}
	// The same findings as 'ajiya check', judged at the same time.
	if d.Checks, err = allFindings(pr, e.clock()); err != nil {
		return err
	}
	st := newStatus(pr, d, e.clock())
	switch {
	case *asJSON:
		return writeJSON(e, st)
	case *brief:
		fmt.Fprint(e.stdout, st.brief())
	default:
		fmt.Fprint(e.stdout, st.text())
	}
	return nil
}

// newStatus condenses the project data that 'ajiya build' collects: the
// milestone counts, the ready list in 'next' order, the checks and the activity.
func newStatus(pr *project, d *build.Data, now time.Time) *statusData {
	st := &statusData{Project: d.Project.Name}
	st.Milestones, st.InProgress, st.NeedsHuman, st.Activity = []statusMilestone{}, []statusActive{}, []statusHuman{}, []statusActivityEntry{}
	st.CanStart.First = []statusTicket{}

	for _, m := range d.Milestones {
		st.Milestones = append(st.Milestones, statusMilestone{m.Name, m.Required.Closed(), m.Required.Total, m.Ready})
	}
	byID := map[string]build.Ticket{}
	for _, t := range d.Tickets {
		byID[t.ID] = t
	}
	brief := func(t build.Ticket) statusTicket { return statusTicket{t.ID, t.App, t.Title} }
	for _, t := range d.Tickets { // by ID
		if t.Status.State == "in_progress" {
			a := statusActive{statusTicket: brief(t)}
			if at, src := startedAt(pr.cfg.Root, t.ID, d.Activity); src != "" {
				secs := max(now.Unix()-at, 0)
				a.SinceSeconds, a.SinceSource = &secs, src
			}
			st.InProgress = append(st.InProgress, a)
		}
		if t.Status.Human != "" && t.Status.State != "done" && t.Status.State != "dropped" {
			st.NeedsHuman = append(st.NeedsHuman, statusHuman{brief(t), t.Status.Human})
		}
	}
	st.CanStart.Count = len(d.Next)
	for _, id := range d.Next[:min(len(d.Next), statusFirst)] {
		st.CanStart.First = append(st.CanStart.First, brief(byID[id]))
	}
	st.Checks.Errors, st.Checks.Warnings = check.Count(d.Checks)
	for _, a := range d.Activity[:min(len(d.Activity), statusActivity)] {
		st.Activity = append(st.Activity, statusActivityEntry{a.Short, a.Date, a.Subject, a.Refs})
	}
	return st
}

// startedAt finds when a ticket was started, as Unix seconds: from the local
// start record if there is one, else from the newest recent commit that made
// it in progress. The source is "" when neither knows.
func startedAt(root, id string, recent []activity.Entry) (int64, string) {
	if at, ok := gitx.StartedAt(root, id); ok {
		return at, startLocal
	}
	for _, en := range recent { // newest first
		for _, c := range en.Changes {
			if c.ID == id && c.To == "in_progress" {
				return en.Time, startHistory
			}
		}
	}
	return 0, ""
}

// ago says how long ago in one short unit: 5m, 3h or 2d.
func ago(secs int64) string {
	switch {
	case secs < 60:
		return "just now"
	case secs < 60*60:
		return fmt.Sprintf("%dm ago", secs/60)
	case secs < 48*60*60:
		return fmt.Sprintf("%dh ago", secs/(60*60))
	}
	return fmt.Sprintf("%dd ago", secs/(24*60*60))
}

func plural(n int, one string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %ss", n, one)
}

func (st *statusData) checksText() string {
	if st.Checks.Errors == 0 && st.Checks.Warnings == 0 {
		return "no problems"
	}
	return plural(st.Checks.Errors, "error") + ", " + plural(st.Checks.Warnings, "warning")
}

func (a statusActive) since() string {
	if a.SinceSeconds == nil {
		return ""
	}
	return ago(*a.SinceSeconds)
}

func (st *statusData) text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Project: %s\n\nMilestones\n", st.Project)
	if len(st.Milestones) == 0 {
		b.WriteString("  none\n")
	}
	for i, m := range st.Milestones {
		fmt.Fprintf(&b, "  %d. %s: %d of %d done, %d can start now\n", i+1, m.Name, m.Closed, m.Required, m.CanStart)
	}
	b.WriteString("\nIn progress\n")
	if len(st.InProgress) == 0 {
		b.WriteString("  none\n")
	}
	for _, a := range st.InProgress {
		fmt.Fprintf(&b, "  %s  %s  %s", a.ID, a.App, a.Title)
		if s := a.since(); s != "" {
			fmt.Fprintf(&b, " (started %s)", s)
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "\nCan start now: %d\n", st.CanStart.Count)
	for _, t := range st.CanStart.First {
		fmt.Fprintf(&b, "  %s  %s  %s\n", t.ID, t.App, t.Title)
	}
	if len(st.NeedsHuman) > 0 {
		b.WriteString("\nNeeds a human\n")
		for _, h := range st.NeedsHuman {
			fmt.Fprintf(&b, "  %s  %s: %s\n", h.ID, h.Title, h.Reason)
		}
	}
	fmt.Fprintf(&b, "\nChecks: %s\n\nRecent activity\n", st.checksText())
	if len(st.Activity) == 0 {
		b.WriteString("  none\n")
	}
	for _, a := range st.Activity {
		fmt.Fprintf(&b, "  %s  %s", a.Date, a.Subject)
		if len(a.Refs) > 0 {
			fmt.Fprintf(&b, " [%s]", strings.Join(a.Refs, ", "))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// brief is three lines: where the milestones stand, what is being worked on,
// and what to do next with the check result.
func (st *statusData) brief() string {
	var b strings.Builder
	b.WriteString(st.Project)
	for _, m := range st.Milestones {
		if m.Closed < m.Required { // the first milestone still open
			fmt.Fprintf(&b, " · %s %d of %d done", m.Name, m.Closed, m.Required)
			break
		}
	}
	fmt.Fprintf(&b, " · %d can start now\n", st.CanStart.Count)

	b.WriteString("In progress: ")
	if len(st.InProgress) == 0 {
		b.WriteString("none")
	}
	for i, a := range st.InProgress {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(a.ID)
		if s := a.since(); s != "" {
			fmt.Fprintf(&b, " (%s)", strings.Replace(s, " ago", "", 1))
		}
	}
	if len(st.NeedsHuman) > 0 {
		var ids []string
		for _, h := range st.NeedsHuman {
			ids = append(ids, h.ID)
		}
		fmt.Fprintf(&b, " · Needs a human: %s", strings.Join(ids, ", "))
	}
	b.WriteString("\n")

	b.WriteString("Next: ")
	if len(st.CanStart.First) == 0 {
		b.WriteString("nothing")
	}
	for i, t := range st.CanStart.First {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(t.ID)
	}
	fmt.Fprintf(&b, " · Checks: %s\n", st.checksText())
	return b.String()
}
