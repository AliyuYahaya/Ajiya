package cli

import (
	"fmt"
	"slices"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// runPhaseRename renames a phase file, moves its tickets and keeps the launch
// target and [phases] order pointing at it.
func runPhaseRename(e *env, args []string) error {
	fs := newFlags("phase rename")
	title := fs.String("title", "", "new phase title (default: keep the current one)")
	pos, err := parse(fs, args, 2)
	if err != nil {
		return err
	}
	old, slug := pos[0], pos[1]
	if !plan.SlugRE.MatchString(slug) {
		return usageErr("slug %q must be lower case letters and digits separated by single hyphens, like go-live", slug)
	}
	if isSet(fs, "title") {
		if *title == "" {
			return usageErr("--title cannot be empty")
		}
		if err := plan.CheckText("title", *title); err != nil {
			return usageErr("%v", err)
		}
	}
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	ph, err := pr.phase(old)
	if err != nil {
		return err
	}
	if _, exists := pr.plan.Raw[slug]; exists {
		return refused("phase %q already exists in %s; pick another slug", slug, plan.Path(slug))
	}
	if *title != "" {
		ph.Title = *title
	}
	if err := pr.plan.Rename(ph, slug); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "Renamed phase %s to %s (%s).\n", old, slug, plan.Path(slug))
	if pr.cfg.Launch.Target == old {
		if err := config.SetLaunchTarget(pr.cfg.Root, slug); err != nil {
			return err
		}
		fmt.Fprintf(e.stdout, "Launch target is now %s.\n", slug)
	}
	if i := slices.Index(pr.cfg.Phases.Order, old); i >= 0 {
		order := slices.Clone(pr.cfg.Phases.Order)
		order[i] = slug
		// A stale entry for the new slug would now be listed twice.
		for j := len(order) - 1; j >= 0; j-- {
			if j != i && order[j] == slug {
				order = slices.Delete(order, j, j+1)
			}
		}
		if _, err := config.SetPhaseOrder(pr.cfg.Root, order); err != nil {
			return err
		}
		fmt.Fprintln(e.stdout, "Updated [phases] order.")
	}
	return nil
}
