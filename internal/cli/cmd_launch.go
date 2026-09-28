package cli

import (
	"fmt"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

func runLaunchSet(e *env, args []string) error {
	pos, err := parse(newFlags("launch set"), args, 1)
	if err != nil {
		return err
	}
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	target := pos[0]
	if _, _, err := pr.plan.Required(target); err != nil {
		return refused("%v; give a phase slug or a ticket ID", err)
	}
	if err := config.SetLaunchTarget(pr.cfg.Root, target); err != nil {
		return err
	}
	pr.cfg.Launch.Target = target
	return showLaunch(e, pr)
}

func runLaunchShow(e *env, args []string) error {
	if _, err := parse(newFlags("launch show"), args, 0); err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	return showLaunch(e, pr)
}

func showLaunch(e *env, pr *project) error {
	target := pr.cfg.Launch.Target
	if target == "" {
		fmt.Fprintln(e.stdout, "No launch target. Set one with 'ajiya launch set <phase|ticket>'.")
		return nil
	}
	req, kind, err := pr.plan.Required(target)
	if err != nil {
		return refused("%v; set another with 'ajiya launch set <phase|ticket>'", err)
	}
	closed, open := 0, []*plan.Ticket{}
	for _, t := range pr.plan.Tickets() {
		if !req[t.ID] {
			continue
		}
		if t.Status.Closed() {
			closed++
		} else {
			open = append(open, t)
		}
	}
	total := len(req)
	fmt.Fprintf(e.stdout, "Launch target: %s (%s)\n", target, kind)
	fmt.Fprintf(e.stdout, "Required: %d of %d closed (%d%%)\n", closed, total, percent(closed, total))
	fmt.Fprintf(e.stdout, "After launch: %d tickets\n", len(pr.plan.Tickets())-total)
	if len(open) > 0 {
		fmt.Fprintln(e.stdout, "Open and required:")
		for _, t := range open {
			fmt.Fprintf(e.stdout, "  %s %s  %s\n", t.Status.Mark(), t.ID, t.Title)
		}
	}
	return nil
}

func percent(n, total int) int {
	if total == 0 {
		return 100
	}
	return n * 100 / total
}
