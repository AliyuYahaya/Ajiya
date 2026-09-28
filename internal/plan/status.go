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
//	🟩 Done[ · <Commit>][ · by <By>][ · <Date>][ · tests passed][ · Note: <Note>]
//	🟩 Dropped: <Reason>, decided by <By>
type Status struct {
	State   State
	Note    string // in progress: what is left; done: a note
	Human   string // Needs a human: why
	Blocked string // pending: why
	Commit  string // done: short commit hash
	By      string // done by a person, or who dropped it
	Date    string // done: YYYY-MM-DD
	Tests   bool   // done: the test command passed
	Reason  string // dropped: why
}

// Closed reports whether the ticket counts as finished: done or dropped.
func (s Status) Closed() bool { return s.State == Done || s.State == Dropped }

// Mark returns the coloured square the cell starts with.
func (s Status) Mark() string {
	switch s.State {
	case InProgress:
		return markPartial
	case Done, Dropped:
		return markDone
	}
	return markPending
}

var (
	hashRE = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
	dateRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

const (
	humanTag   = "Needs a human: "
	blockedTag = "Blocked: "
	noteTag    = "Note: "
	byTag      = "by "
	decidedTag = ", decided by "
	testsTag   = "tests passed"
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
		return markDone + " Dropped: " + s.Reason + decidedTag + s.By
	}
	return strings.Join(parts, sep)
}

// ParseStatus parses a Status cell. It accepts the parts in any order; the
// round-trip check reports cells that are not in canonical order.
func ParseStatus(cell string) (Status, error) {
	var s Status
	if rest, ok := strings.CutPrefix(cell, markDone+" Dropped: "); ok {
		i := strings.LastIndex(rest, decidedTag)
		if i < 0 {
			return s, errors.New(`dropped status needs ", decided by <name>"`)
		}
		s.State, s.Reason, s.By = Dropped, rest[:i], rest[i+len(decidedTag):]
		if s.Reason == "" || s.By == "" {
			return s, errors.New("dropped status needs a reason and a name")
		}
		return s, nil
	}
	parts := strings.Split(cell, sep)
	head, parts := parts[0], parts[1:]
	switch {
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
		return s, fmt.Errorf("status %q must start with 🟥 Pending, 🟨 In progress, 🟩 Done or 🟩 Dropped", head)
	}
	for _, p := range parts {
		if err := s.setPart(p); err != nil {
			return s, err
		}
	}
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
	pendingOrActive := s.State == Pending || s.State == InProgress
	switch {
	case pendingOrActive && strings.HasPrefix(p, humanTag):
		return set(&s.Human, strings.TrimPrefix(p, humanTag))
	case s.State == Pending && strings.HasPrefix(p, blockedTag):
		return set(&s.Blocked, strings.TrimPrefix(p, blockedTag))
	case s.State == Done && hashRE.MatchString(p):
		return set(&s.Commit, p)
	case s.State == Done && dateRE.MatchString(p):
		return set(&s.Date, p)
	case s.State == Done && strings.HasPrefix(p, byTag):
		return set(&s.By, strings.TrimPrefix(p, byTag))
	case s.State == Done && strings.HasPrefix(p, noteTag):
		return set(&s.Note, strings.TrimPrefix(p, noteTag))
	case s.State == Done && p == testsTag:
		if s.Tests {
			return fmt.Errorf("status part %q is repeated", p)
		}
		s.Tests = true
		return nil
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
