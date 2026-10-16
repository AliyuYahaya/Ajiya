package plan

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// State is the mark a Status cell starts with.
type State int

const (
	Pending    State = iota // 🟥
	InProgress              // 🟨
	Done                    // 🟩 Done
	Dropped                 // 🟩 Dropped
	Superseded              // 🟩 Superseded
)

const (
	markPending = "🟥"
	markPartial = "🟨"
	markDone    = "🟩"
	sep         = " · "
)

// Status is the parsed Status cell of a ticket row.
//
//	🟥 Pending[ · Needs a human: <Human>][ · Blocked: <Blocked>]
//	🟨 In progress[: <Note>][ · Needs a human: <Human>]
//	🟩 Done[ · <Commit>][ · issue #<Issue>][ · Done before Ajiya][ · by <By>][ · <Date>][ · tests passed][ · Note: <Note>]
//	🟩 Dropped: <Reason>, decided by <By>
//	🟩 Superseded[ by <ID>] · <Change> · was: <the previous status, whole>
//
// Any of them (but Superseded, which keeps them inside its was: part) may end with [ · Link: <Link>][ · Was: <Aliases>], which every
// status change keeps: the URL of an imported issue, and a ticket's old IDs.
type Status struct {
	State   State
	Note    string // in progress: what is left; done: a note
	Human   string // Needs a human: why
	Blocked string // pending: why
	Commit  string // done: short commit hash
	Issue   string // done: the number of the closed issue it was imported from
	Before  bool   // done: finished before the project used Ajiya
	By      string // done by a person, or who dropped it
	Date    string // done: YYYY-MM-DD
	Tests   bool   // done: the test command passed
	Reason  string // dropped: why

	// Superseded: the ticket that replaces it ("" when invalidated with no
	// replacement), the change that did it, and the status it had before.
	SupersededBy string
	Change       string
	Prev         *Status

	Link    string   // the ticket's source, such as an issue URL
	Aliases []string // old IDs the ticket had before it was imported
}

// Closed reports whether the ticket counts as finished: done, dropped or
// superseded. Dependants treat all three as closed.
func (s Status) Closed() bool {
	return s.State == Done || s.State == Dropped || s.State == Superseded
}

// Excluded reports whether the ticket stays out of progress totals and
// milestone required sets: dropped and superseded work is no longer wanted.
func (s Status) Excluded() bool { return s.State == Dropped || s.State == Superseded }

// Supersede returns the status of a ticket replaced by ticket by (empty when
// invalidated with no replacement) in change. It keeps the whole previous
// status, evidence included. A superseded ticket cannot be superseded again.
func (s Status) Supersede(by, change string) (Status, error) {
	if s.State == Superseded {
		return s, errors.New("the ticket is already superseded")
	}
	if change == "" {
		return s, errors.New("a superseded status needs the change that did it")
	}
	if err := CheckStatusText("replacement", by); err != nil {
		return s, err
	}
	if err := CheckStatusText("change", change); err != nil {
		return s, err
	}
	prev := s
	return Status{State: Superseded, SupersededBy: by, Change: change, Prev: &prev, Link: s.Link, Aliases: s.Aliases}, nil
}

// HasEvidence reports whether a done ticket says how it was done.
func (s Status) HasEvidence() bool {
	return s.Commit != "" || s.By != "" || s.Issue != "" || s.Before
}

// Carry returns a new status in state st that keeps what outlives a status
// change: Needs a human (while the ticket is open), the link and the aliases.
func (s Status) Carry(st State) Status {
	n := Status{State: st, Link: s.Link, Aliases: s.Aliases}
	if st == Pending || st == InProgress {
		n.Human = s.Human
	}
	return n
}

// Mark returns the coloured square the cell starts with.
func (s Status) Mark() string {
	switch s.State {
	case InProgress:
		return markPartial
	case Done, Dropped, Superseded:
		return markDone
	}
	return markPending
}

var (
	hashRE  = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
	dateRE  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	issueRE = regexp.MustCompile(`^issue #(\d+)$`)
)

