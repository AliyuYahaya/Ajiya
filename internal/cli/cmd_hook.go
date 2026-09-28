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
		return usageErr("name a hook: commit-msg")
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
