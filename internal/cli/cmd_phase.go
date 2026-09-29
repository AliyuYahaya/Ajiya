package cli

import (
	"fmt"
	"slices"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// runPhaseRemove deletes an empty phase: its file and its [phases] order entry.
// It refuses while the phase holds a ticket or a milestone or the launch target
// names it.
func runPhaseRemove(e *env, args []string) error {
	fs := newFlags("phase remove")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	slug := pos[0]
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	ph, err := pr.phase(slug)
	if err != nil {
		return err
	}
	if n := len(ph.Tickets); n > 0 {
		return refused("phase %s still holds %d ticket(s); move them with 'ajiya ticket edit <ID> --phase <slug>' first", slug, n)
	}
	if pr.cfg.Launch.Target == slug {
		return refused("the launch target is phase %s; point it elsewhere with 'ajiya launch set <phase|ticket>' first", slug)
	}
	for _, m := range pr.cfg.Milestones {
		if slices.Contains(m.Targets, slug) {
			return refused("milestone %s targets phase %s; remove that milestone with 'ajiya milestone remove %s' or re-create it with other targets first", m.Name, slug, m.Name)
		}
	}
	// Drop the order entry first: a stale entry would name a missing phase,
	// while a phase file left behind is still a valid (empty) phase.
	if i := slices.Index(pr.cfg.Phases.Order, slug); i >= 0 {
		order := slices.Delete(slices.Clone(pr.cfg.Phases.Order), i, i+1)
		if _, err := config.SetPhaseOrder(pr.cfg.Root, order); err != nil {
			return err
		}
		fmt.Fprintln(e.stdout, "Updated [phases] order.")
	}
	if err := pr.plan.Remove(ph); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "Removed phase %s (%s).\n", slug, plan.Path(slug))
	return nil
}

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
	for _, m := range pr.cfg.Milestones {
		if !slices.Contains(m.Targets, old) {
			continue
		}
		targets := slices.Clone(m.Targets)
		for i, t := range targets {
			if t == old {
				targets[i] = slug
			}
		}
		name := m.Name
		if err := writeConfig(pr, func(text string) (string, bool) {
			out, err := config.SetMilestoneTargetsText(text, name, slices.Compact(targets))
			return out, err == nil
		}); err != nil {
			return err
		}
		fmt.Fprintf(e.stdout, "Milestone %s now targets %s.\n", name, slug)
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
