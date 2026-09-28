// Package activity reads the recent history of a project from git: each
// commit, the tickets it names, the ticket statuses it changed in ajiya/, and
// whether an agent co-authored it.
//
// The result depends only on the repository, not on the clock: the window is
// measured back from the newest commit on HEAD.
package activity

import (
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/gitx"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// Days is the default length of the window.
const Days = 30

// Entry is one commit in the feed.
type Entry struct {
	Hash    string   `json:"hash"`
	Short   string   `json:"short"`
	Date    string   `json:"date"` // committer date, YYYY-MM-DD
	Author  string   `json:"author"`
	Subject string   `json:"subject"`
	Refs    []string `json:"refs"` // ticket IDs named in Ajiya trailers
	Chore   bool     `json:"chore"`
	Merge   bool     `json:"merge"`
	Revert  bool     `json:"revert"`
	Agent   string   `json:"agent"` // agent named in a Co-Authored-By trailer, such as "Claude", or ""
	Changes []Change `json:"changes"`
	Time    int64    `json:"time"` // committer time, Unix seconds
}

// Change is a ticket whose Status cell a commit changed or added.
type Change struct {
	ID    string `json:"id"`
	Phase string `json:"phase"` // slug of the phase holding the ticket after the commit
	From  string `json:"from"`  // pending, in_progress, done or dropped; "" for a new ticket
	To    string `json:"to"`
	Text  string `json:"text"` // the new Status cell
}

// Agents are the coding agents recognised in Co-Authored-By trailers, by
// canonical name.
var Agents = []string{"Claude", "Codex", "Copilot", "Cursor", "Gemini", "Devin", "Aider", "Jules", "Amp"}

var agentRE = regexp.MustCompile(`(?i)\b(` + strings.Join(Agents, "|") + `)\b`)

var stateNames = map[plan.State]string{plan.Pending: "pending", plan.InProgress: "in_progress", plan.Done: "done", plan.Dropped: "dropped"}

// Recent returns the commits on HEAD from the given number of days before the
// newest one up to it, newest first (ties by hash). A repository without
// commits has no activity.
func Recent(root string, days int) ([]Entry, error) {
	if days <= 0 {
		days = Days
	}
	head, ok, err := gitx.HeadTime(root)
	if err != nil || !ok {
		return nil, err
	}
	log, err := gitx.LogSince(root, head-int64(days)*24*60*60)
	if err != nil {
		return nil, err
	}
	entries := make([]Entry, 0, len(log))
	for _, c := range log {
		e := Entry{
			Hash: c.Hash, Short: c.Short(), Date: c.Date, Author: c.Author, Subject: c.Subject,
			Refs: c.IDs, Chore: c.Chore, Merge: c.Merge, Revert: c.Revert,
			Agent: Agent(c.CoAuthors), Time: c.Time,
		}
		if e.Refs == nil {
			e.Refs = []string{}
		}
		// A merge lists no changes: its tickets changed in the commits it merged.
		if !c.Merge {
			if e.Changes, err = changes(root, c.Hash, phaseFiles(c.Files)); err != nil {
				return nil, err
			}
		}
		if e.Changes == nil {
			e.Changes = []Change{}
		}
		entries = append(entries, e)
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Time != entries[j].Time {
			return entries[i].Time > entries[j].Time
		}
		return entries[i].Hash < entries[j].Hash
	})
	return entries, nil
}

// Agent returns the canonical name of the first agent named in Co-Authored-By
// values, or "" if they name only people.
func Agent(coAuthors []string) string {
	for _, v := range coAuthors {
		name, _, _ := strings.Cut(v, "<") // match the name, not the email address
		if m := agentRE.FindString(name); m != "" {
			for _, a := range Agents {
				if strings.EqualFold(a, m) {
					return a
				}
			}
		}
	}
	return ""
}

// phaseFiles returns the phase files among a commit's paths.
func phaseFiles(files []string) []string {
	var out []string
	for _, f := range files {
		if dir, name := path.Split(f); dir == plan.Dir+"/" && strings.HasSuffix(name, ".md") && plan.SlugRE.MatchString(strings.TrimSuffix(name, ".md")) {
			out = append(out, f)
		}
	}
	return out
}

// changes compares the phase files a commit touched with its first parent's.
// Tickets are matched by ID across files, so a ticket moved between phases
// counts only if its status changed. A file that does not parse on either
// side is skipped: history may hold broken files, and the feed still builds.
func changes(root, rev string, files []string) ([]Change, error) {
	if len(files) == 0 {
		return nil, nil
	}
	before := map[string]plan.Status{}
	var after []*plan.Ticket
	for _, f := range files {
		slug := strings.TrimSuffix(path.Base(f), ".md")
		old, err := readPhase(root, rev+"^:"+f, slug)
		if err != nil {
			return nil, err
		}
		cur, err := readPhase(root, rev+":"+f, slug)
		if err != nil {
			return nil, err
		}
		if old != nil {
			for _, t := range old.Tickets {
				before[t.ID] = t.Status
			}
		}
		if cur != nil {
			after = append(after, cur.Tickets...)
		}
	}
	var out []Change
	for _, t := range after {
		text := t.Status.String()
		c := Change{ID: t.ID, Phase: t.Phase, To: stateNames[t.Status.State], Text: text}
		if old, ok := before[t.ID]; ok {
			if old.String() == text {
				continue
			}
			c.From = stateNames[old.State]
		}
		out = append(out, c)
	}
	sort.SliceStable(out, func(i, j int) bool { return plan.LessID(out[i].ID, out[j].ID) })
	return out, nil
}

// readPhase reads and parses a phase file at a revision; nil if it is missing
// there (as in a root commit's parent) or does not parse.
func readPhase(root, spec, slug string) (*plan.Phase, error) {
	data, ok, err := gitx.Show(root, spec)
	if err != nil || !ok {
		return nil, err
	}
	ph, err := plan.ParsePhase(slug, data)
	if err != nil {
		return nil, nil
	}
	return ph, nil
}
