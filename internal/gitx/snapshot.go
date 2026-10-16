package gitx

import (
	"errors"
	"fmt"
	"strings"
)

// SnapshotPrefix starts the name of every snapshot tag: "ajiya/plan/<name>".
const SnapshotPrefix = "ajiya/plan/"

// Snapshot is a named point in the plan's history: an annotated git tag.
type Snapshot struct {
	Name    string `json:"name"`
	Tag     string `json:"tag"`    // ajiya/plan/<name>
	Commit  string `json:"commit"` // short hash of the commit it names
	Date    string `json:"date"`   // YYYY-MM-DD, when the tag was made
	Message string `json:"message"`
}

// CheckSnapshotName reports why name cannot be a snapshot name: it must make a
// valid git tag name under ajiya/plan/, and cannot be "list", which is a
// subcommand.
func CheckSnapshotName(dir, name string) error {
	switch {
	case name == "":
		return errors.New("a snapshot needs a name")
	case name == "list":
		return errors.New(`"list" cannot be a snapshot name`)
	case strings.ContainsAny(name, " \t\r\n"):
		return fmt.Errorf("snapshot name %q must not contain spaces", name)
	case strings.HasPrefix(name, "-"):
		return fmt.Errorf("snapshot name %q must not start with '-'", name)
	}
	if _, err := git(dir, nil, "check-ref-format", "refs/tags/"+SnapshotPrefix+name); err != nil {
		if errors.Is(err, ErrNotRepo) {
			return err
		}
		return fmt.Errorf("%q is not a valid git tag name", name)
	}
	return nil
}

// maxDirtyListed is how many changed files a refusal names.
const maxDirtyListed = 8

// CreateSnapshot tags the current commit as ajiya/plan/<name> with an
// annotated tag. It refuses a bad name, a name already taken, a repository with
// no commits, and uncommitted or untracked changes under planDir (the plan
// folder, relative to dir), naming the files. A snapshot is a tag, not a copy:
// the plan is already in git.
func CreateSnapshot(dir, planDir, name string) (Snapshot, error) {
	if err := CheckSnapshotName(dir, name); err != nil {
		return Snapshot{}, err
	}
	ok, err := hasHead(dir)
	if err != nil {
		return Snapshot{}, err
	}
	if !ok {
		return Snapshot{}, errors.New("the repository has no commits yet; commit the plan first")
	}
	tag := SnapshotPrefix + name
	if _, err := git(dir, nil, "rev-parse", "-q", "--verify", "refs/tags/"+tag); err == nil {
		return Snapshot{}, fmt.Errorf("snapshot %q already exists (tag %s)", name, tag)
	}
	out, err := git(dir, nil, "status", "--porcelain", "--untracked-files=all", "--", planDir)
	if err != nil {
		return Snapshot{}, err
	}
	if lines := splitNonEmpty(string(out), "\n"); len(lines) > 0 {
		files := make([]string, 0, len(lines))
		for _, l := range lines {
			if len(l) > 2 { // "XY path"; the first line may have lost its leading space
				l = strings.TrimSpace(l[2:])
			}
			files = append(files, l)
		}
		more := ""
		if len(files) > maxDirtyListed {
			more = fmt.Sprintf(" and %d more", len(files)-maxDirtyListed)
			files = files[:maxDirtyListed]
		}
		return Snapshot{}, fmt.Errorf("%s/ has uncommitted or untracked changes: %s%s; commit them first so the snapshot holds them",
			planDir, strings.Join(files, ", "), more)
	}
	if _, err := git(dir, nil, "tag", "-a", tag, "-m", "Ajiya plan snapshot "+name); err != nil {
		return Snapshot{}, err
	}
	snaps, err := Snapshots(dir)
	if err != nil {
		return Snapshot{}, err
	}
	for _, s := range snaps {
		if s.Name == name {
			return s, nil
		}
	}
	return Snapshot{}, fmt.Errorf("tag %s was not created", tag)
}

// Snapshots lists the snapshot tags, newest first.
func Snapshots(dir string) ([]Snapshot, error) {
	// For an annotated tag, %(*objectname) is the commit and %(objectname) the tag.
	format := "%(refname)%09%(*objectname:short)%09%(objectname:short)%09%(creatordate:short)%09%(contents:subject)"
	out, err := git(dir, nil, "for-each-ref", "--sort=-creatordate", "--format="+format, "refs/tags/"+SnapshotPrefix)
	if err != nil {
		return nil, err
	}
	var snaps []Snapshot
	for _, line := range splitNonEmpty(string(out), "\n") {
		f := strings.SplitN(line, "\t", 5)
		if len(f) < 5 {
			continue
		}
		tag := strings.TrimPrefix(f[0], "refs/tags/")
		commit := f[1]
		if commit == "" { // a lightweight tag points at the commit itself
			commit = f[2]
		}
		snaps = append(snaps, Snapshot{
			Name: strings.TrimPrefix(tag, SnapshotPrefix), Tag: tag, Commit: commit, Date: f[3], Message: f[4],
		})
	}
	return snaps, nil
}
