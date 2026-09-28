package cli

import (
	"errors"
	"flag"
	"io"
	"strings"
	"time"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// newFlags returns a flag set that reports errors through parse, not stderr.
func newFlags(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

// parse parses flags that may appear before, between or after positional
// arguments, and returns the positional arguments. Everything after "--" is
// positional.
func parse(fs *flag.FlagSet, args []string, npos int) ([]string, error) {
	var pos []string
	for {
		if err := fs.Parse(args); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil, err
			}
			return nil, usageErr("%v", err)
		}
		args = fs.Args()
		if len(args) == 0 {
			break
		}
		if args[0] == "--" {
			pos = append(pos, args[1:]...)
			break
		}
		pos = append(pos, args[0])
		args = args[1:]
	}
	if len(pos) != npos {
		return nil, usageErr("expected %d argument(s), got %d; run with --help for usage", npos, len(pos))
	}
	return pos, nil
}

// isSet reports whether a flag was given on the command line.
func isSet(fs *flag.FlagSet, name string) bool {
	set := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			set = true
		}
	})
	return set
}

// project is a loaded config and plan.
type project struct {
	cfg  *config.Config
	plan *plan.Plan
}

// load finds ajiya.toml from the working directory and reads the plan.
func load(e *env) (*project, error) {
	root, err := config.Find(e.dir)
	if err != nil {
		return nil, usageErr("%v", err)
	}
	cfg, err := config.Load(root)
	if err != nil {
		return nil, usageErr("%v", err)
	}
	p, err := plan.Load(root)
	if err != nil {
		return nil, err
	}
	return &project{cfg: cfg, plan: p}, nil
}

// loadForChange loads the project and refuses if a phase file cannot be read,
// because writing the plan back would lose that file's contents.
func loadForChange(e *env) (*project, error) {
	pr, err := load(e)
	if err != nil {
		return nil, err
	}
	for slug, pe := range pr.plan.Broken {
		return nil, refused("%s:%d: %s; fix this file before changing the plan (see 'ajiya check')", plan.Path(slug), pe.Line, pe.Msg)
	}
	return pr, nil
}

func (pr *project) ticket(id string) (*plan.Ticket, error) {
	t := pr.plan.Ticket(id)
	if t == nil {
		return nil, refused("no ticket %s; see 'ajiya next' or the files in ajiya/", id)
	}
	return t, nil
}

// checkDepends validates a dependency list for ticket id ("" for a new ticket).
func (pr *project) checkDepends(id string, deps []string) error {
	for _, d := range deps {
		if d == id {
			return refused("%s cannot depend on itself", id)
		}
		if pr.plan.Ticket(d) == nil {
			return refused("dependency %s does not exist", d)
		}
	}
	if id != "" {
		if loop := plan.LoopWith(pr.plan.Tickets(), id, deps); loop != nil {
			return refused("that would make a dependency loop: %s", strings.Join(loop, " -> "))
		}
	}
	return nil
}

func (pr *project) checkApp(app string) error {
	if !pr.cfg.HasApp(app) {
		return refused("app %q is not registered; use one of %s, or add it to ajiya.toml", app, strings.Join(pr.cfg.AppNames(), ", "))
	}
	return nil
}

func (pr *project) phase(slug string) (*plan.Phase, error) {
	ph := pr.plan.Phase(slug)
	if ph == nil {
		var slugs []string
		for _, p := range pr.plan.Phases {
			slugs = append(slugs, p.Slug)
		}
		if len(slugs) == 0 {
			return nil, refused("no phase %q; add one with 'ajiya phase add %s \"<goal>\"'", slug, slug)
		}
		return nil, refused("no phase %q; phases are %s", slug, strings.Join(slugs, ", "))
	}
	return ph, nil
}

// saveTicket writes the phase holding t.
func (pr *project) saveTicket(t *plan.Ticket) error {
	return pr.plan.Save(pr.plan.Phase(t.Phase))
}

func today() string { return time.Now().Format("2006-01-02") }

// ensurePhase returns the phase with slug, creating it with goal if needed.
// The caller saves it.
func (pr *project) ensurePhase(slug, goal string) *plan.Phase {
	if ph := pr.plan.Phase(slug); ph != nil {
		return ph
	}
	ph := &plan.Phase{Slug: slug, Title: plan.TitleFromSlug(slug), Goal: goal}
	pr.plan.Phases = append(pr.plan.Phases, ph)
	return ph
}

// newTicket adds a ticket with the next ID to a phase and returns it. The
// caller validates the fields and saves the phase. Importers use it too.
func (pr *project) newTicket(ph *plan.Phase, app, title, doneWhen string, deps []string, status plan.Status) *plan.Ticket {
	t := &plan.Ticket{
		ID:       pr.plan.NextID(pr.cfg.Project.Prefix),
		App:      app,
		Title:    title,
		DoneWhen: doneWhen,
		Depends:  deps,
		Status:   status,
		Phase:    ph.Slug,
	}
	ph.Tickets = append(ph.Tickets, t)
	return t
}
