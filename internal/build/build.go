// Package build writes ajiya/PROGRESS.md and ajiya/data.js from the plan.
//
// Output is deterministic: lists are sorted, and the generated date is the
// date of the newest change to the plan's sources (the newest commit that is
// not only build output, or today when ajiya/ or ajiya.toml has uncommitted
// changes). Files are written only when their content changes.
package build

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/AliyuYahaya/Ajiya/internal/activity"
	"github.com/AliyuYahaya/Ajiya/internal/check"
	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/dashboard"
	"github.com/AliyuYahaya/Ajiya/internal/gitx"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
	"github.com/AliyuYahaya/Ajiya/internal/view"
)

// Schema is the version of the data.js layout. Raise it when a field changes
// meaning or goes away; adding fields does not need a new version.
const Schema = 1

// Data is everything the dashboard and PROGRESS.md show.
type Data struct {
	Schema     int              `json:"schema"`
	Generated  string           `json:"generated"` // YYYY-MM-DD
	Project    Project          `json:"project"`
	Milestones []Milestone      `json:"milestones"`
	Phases     []Phase          `json:"phases"` // in display order
	Apps       []App            `json:"apps"`
	Tickets    []Ticket         `json:"tickets"` // by ID
	Next       []string         `json:"next"`    // IDs that can start now, in 'ajiya next' order
	Checks     []check.Finding  `json:"checks"`
	Activity   []activity.Entry `json:"activity"`
}

type Project struct {
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
}

// Counts are tickets by state. Total includes dropped tickets.
type Counts struct {
	Total      int `json:"total"`
	Done       int `json:"done"`
	InProgress int `json:"in_progress"`
	Pending    int `json:"pending"`
	Dropped    int `json:"dropped"`
}

// Closed is done plus dropped.
func (c Counts) Closed() int { return c.Done + c.Dropped }

func (c *Counts) add(t *plan.Ticket) {
	c.Total++
	switch t.Status.State {
	case plan.Done:
		c.Done++
	case plan.InProgress:
		c.InProgress++
	case plan.Dropped:
		c.Dropped++
	default:
		c.Pending++
	}
}

type Milestone struct {
	Name     string   `json:"name"`
	Targets  []string `json:"targets"`
	Required Counts   `json:"required"`
	Percent  int      `json:"percent"` // closed of required
	Ready    int      `json:"ready"`   // required tickets that can start now
}

type Phase struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Goal      string `json:"goal"`
	Counts    Counts `json:"counts"`
	Percent   int    `json:"percent"`
	Milestone string `json:"milestone"` // earliest milestone of its tickets, or ""
}

type App struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Kind   string `json:"kind,omitempty"`
	Counts Counts `json:"counts"`
}

// Ticket is a ticket with what the plan says about it.
type Ticket struct {
	view.Ticket
	Milestone  string   `json:"milestone"`  // earliest milestone, or "" if unscheduled
	Ready      bool     `json:"ready"`      // open, not blocked, dependencies closed
	WaitingOn  []string `json:"waiting_on"` // open dependencies
	Dependants []string `json:"dependants"` // tickets that depend on it directly
	Unblocks   int      `json:"unblocks"`   // open tickets waiting on it, directly or not
}

func percent(n, total int) int {
	if total == 0 {
		return 100
	}
	return n * 100 / total
}

