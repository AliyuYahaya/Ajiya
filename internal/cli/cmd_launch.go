package cli

import (
	"fmt"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// runLaunchSet sets the targets of the milestone named launch to one target.
// A file without [[milestones]] keeps the older [launch] target form; a file
// with them gets its launch milestone changed, or added at the end.
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
		return refused("launch %v; give a phase slug or a ticket ID", err)
	}
	if len(pr.cfg.Milestones) > 0 {
		if err := editConfig(pr, func(text string) (string, error) {
			return config.SetMilestoneTargetsText(text, config.LaunchMilestone, []string{target})
		}); err != nil {
			return err
		}
	} else {
		if err := config.SetLaunchTarget(pr.cfg.Root, target); err != nil {
			return err
		}
		pr.cfg.Launch.Target = target
	}
	return showLaunch(e, pr)
}

func runLaunchShow(e *env, args []string) error {
	fs := newFlags("launch show")
	asJSON := fs.Bool("json", false, "print JSON")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	if *asJSON {
		return launchJSON(e, pr)
	}
	return showLaunch(e, pr)
}

// launchRequired returns the targets of the launch milestone, each with its
// kind, and its required set. It returns no targets when there is no launch
// milestone.
func launchRequired(pr *project) (targets, kinds []string, req map[string]bool, err error) {
	m := pr.milestone(config.LaunchMilestone)
	if m == nil {
		return nil, nil, nil, nil
	}
	req = map[string]bool{}
	for _, t := range m.Targets {
		r, kind, err := pr.plan.Required(t)
		if err != nil {
			return nil, nil, nil, refused("launch %v; set another with 'ajiya launch set <phase|ticket>'", err)
		}
		for id := range r {
			req[id] = true
		}
		kinds = append(kinds, kind)
	}
	return m.Targets, kinds, req, nil
}

func launchJSON(e *env, pr *project) error {
	out := struct {
		Target      *string      `json:"target"` // the first target; null when none is set
		Targets     []string     `json:"targets"`
		Kind        string       `json:"kind,omitempty"`
		Required    int          `json:"required"`
		Closed      int          `json:"closed"`
		Percent     int          `json:"percent"`
		AfterLaunch int          `json:"after_launch"`
		Open        []jsonTicket `json:"open"`
	}{Targets: []string{}, Open: []jsonTicket{}}
	targets, kinds, req, err := launchRequired(pr)
	if err != nil {
		return err
	}
	if len(targets) > 0 {
		out.Target, out.Targets, out.Kind, out.Required = &targets[0], targets, kinds[0], len(req)
		for _, t := range pr.plan.Tickets() {
			switch {
			case !req[t.ID]:
				out.AfterLaunch++
			case t.Status.Closed():
				out.Closed++
			default:
				out.Open = append(out.Open, toJSON(t))
			}
		}
		out.Percent = percent(out.Closed, out.Required)
	} else {
		out.AfterLaunch = len(pr.plan.Tickets())
	}
	return writeJSON(e, out)
}

func showLaunch(e *env, pr *project) error {
	targets, kinds, req, err := launchRequired(pr)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		fmt.Fprintln(e.stdout, "No launch target. Set one with 'ajiya launch set <phase|ticket>'.")
		return nil
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
	shown := make([]string, len(targets))
	for i := range targets {
		shown[i] = fmt.Sprintf("%s (%s)", targets[i], kinds[i])
	}
	total := len(req)
	fmt.Fprintf(e.stdout, "Launch target: %s\n", strings.Join(shown, ", "))
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
