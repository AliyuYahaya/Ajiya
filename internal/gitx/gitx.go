// Package gitx reads the commit history ajiya needs, through the git command.
package gitx

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

// TrailerKey is the commit trailer that names tickets: "Ajiya: CB-0007, CB-0008".
const TrailerKey = "Ajiya"

// Commit is one commit and the tickets its trailer names.
type Commit struct {
	Hash  string   // full hash
	Date  string   // committer date, YYYY-MM-DD
	Refs  []string // ticket IDs named in Ajiya trailers, in order, no duplicates
	Chore bool     // an "Ajiya: chore" trailer
}

// Short returns the first seven characters of the hash.
func (c Commit) Short() string {
	if len(c.Hash) > 7 {
		return c.Hash[:7]
	}
	return c.Hash
}

// ErrNotRepo means the directory is not inside a git work tree.
var ErrNotRepo = errors.New("not a git repository; run 'git init' first")

// git runs a git command in dir and returns its standard output.
func git(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
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
	if _, err := git(dir, "rev-parse", "--is-inside-work-tree"); err != nil {
		return false, err
	}
	_, err := git(dir, "rev-parse", "-q", "--verify", "HEAD")
	return err == nil, nil
}

// Log returns the commits reachable from HEAD, newest first. The trailers are
// read by git itself, so ajiya and git always agree on what a trailer is.
func Log(dir string) ([]Commit, error) {
	ok, err := hasHead(dir)
	if err != nil || !ok {
		return nil, err
	}
	out, err := git(dir, "log", "--format=%H%x1f%cs%x1f%(trailers:key="+TrailerKey+",valueonly,separator=%x1f)%x1e", "HEAD")
	if err != nil {
		return nil, err
	}
	var commits []Commit
	for rec := range strings.SplitSeq(string(out), "\x1e") {
		rec = strings.TrimLeft(rec, "\n")
		if rec == "" {
			continue
		}
		f := strings.Split(strings.TrimRight(rec, "\n"), "\x1f")
		if len(f) < 2 {
			return nil, fmt.Errorf("git log: unexpected output %q", rec)
		}
		c := Commit{Hash: f[0], Date: f[1]}
		c.Refs, c.Chore = ParseTrailerValues(f[2:])
		commits = append(commits, c)
	}
	return commits, nil
}

// ParseTrailerValues reads the values of Ajiya trailers: IDs separated by
// commas or spaces, or the word chore.
func ParseTrailerValues(values []string) (refs []string, chore bool) {
	seen := map[string]bool{}
	for _, v := range values {
		for _, f := range strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' }) {
			switch {
			case strings.EqualFold(f, "chore"):
				chore = true
			case !seen[f]:
				seen[f] = true
				refs = append(refs, f)
			}
		}
	}
	return refs, chore
}

// Referencing returns the commits whose trailer names id, newest first.
func Referencing(dir, id string) ([]Commit, error) {
	all, err := Log(dir)
	if err != nil {
		return nil, err
	}
	var out []Commit
	for _, c := range all {
		if slices.Contains(c.Refs, id) {
			out = append(out, c)
		}
	}
	return out, nil
}