// Collect gathers the data for a project. order is the phase display order
// (nil for slug order).
func Collect(cfg *config.Config, p *plan.Plan, order []*plan.Phase) (*Data, error) {
	root := cfg.Root
	d := &Data{
		Schema:     Schema,
		Project:    Project{Name: cfg.Project.Name, Prefix: cfg.Project.Prefix},
		Milestones: []Milestone{}, Phases: []Phase{}, Apps: []App{}, Tickets: []Ticket{},
		Next: []string{}, Checks: []check.Finding{}, Activity: []activity.Entry{},
	}

	var err error
	if d.Activity, err = activity.Recent(root, activity.Days); err != nil && !errors.Is(err, gitx.ErrNotRepo) {
		return nil, err
	}
	if d.Activity == nil {
		d.Activity = []activity.Entry{}
	}
	now, err := generatedAt(root, d.Activity)
	if err != nil {
		return nil, err
	}
	d.Generated = now.Format("2006-01-02")

	var ms []plan.Milestone
	for _, m := range cfg.MilestoneList() {
		ms = append(ms, plan.Milestone{Name: m.Name, Targets: m.Targets})
	}
	sched, schedErr := p.Schedule(ms) // a bad target is reported by the checks

	tickets := p.Tickets()
	g := plan.NewGraph(tickets)
	byID := map[string]*Ticket{}
	for _, t := range tickets {
		bt := Ticket{Ticket: view.TicketOf(t), WaitingOn: []string{}, Dependants: g.Dependants(t.ID)}
		if bt.Dependants == nil {
			bt.Dependants = []string{}
		}
		if schedErr == nil {
			bt.Milestone = sched.Of[t.ID]
		}
		for _, id := range t.Depends {
			if dep := p.Ticket(id); dep != nil && !dep.Status.Closed() {
				bt.WaitingOn = append(bt.WaitingOn, id)
			}
		}
		if !t.Status.Closed() {
			for _, id := range g.Downstream(t.ID) {
				if dt := p.Ticket(id); dt != nil && !dt.Status.Closed() {
					bt.Unblocks++
				}
			}
		}
		bt.Ready = !t.Status.Closed() && t.Status.Blocked == "" && len(bt.WaitingOn) == 0
		d.Tickets = append(d.Tickets, bt)
		byID[t.ID] = &d.Tickets[len(d.Tickets)-1]
	}
	d.Next = nextOrder(d.Tickets, sched, schedErr)

	if schedErr == nil {
		for i, m := range ms {
			bm := Milestone{Name: m.Name, Targets: m.Targets}
			for _, t := range tickets {
				if sched.Required[i][t.ID] {
					bm.Required.add(t)
					if byID[t.ID].Ready {
						bm.Ready++
					}
				}
			}
			bm.Percent = percent(bm.Required.Closed(), bm.Required.Total)
			d.Milestones = append(d.Milestones, bm)
		}
	}

	if order == nil {
		order = p.Phases
	}
	for _, ph := range order {
		bp := Phase{Slug: ph.Slug, Title: ph.Title, Goal: ph.Goal}
		best := -1
		for _, t := range ph.Tickets {
			bp.Counts.add(t)
			if schedErr == nil {
				if i := sched.Index(sched.Of[t.ID]); i >= 0 && (best < 0 || i < best) {
					best = i
				}
			}
		}
		if best >= 0 {
			bp.Milestone = ms[best].Name
		}
		bp.Percent = percent(bp.Counts.Closed(), bp.Counts.Total)
		d.Phases = append(d.Phases, bp)
	}

	apps := append([]config.App{{Name: config.InfraApp}}, cfg.Apps...)
	for _, a := range apps {
		ba := App{Name: a.Name, Path: a.Path, Kind: a.Kind}
		for _, t := range tickets {
			if t.App == a.Name {
				ba.Counts.add(t)
			}
		}
		if a.Name != config.InfraApp || ba.Counts.Total > 0 {
			d.Apps = append(d.Apps, ba)
		}
	}

	d.Checks = check.Run(cfg, p)
	history, err := check.History(cfg, p, now)
	if err != nil {
		return nil, err
	}
	d.Checks = append(d.Checks, history...)
	check.Sort(d.Checks)
	if d.Checks == nil {
		d.Checks = []check.Finding{}
	}
	return d, nil
}

// nextOrder sorts ready tickets as 'ajiya next' does: earliest milestone first
// (unscheduled last), then by how many open tickets each unblocks, then ID.
func nextOrder(ts []Ticket, sched *plan.Schedule, schedErr error) []string {
	rank := func(t Ticket) int {
		if schedErr != nil || t.Milestone == "" {
			return 1 << 30
		}
		return sched.Index(t.Milestone)
	}
	var ready []Ticket
	for _, t := range ts {
		if t.Ready {
			ready = append(ready, t)
		}
	}
	sort.SliceStable(ready, func(i, j int) bool {
		a, b := ready[i], ready[j]
		if rank(a) != rank(b) {
			return rank(a) < rank(b)
		}
		if a.Unblocks != b.Unblocks {
			return a.Unblocks > b.Unblocks
		}
		return plan.LessID(a.ID, b.ID)
	})
	ids := []string{}
	for _, t := range ready {
		ids = append(ids, t.ID)
	}
	return ids
}

// generatedAt is when the sources last changed: now if ajiya/ or ajiya.toml
// has uncommitted changes other than the build output, else the newest
// activity entry, else now.
func generatedAt(root string, entries []activity.Entry) (time.Time, error) {
	paths := []string{plan.Dir, config.FileName}
	for _, g := range activity.Generated {
		paths = append(paths, ":(exclude)"+g)
	}
	dirty, err := gitx.Dirty(root, paths...)
	if err != nil && !errors.Is(err, gitx.ErrNotRepo) {
		return time.Time{}, err
	}
	if dirty || len(entries) == 0 {
		return time.Now(), nil
	}
	return time.Unix(entries[0].Time, 0), nil
}

// Files returns the contents of PROGRESS.md and data.js.
func Files(d *Data) (progress, dataJS []byte, err error) {
	js, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	var b bytes.Buffer
	b.WriteString("// Written by ajiya build. Do not edit: run 'ajiya build'.\nwindow.AJIYA = ")
	b.Write(js)
	b.WriteString(";\n")
	return []byte(Markdown(d)), b.Bytes(), nil
}

// Write writes PROGRESS.md, data.js and the dashboard page into root/ajiya when their content changed, and
// returns the paths it wrote.
func Write(root string, d *Data) ([]string, error) {
	progress, dataJS, err := Files(d)
	if err != nil {
		return nil, err
	}
	var wrote []string
	for _, f := range []struct {
		path string
		data []byte
	}{{activity.Generated[0], progress}, {activity.Generated[1], dataJS}, {activity.Generated[2], dashboard.HTML()}} {
		full := filepath.Join(root, filepath.FromSlash(f.path))
		if old, err := os.ReadFile(full); err == nil && bytes.Equal(old, f.data) {
			continue
		}
		if err := os.WriteFile(full, f.data, 0o644); err != nil {
			return wrote, err
		}
		wrote = append(wrote, f.path)
	}
	return wrote, nil
}