const (
	humanTag   = "Needs a human: "
	blockedTag = "Blocked: "
	noteTag    = "Note: "
	byTag      = "by "
	decidedTag = ", decided by "
	testsTag   = "tests passed"
	beforeTag  = "Done before Ajiya"
	linkTag    = "Link: "
	wasTag     = "Was: "
	supTag     = "Superseded"
	supByTag   = " by "
	prevTag    = "was: "
)

// String formats the status in canonical form.
func (s Status) String() string {
	parts := []string{}
	switch s.State {
	case Pending:
		parts = append(parts, markPending+" Pending")
		if s.Human != "" {
			parts = append(parts, humanTag+s.Human)
		}
		if s.Blocked != "" {
			parts = append(parts, blockedTag+s.Blocked)
		}
	case InProgress:
		head := markPartial + " In progress"
		if s.Note != "" {
			head += ": " + s.Note
		}
		parts = append(parts, head)
		if s.Human != "" {
			parts = append(parts, humanTag+s.Human)
		}
	case Done:
		parts = append(parts, markDone+" Done")
		if s.Commit != "" {
			parts = append(parts, s.Commit)
		}
		if s.Issue != "" {
			parts = append(parts, "issue #"+s.Issue)
		}
		if s.Before {
			parts = append(parts, beforeTag)
		}
		if s.By != "" {
			parts = append(parts, byTag+s.By)
		}
		if s.Date != "" {
			parts = append(parts, s.Date)
		}
		if s.Tests {
			parts = append(parts, testsTag)
		}
		if s.Note != "" {
			parts = append(parts, noteTag+s.Note)
		}
	case Dropped:
		parts = append(parts, markDone+" Dropped: "+s.Reason+decidedTag+s.By)
	case Superseded:
		head := markDone + " " + supTag
		if s.SupersededBy != "" {
			head += supByTag + s.SupersededBy
		}
		prev := ""
		if s.Prev != nil {
			// Written without its coloured square: the whole cell is green.
			_, prev, _ = strings.Cut(s.Prev.String(), " ")
		}
		// The previous status carries the link and the aliases.
		return strings.Join([]string{head, s.Change, prevTag + prev}, sep)
	}
	if s.Link != "" {
		parts = append(parts, linkTag+s.Link)
	}
	if len(s.Aliases) > 0 {
		parts = append(parts, wasTag+strings.Join(s.Aliases, ", "))
	}
	return strings.Join(parts, sep)
}

// ParseStatus parses a Status cell. It accepts the parts in any order; the
// round-trip check reports cells that are not in canonical order.
func ParseStatus(cell string) (Status, error) {
	var s Status
	parts := strings.Split(cell, sep)
	head, parts := parts[0], parts[1:]
	switch {
	case head == markDone+" "+supTag || strings.HasPrefix(head, markDone+" "+supTag+supByTag):
		return parseSuperseded(head, parts)
	case strings.HasPrefix(head, markDone+" Dropped: "):
		rest := strings.TrimPrefix(head, markDone+" Dropped: ")
		i := strings.LastIndex(rest, decidedTag)
		if i < 0 {
			return s, errors.New(`dropped status needs ", decided by <name>"`)
		}
		s.State, s.Reason, s.By = Dropped, rest[:i], rest[i+len(decidedTag):]
		if s.Reason == "" || s.By == "" {
			return s, errors.New("dropped status needs a reason and a name")
		}
	case head == markPending+" Pending":
		s.State = Pending
	case head == markPartial+" In progress":
		s.State = InProgress
	case strings.HasPrefix(head, markPartial+" In progress: "):
		s.State = InProgress
		s.Note = strings.TrimPrefix(head, markPartial+" In progress: ")
	case head == markDone+" Done":
		s.State = Done
	default:
		return s, fmt.Errorf("status %q must start with 🟥 Pending, 🟨 In progress, 🟩 Done, 🟩 Dropped or 🟩 Superseded", head)
	}
	for _, p := range parts {
		if err := s.setPart(p); err != nil {
			return s, err
		}
	}
	return s, nil
}

