// Package gitx reads the commit history ajiya needs, through the git command.
//
// Trailers are always read by git itself (git interpret-trailers for a message
// file, %(trailers) for history), so the hooks, 'ticket done' and 'check'
// agree with git and with each other on what counts as a trailer.
package gitx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"strings"
)

// TrailerKey is the commit trailer that names tickets: "Ajiya: CB-0007, CB-0008".
// Git matches trailer keys without regard to case.
const TrailerKey = "Ajiya"

// Refs is what a commit message says about tickets.
type Refs struct {
	IDs    []string // ticket IDs named in Ajiya trailers, in order, no duplicates
	Chore  bool     // an "Ajiya: chore" trailer
	Revert bool     // a message written by git revert
}

// Commit is one commit and the tickets its message names.
type Commit struct {
	Hash    string // full hash
	Date    string // committer date, YYYY-MM-DD
	Subject string
	Merge   bool // more than one parent
	Refs
}

// Exempt reports whether the commit rule does not apply: merges, reverts and
// commits marked "Ajiya: chore".
func (c Commit) Exempt() bool { return c.Merge || c.Revert || c.Chore }

// Short returns the first seven characters of the hash.
func (c Commit) Short() string {
	if len(c.Hash) > 7 {
		return c.Hash[:7]
	}
	return c.Hash
}

// ErrNotRepo means the directory is not inside a git work tree.
var ErrNotRepo = errors.New("not a git repository; run 'git init' first")

// git runs a git command in dir, with stdin if not nil, and returns its output.
func git(dir string, stdin io.Reader, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdin = stdin
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, errors.New("git is not installed or not on PATH")
		}
		msg := strings.TrimSpace(stderr.String())
		if strings.Contains(msg, "not a git repository") {
			return nil, ErrNotRepo
		}
		return nil, fmt.Errorf("git %s: %s", args[0], firstLine(msg))
	}
	return out, nil
}

func firstLine(s string) string {
	s, _, _ = strings.Cut(s, "\n")
	return s
}

// hasHead reports whether the repository has at least one commit.
func hasHead(dir string) (bool, error) {
	if _, err := git(dir, nil, "rev-parse", "--is-inside-work-tree"); err != nil {
		return false, err
	}
	_, err := git(dir, nil, "rev-parse", "-q", "--verify", "HEAD")
	return err == nil, nil
}

// ParseMessage reads the Ajiya trailers of a commit message, as git would.
// Only the last paragraph can hold trailers; comment lines are ignored.
func ParseMessage(dir string, msg []byte) (Refs, error) {
	out, err := git(dir, bytes.NewReader(msg), "interpret-trailers", "--parse")
	if err != nil {
		return Refs{}, err
	}
	var values []string
	for line := range strings.SplitSeq(string(out), "\n") {
		key, value, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(key), TrailerKey) {
			values = append(values, value)
		}
	}
	r := parseValues(values)
	r.Revert = IsRevert(subjectOf(string(msg)))
	return r, nil
}

// subjectOf returns the first line of a message that is not blank.
func subjectOf(msg string) string {
	for line := range strings.SplitSeq(msg, "\n") {
		if strings.TrimSpace(line) != "" {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

// IsRevert reports whether a subject is one git revert writes.
func IsRevert(subject string) bool {
	return strings.HasPrefix(subject, `Revert "`) || strings.HasPrefix(subject, `Reapply "`)
}

// Log returns the commits reachable from HEAD, newest first.
func Log(dir string) ([]Commit, error) {
	ok, err := hasHead(dir)
	if err != nil || !ok {
		return nil, err
	}
	return logRevs(dir, "HEAD")
}

// LogRange returns the commits in a revision range such as main..HEAD, newest first.
func LogRange(dir, revRange string) ([]Commit, error) {
	if strings.HasPrefix(revRange, "-") {
		return nil, fmt.Errorf("range %q must not start with '-'", revRange)
	}
	if _, err := git(dir, nil, "rev-parse", "--is-inside-work-tree"); err != nil {
		return nil, err
	}
	return logRevs(dir, revRange, "--")
}

func logRevs(dir string, revs ...string) ([]Commit, error) {
	format := "--format=%H%x1f%cs%x1f%P%x1f%s%x1f%(trailers:key=" + TrailerKey + ",valueonly,unfold,separator=%x1f)%x1e"
	out, err := git(dir, nil, append([]string{"log", format}, revs...)...)
	if err != nil {
		return nil, err
	}
	var commits []Commit
	for rec := range strings.SplitSeq(string(out), "\x1e") {
		rec = strings.Trim(rec, "\n")
		if rec == "" {
			continue
		}
		f := strings.Split(rec, "\x1f")
		if len(f) < 4 {
			return nil, fmt.Errorf("git log: unexpected output %q", rec)
		}
		c := Commit{Hash: f[0], Date: f[1], Merge: len(strings.Fields(f[2])) > 1, Subject: f[3]}
		c.Refs = parseValues(f[4:])
		c.Revert = IsRevert(c.Subject)
		commits = append(commits, c)
	}
	return commits, nil
}

// parseValues reads the values of Ajiya trailers: IDs separated by commas or
// spaces, or the word chore.
func parseValues(values []string) Refs {
	var r Refs
	for _, v := range values {
		for _, f := range strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' }) {
			switch {
			case strings.EqualFold(f, "chore"):
				r.Chore = true
			case !slices.Contains(r.IDs, f):
				r.IDs = append(r.IDs, f)
			}
		}
	}
	return r
}

// Referencing returns the commits whose trailer names id, newest first.
func Referencing(dir, id string) ([]Commit, error) {
	all, err := Log(dir)
	if err != nil {
		return nil, err
	}
	var out []Commit
	for _, c := range all {
		if slices.Contains(c.IDs, id) {
			out = append(out, c)
		}
	}
	return out, nil
}
