package cli

import (
	"fmt"

	"github.com/AliyuYahaya/Ajiya/internal/build"
)

// runBuild writes ajiya/PROGRESS.md and ajiya/data.js.
func runBuild(e *env, args []string) error {
	if _, err := parse(newFlags("build"), args, 0); err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	d, err := build.Collect(pr.cfg, pr.plan, nil)
	if err != nil {
		return err
	}
	wrote, err := build.Write(pr.cfg.Root, d)
	if err != nil {
		return err
	}
	if len(wrote) == 0 {
		fmt.Fprintln(e.stdout, "Up to date.")
		return nil
	}
	for _, f := range wrote {
		fmt.Fprintf(e.stdout, "Wrote %s\n", f)
	}
	return nil
}
