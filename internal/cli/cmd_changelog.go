package cli

import (
	"errors"
	"flag"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/changelog"
	"github.com/AliyuYahaya/Ajiya/internal/gitx"
)

func runChangelog(e *env, args []string) error {
	fs := newFlags("changelog")
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parse(fs, args, 1)
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		pos, err = parse(fs, args, 0) // the range is optional
		if err != nil && strings.HasPrefix(err.Error(), "expected") {
			return usageErr("expected at most 1 argument, a revision range such as v0.1.0..HEAD")
		}
	}
	if err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	var log []gitx.Commit
	if len(pos) == 1 {
		if pos[0] == "" {
			return usageErr("the range must not be empty, such as v0.1.0..HEAD")
		}
		if log, err = gitx.LogRange(pr.cfg.Root, pos[0]); err != nil {
			return usageErr("%s: %v", pos[0], err)
		}
	} else if log, err = gitx.Log(pr.cfg.Root); err != nil {
		return err
	}
	cl := changelog.BuildIn(pr.plan, pr.displayPhases(), log)
	if *asJSON {
		return writeJSON(e, cl)
	}
	return cl.WriteMarkdown(e.stdout)
}