// parseSuperseded parses what follows "🟩 Superseded[ by <ID>]": the change,
// then "was: " and the previous status, which takes up the rest of the cell.
func parseSuperseded(head string, parts []string) (Status, error) {
	s := Status{State: Superseded}
	s.SupersededBy = strings.TrimPrefix(strings.TrimPrefix(head, markDone+" "+supTag), supByTag)
	if strings.HasPrefix(head, markDone+" "+supTag+supByTag) && s.SupersededBy == "" {
		return s, errors.New("superseded status needs an ID after \"by\"")
	}
	if len(parts) < 2 || parts[0] == "" || !strings.HasPrefix(parts[1], prevTag) {
		return s, errors.New(`superseded status needs "· <change> · was: <previous status>"`)
	}
	s.Change = parts[0]
	prev := strings.Join(append([]string{strings.TrimPrefix(parts[1], prevTag)}, parts[2:]...), sep)
	switch {
	case strings.HasPrefix(prev, "Pending"):
		prev = markPending + " " + prev
	case strings.HasPrefix(prev, "In progress"):
		prev = markPartial + " " + prev
	default:
		prev = markDone + " " + prev
	}
	ps, err := ParseStatus(prev)
	if err != nil {
		return s, fmt.Errorf("superseded status: previous status: %w", err)
	}
	if ps.State == Superseded {
		return s, errors.New("superseded status: previous status cannot be superseded")
	}
	s.Prev, s.Link, s.Aliases = &ps, ps.Link, ps.Aliases
	return s, nil
}

func (s *Status) setPart(p string) error {
	set := func(field *string, v string) error {
		if *field != "" {
			return fmt.Errorf("status part %q is repeated", p)
		}
		if v == "" {
			return fmt.Errorf("status part %q is empty", p)
		}
		*field = v
		return nil
	}
	flag := func(field *bool) error {
		if *field {
			return fmt.Errorf("status part %q is repeated", p)
		}
		*field = true
		return nil
	}
	open := s.State == Pending || s.State == InProgress
	done := s.State == Done
	switch {
	case strings.HasPrefix(p, linkTag):
		return set(&s.Link, strings.TrimPrefix(p, linkTag))
	case strings.HasPrefix(p, wasTag):
		if s.Aliases != nil {
			return fmt.Errorf("status part %q is repeated", p)
		}
		s.Aliases = ParseIDList(strings.TrimPrefix(p, wasTag))
		if len(s.Aliases) == 0 {
			return fmt.Errorf("status part %q is empty", p)
		}
		return nil
	case open && strings.HasPrefix(p, humanTag):
		return set(&s.Human, strings.TrimPrefix(p, humanTag))
	case s.State == Pending && strings.HasPrefix(p, blockedTag):
		return set(&s.Blocked, strings.TrimPrefix(p, blockedTag))
	case done && hashRE.MatchString(p):
		return set(&s.Commit, p)
	case done && issueRE.MatchString(p):
		return set(&s.Issue, issueRE.FindStringSubmatch(p)[1])
	case done && p == beforeTag:
		return flag(&s.Before)
	case done && dateRE.MatchString(p):
		return set(&s.Date, p)
	case done && strings.HasPrefix(p, byTag):
		return set(&s.By, strings.TrimPrefix(p, byTag))
	case done && strings.HasPrefix(p, noteTag):
		return set(&s.Note, strings.TrimPrefix(p, noteTag))
	case done && p == testsTag:
		return flag(&s.Tests)
	}
	return fmt.Errorf("status part %q is not understood here", p)
}

// CheckStatusText reports text that cannot be stored in a Status cell.
func CheckStatusText(what, v string) error {
	if err := CheckText(what, v); err != nil {
		return err
	}
	if strings.Contains(v, "·") {
		return fmt.Errorf("%s must not contain '·', which separates status parts", what)
	}
	return nil
}

// CheckText reports text that cannot be stored in a table cell.
func CheckText(what, v string) error {
	if strings.ContainsAny(v, "\r\n") {
		return fmt.Errorf("%s must be one line", what)
	}
	if strings.TrimSpace(v) != v {
		return fmt.Errorf("%s must not start or end with spaces", what)
	}
	return nil
}
