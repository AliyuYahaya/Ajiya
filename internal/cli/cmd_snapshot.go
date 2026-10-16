package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/AliyuYahaya/Ajiya/internal/gitx"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// runSnapshot tags the current commit as ajiya/plan/<name>. The plan is in
// version control, so a snapshot is a tag, not a copy.
func runSnapshot(e *env, args []string) error {
	fs := newFlags("snapshot")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	s, err := gitx.CreateSnapshot(pr.cfg.Root, plan.Dir, pos[0])
	if err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "Snapshot %s: tag %s at %s.\n", s.Name, s.Tag, s.Commit)
	return nil
}

func runSnapshotList(e *env, args []string) error {
	fs := newFlags("snapshot list")
	asJSON := fs.Bool("json", false, "print JSON")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	snaps, err := gitx.Snapshots(pr.cfg.Root)
	if err != nil {
		return err
	}
	if snaps == nil {
		snaps = []gitx.Snapshot{}
	}
	if *asJSON {
		return writeJSON(e, snaps)
	}
	if len(snaps) == 0 {
		fmt.Fprintln(e.stdout, "No snapshots yet. Make one with 'ajiya snapshot <name>'.")
		return nil
	}
	w := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SNAPSHOT\tDATE\tCOMMIT")
	for _, s := range snaps {
		fmt.Fprintf(w, "%s\t%s\t%s\n", s.Name, s.Date, s.Commit)
	}
	return w.Flush()
}
