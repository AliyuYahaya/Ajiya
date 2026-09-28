package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/check"
	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/gitx"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// runHookRun is called by the git hook scripts: ajiya hook run <hook> <args>.
func runHookRun(e *env, args []string) error {
	if len(args) == 0 {
		return usageErr("name a hook: commit-msg or prepare-commit-msg")
	}
	if _, err := config.Find(e.dir); errors.Is(err, config.ErrNotFound) {
		return nil // not an ajiya project: the hook has nothing to enforce
	}
	switch args[0] {
	case "commit-msg":
		if len(args) != 2 {
			return usageErr("commit-msg takes the message file")
		}
		return hookCommitMsg(e, args[1])
	case "prepare-commit-msg":
		if len(args) < 2 || len(args) > 4 {
			return usageErr("prepare-commit-msg takes the message file, and optionally the source and commit")
		}
		source := ""
		if len(args) > 2 {
			source = args[2]
		}
		if err := hookPrepare(e, args[1], source); err != nil {
			// Never block a commit here: commit-msg enforces the rule.
			fmt.Fprintf(e.stderr, "ajiya: could not pre-fill the trailer: %v\n", err)
		}
		return nil
	}
	return usageErr("unknown hook %q", args[0])
}

// hookCommitMsg enforces the commit rule on a message about to be committed.
func hookCommitMsg(e *env, file string) error {
	pr, err := load(e)
	if err != nil {
		return err
	}
	root := pr.cfg.Root
	if merging, err := inMerge(root); err != nil || merging {
		return err
	}
	msg, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	refs, err := gitx.ParseMessage(root, msg)
	if err != nil {
		return err
	}
	var unknown []string
	for _, id := range refs.IDs {
		if pr.plan.Ticket(id) == nil {
			unknown = append(unknown, id)
		}
	}
	switch n := len(refs.IDs); {
	case len(unknown) > 0:
		return refused("commit refused: the trailer names %s, which does not exist; see 'ajiya next' for tickets", strings.Join(unknown, ", "))
	case refs.Chore || refs.Revert:
		return nil
	case n == 0:
		return refused("commit refused: the message names no ticket. End it with a paragraph 'Ajiya: <ID>' (1 to %d IDs), or 'Ajiya: chore' for upkeep.%s",
			check.MaxRefs, inProgressHint(pr))
	case n > check.MaxRefs:
		return refused("commit refused: the message names %d tickets, more than %d; split the work into smaller commits", n, check.MaxRefs)
	}
	return nil
}

// inMerge reports whether a merge is being committed.
func inMerge(root string) (bool, error) {
	p, err := gitx.GitPath(root, "MERGE_HEAD")
	if err != nil {
		return false, err
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(root, p)
	}
	_, err = os.Stat(p)
	return err == nil, nil
}

// inProgress returns the IDs of tickets in progress, sorted.
func inProgress(pr *project) []string {
	var ids []string
	for _, t := range pr.plan.Tickets() {
		if t.Status.State == plan.InProgress {
			ids = append(ids, t.ID)
		}
	}
	return ids
}

func inProgressHint(pr *project) string {
	ids := inProgress(pr)
	if len(ids) == 0 {
		return " No ticket is in progress; start one with 'ajiya ticket start <ID>'."
	}
	return fmt.Sprintf(" In progress: %s.", strings.Join(ids, ", "))
}

// hookPrepare pre-fills the Ajiya trailer before the message is edited: with the
// tickets whose rows changed when only ajiya/ files are staged, otherwise with
// the tickets in progress. Merges, squashes, amends and messages that already
// have a trailer are left alone.
func hookPrepare(e *env, file, source string) error {
	switch source {
	case "merge", "squash", "commit":
		return nil
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	root := pr.cfg.Root
	msg, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	refs, err := gitx.ParseMessage(root, msg)
	if err != nil {
		return err
	}
	if len(refs.IDs) > 0 || refs.Chore {
		return nil
	}
	ids, err := trailerCandidates(pr)
	if err != nil || len(ids) == 0 {
		return err
	}
	if len(ids) > check.MaxRefs {
		if source == "message" {
			return nil // no editor will open: let commit-msg explain
		}
		hint := fmt.Sprintf("# Ajiya: name 1 to %d of %s\n", check.MaxRefs, strings.Join(ids, ", "))
		return os.WriteFile(file, append([]byte(hint), msg...), 0o644)
	}
	value := strings.Join(ids, ", ")
	if strings.TrimSpace(stripComments(string(msg))) == "" {
		// Leave line 1 for the subject and a blank line under it, so the
		// trailer stays in a paragraph of its own.
		err = os.WriteFile(file, []byte("\n\n"+gitx.TrailerKey+": "+value+"\n"+string(msg)), 0o644)
	} else {
		err = gitx.AddTrailer(root, file, value)
	}
	if err == nil && source == "message" {
		fmt.Fprintf(e.stderr, "ajiya: added 'Ajiya: %s' to the commit message\n", value)
	}
	return err
}

func stripComments(msg string) string {
	var b strings.Builder
	for line := range strings.SplitSeq(msg, "\n") {
		if !strings.HasPrefix(line, "#") {
			b.WriteString(line + "\n")
		}
	}
	return b.String()
}

// trailerCandidates returns the tickets a commit is probably about.
func trailerCandidates(pr *project) ([]string, error) {
	root := pr.cfg.Root
	staged, err := gitx.StagedFiles(root)
	if err != nil {
		return nil, err
	}
	onlyPlan := len(staged) > 0
	for _, f := range staged {
		if !strings.HasPrefix(f, plan.Dir+"/") {
			onlyPlan = false
		}
	}
	if !onlyPlan {
		return inProgress(pr), nil
	}
	seen := map[string]bool{}
	for _, f := range staged {
		slug, ok := strings.CutSuffix(strings.TrimPrefix(f, plan.Dir+"/"), ".md")
		if !ok || !plan.SlugRE.MatchString(slug) {
			continue
		}
		before := rowsAt(root, "HEAD:"+f, slug)
		after := rowsAt(root, ":"+f, slug)
		for id, row := range after {
			if before[id] != row {
				seen[id] = true
			}
		}
		for id := range before {
			if _, kept := after[id]; !kept {
				seen[id] = true
			}
		}
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	plan.SortIDs(ids)
	return ids, nil
}

// rowsAt returns each ticket's formatted row in a phase file at a revision.
func rowsAt(root, spec, slug string) map[string]string {
	rows := map[string]string{}
	data, ok, err := gitx.Show(root, spec)
	if err != nil || !ok {
		return rows
	}
	ph, err := plan.ParsePhase(slug, data)
	if err != nil {
		return rows
	}
	for _, t := range ph.Tickets {
		one := &plan.Phase{Tickets: []*plan.Ticket{t}}
		rows[t.ID] = string(one.Format())
	}
	return rows
}
