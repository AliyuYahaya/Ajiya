package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/detect"
	"github.com/AliyuYahaya/Ajiya/internal/kit"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

func runInit(e *env, args []string) error {
	fs := newFlags("init")
	name := fs.String("name", filepath.Base(e.dir), "project name")
	prefix := fs.String("prefix", "", "ticket ID prefix")
	yes := fs.Bool("yes", false, "register every suggested app without asking")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	if *prefix == "" {
		*prefix = config.SuggestPrefix(*name)
	}
	path := filepath.Join(e.dir, config.FileName)
	if _, err := os.Stat(path); err == nil {
		// Already set up: bring the agent kit up to date, nothing else.
		fmt.Fprintf(e.stdout, "%s already exists: kept as it is.\n", config.FileName)
		if err := installKit(e); err != nil {
			return err
		}
		return offerAgents(e, *yes)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	text := config.Template(*name, *prefix)
	testCmd := detect.TestCommand(e.dir)
	if testCmd != "" {
		text = config.WithTestCommand(text, testCmd)
	}
	if _, err := config.Parse([]byte(text)); err != nil {
		return usageErr("%v", err)
	}
	cands, err := detect.Suggest(e.dir, *name)
	if err != nil {
		return err
	}
	chosen, err := chooseApps(e, cands, *yes)
	if err != nil {
		return err
	}
	for _, c := range chosen {
		text = config.AddAppText(text, config.App{Name: c.Name, Path: c.Path, Kind: c.Kind})
	}
	if _, err := config.Parse([]byte(text)); err != nil {
		return usageErr("%v", err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(e.dir, plan.Dir), 0o755); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "Set up %s with ticket IDs like %s.\nWrote %s and %s/.\n", *name, plan.FormatID(*prefix, 1), config.FileName, plan.Dir)
	if len(chosen) > 0 {
		fmt.Fprintf(e.stdout, "Registered %d app(s); 'infra' is always there for work outside them.\n", len(chosen))
	}
	if testCmd != "" {
		fmt.Fprintf(e.stdout, "Test command set to %q in [test] of %s; change it there if it is wrong.\n", testCmd, config.FileName)
	} else {
		fmt.Fprintf(e.stdout, "No test command found: set one in [test] of %s so 'ajiya ticket done --test' can run it.\n", config.FileName)
	}
	if err := installKit(e); err != nil {
		return err
	}
	fmt.Fprintln(e.stdout, "Next: 'ajiya hook install', then plan with .ajiya/guide/setup.md.")
	// Last, so that when init registers Ajiya the output ends with the undo line.
	return offerAgents(e, *yes)
}

// installKit writes or updates the agent kit and says what changed.
func installKit(e *env) error {
	rs, err := kit.Install(e.dir)
	if err != nil {
		return refused("agent kit not written: %v", err)
	}
	fmt.Fprintln(e.stdout, "Agent kit:")
	w := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
	for _, r := range rs {
		fmt.Fprintf(w, "  %s\t%s\n", r.Action, r.Path)
	}
	return w.Flush()
}

// chooseApps shows the suggested apps and returns the ones to register: all
// with --yes, the person's choice at a terminal, and none otherwise.
func chooseApps(e *env, cands []detect.Candidate, yes bool) ([]detect.Candidate, error) {
	if len(cands) == 0 {
		fmt.Fprintln(e.stdout, "No apps found. Register them with 'ajiya app add <name> --path <dir>'.")
		return nil, nil
	}
	fmt.Fprintln(e.stdout, "Apps found:")
	w := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
	for i, c := range cands {
		why := c.Source
		if len(c.Signs) > 0 {
			why += "; " + strings.Join(c.Signs, ", ")
		}
		fmt.Fprintf(w, "  %d.\t%s\t%s\t%s\t(%s)\n", i+1, c.Name, c.Path, c.Kind, why)
	}
	w.Flush()
	if yes {
		return cands, nil
	}
	if !e.interactive {
		fmt.Fprintln(e.stdout, "Nothing registered: run with --yes to accept these, or register each one:")
		for _, c := range cands {
			lib := ""
			if c.Kind == "library" {
				lib = " --library"
			}
			fmt.Fprintf(e.stdout, "  ajiya app add %s --path %s%s\n", c.Name, c.Path, lib)
		}
		return nil, nil
	}
	fmt.Fprint(e.stdout, "Register which? all, none, or numbers such as 1,3 [all]: ")
	return pickApps(cands, e.readLine())
}

// pickApps reads an answer: "", "all", "none", or numbers separated by commas or spaces.
func pickApps(cands []detect.Candidate, answer string) ([]detect.Candidate, error) {
	answer = strings.ToLower(strings.TrimSpace(answer))
	switch answer {
	case "", "all", "y", "yes":
		return cands, nil
	case "none", "n", "no":
		return nil, nil
	}
	var out []detect.Candidate
	seen := map[int]bool{}
	for _, f := range strings.FieldsFunc(answer, func(r rune) bool { return r == ',' || r == ' ' }) {
		n, err := strconv.Atoi(f)
		if err != nil || n < 1 || n > len(cands) {
			return nil, usageErr("%q is not a number from 1 to %d", f, len(cands))
		}
		if !seen[n] {
			seen[n] = true
			out = append(out, cands[n-1])
		}
	}
	return out, nil
}

func runPhaseAdd(e *env, args []string) error {
	fs := newFlags("phase add")
	title := fs.String("title", "", "phase title (default: from the slug)")
	pos, err := parse(fs, args, 2)
	if err != nil {
		return err
	}
	slug, goal := pos[0], pos[1]
	if !plan.SlugRE.MatchString(slug) {
		return usageErr("slug %q must be lower case letters and digits separated by single hyphens, like go-live", slug)
	}
	if *title == "" {
		*title = plan.TitleFromSlug(slug)
	}
	if goal == "" {
		return usageErr("the goal is empty; say what is true when this phase is done")
	}
	if err := plan.CheckText("goal", goal); err != nil {
		return usageErr("%v", err)
	}
	if err := plan.CheckText("title", *title); err != nil {
		return usageErr("%v", err)
	}
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	if _, exists := pr.plan.Raw[slug]; exists {
		return refused("phase %q already exists in %s", slug, plan.Path(slug))
	}
	ph := &plan.Phase{Slug: slug, Title: *title, Goal: goal}
	if err := pr.plan.Save(ph); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "Added phase %s (%s).\n", slug, plan.Path(slug))
	return nil
}
