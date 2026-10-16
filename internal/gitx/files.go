package gitx

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// RepoFile is a file changed by a commit, and the repository it is in.
type RepoFile struct {
	Repo string // the work tree
	Path string // relative to Repo, with / separators
}

// TicketFiles is what FilesByTicket finds for one ticket.
type TicketFiles struct {
	Commits int        // commits naming the ticket, merge commits included
	Files   []RepoFile // sorted by repository, then path, no duplicates
}

// FilesByTicket reads the history of every repo once and returns, for each
// ticket ID named in an Ajiya trailer, the union of the files its commits
// changed. A commit naming several tickets counts for each. Merge commits add
// no files (git shows no diff for them), and renames are listed as the new
// path. Files under any directory in skip (absolute paths) are left out.
// Repositories without commits add nothing; with no repos it fails with
// ErrNotRepo. Tickets no commit names are absent from the result.
func FilesByTicket(repos, skip []string) (map[string]*TicketFiles, error) {
	if len(repos) == 0 {
		return nil, ErrNotRepo
	}
	out := map[string]*TicketFiles{}
	seen := map[string]map[RepoFile]bool{}
	for _, repo := range repos {
		ok, err := hasHead(repo)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		prefixes := skipIn(repo, skip)
		format := "--format=%x1e%H%x1f%P%x1f%(trailers:key=" + TrailerKey + ",valueonly,unfold,separator=%x1f)"
		data, err := git(repo, nil, "log", "--no-renames", "--name-only", "-z", format, "HEAD", "--")
		if err != nil {
			return nil, err
		}
		for rec := range strings.SplitSeq(string(data), "\x1e") {
			if rec == "" {
				continue
			}
			header, rest, _ := strings.Cut(rec, "\x00")
			f := strings.Split(header, "\x1f")
			if len(f) < 3 {
				return nil, fmt.Errorf("git log: unexpected output %q", header)
			}
			ids := parseValues(f[2:]).IDs
			if len(ids) == 0 {
				continue
			}
			var files []string
			for p := range strings.SplitSeq(strings.TrimPrefix(rest, "\n"), "\x00") {
				if p != "" && !underAny(p, prefixes) {
					files = append(files, p)
				}
			}
			for _, id := range ids {
				tf := out[id]
				if tf == nil {
					tf = &TicketFiles{}
					out[id] = tf
					seen[id] = map[RepoFile]bool{}
				}
				tf.Commits++
				for _, p := range files {
					rf := RepoFile{repo, p}
					if !seen[id][rf] {
						seen[id][rf] = true
						tf.Files = append(tf.Files, rf)
					}
				}
			}
		}
	}
	for _, tf := range out {
		sort.Slice(tf.Files, func(i, j int) bool {
			a, b := tf.Files[i], tf.Files[j]
			if a.Repo != b.Repo {
				return a.Repo < b.Repo
			}
			return a.Path < b.Path
		})
	}
	return out, nil
}

// resolve follows symlinks in the longest existing part of path, so a folder
// that does not exist yet still compares equal to one inside a resolved repo.
func resolve(path string) string {
	var tail []string
	for p := path; ; p = filepath.Dir(p) {
		if r, err := filepath.EvalSymlinks(p); err == nil {
			return filepath.Join(append([]string{r}, tail...)...)
		}
		if filepath.Dir(p) == p {
			return path
		}
		tail = append([]string{filepath.Base(p)}, tail...)
	}
}

// skipIn turns absolute directories into prefixes ("dir/") relative to repo,
// keeping those inside it.
func skipIn(repo string, skip []string) []string {
	repo = resolve(repo)
	var out []string
	for _, s := range skip {
		rel, err := filepath.Rel(repo, resolve(s))
		if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		out = append(out, filepath.ToSlash(rel)+"/")
	}
	return out
}

func underAny(p string, prefixes []string) bool {
	for _, pre := range prefixes {
		if strings.HasPrefix(p, pre) {
			return true
		}
	}
	return false
}
