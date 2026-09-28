package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

func runAppAdd(e *env, args []string) error {
	fs := newFlags("app add")
	path := fs.String("path", "", "the app's folder, relative to the project root")
	library := fs.Bool("library", false, "a shared package rather than an app that runs on its own")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	if *path == "" {
		return usageErr("--path is required; use --path . for the repository root")
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	name := pos[0]
	if pr.cfg.HasApp(name) {
		return refused("app %q is already registered", name)
	}
	a := config.App{Name: name, Path: filepath.ToSlash(*path)}
	if *library {
		a.Kind = "library"
	}
	if err := writeConfig(pr, func(text string) (string, bool) { return config.AddAppText(text, a), true }); err != nil {
		return err
	}
	warnMissingDir(e, pr, a.Path)
	fmt.Fprintf(e.stdout, "Registered app %s at %s.\n", name, a.Path)
	return nil
}

func runAppMove(e *env, args []string) error {
	fs := newFlags("app move")
	path := fs.String("path", "", "the app's new folder")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	if *path == "" {
		return usageErr("--path is required")
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	name, p := pos[0], filepath.ToSlash(*path)
	if err := pr.checkRegistered(name); err != nil {
		return err
	}
	if err := writeConfig(pr, func(text string) (string, bool) { return config.SetAppPathText(text, name, p) }); err != nil {
		return err
	}
	warnMissingDir(e, pr, p)
	fmt.Fprintf(e.stdout, "App %s is now at %s.\n", name, p)
	return nil
}

// runAppRemove unregisters an app. Its open tickets must move to another app
// with --to; closed tickets keep the app name as a record of where work was done.
func runAppRemove(e *env, args []string) error {
	fs := newFlags("app remove")
	to := fs.String("to", "", "move the app's open tickets to this app")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	name := pos[0]
	if err := pr.checkRegistered(name); err != nil {
		return err
	}
	var open []*plan.Ticket
	for _, t := range pr.plan.Tickets() {
		if t.App == name && !t.Status.Closed() {
			open = append(open, t)
		}
	}
	if len(open) > 0 && *to == "" {
		ids := make([]string, len(open))
		for i, t := range open {
			ids[i] = t.ID
		}
		return refused("app %s has open tickets (%s); move them with --to <app>", name, strings.Join(ids, ", "))
	}
	if *to != "" {
		if *to == name {
			return usageErr("--to must name another app")
		}
		if err := pr.checkApp(*to); err != nil {
			return err
		}
	}
	if err := writeConfig(pr, func(text string) (string, bool) { return config.RemoveAppText(text, name) }); err != nil {
		return err
	}
	for _, t := range open {
		t.App = *to
		if err := pr.saveTicket(t); err != nil {
			return err
		}
	}
	fmt.Fprintf(e.stdout, "Removed app %s.", name)
	if len(open) > 0 {
		fmt.Fprintf(e.stdout, " Moved %d open ticket(s) to %s.", len(open), *to)
	}
	fmt.Fprintln(e.stdout)
	return nil
}

// checkRegistered refuses names that are not in ajiya.toml, infra included.
func (pr *project) checkRegistered(name string) error {
	if name == config.InfraApp {
		return refused("%s is built in and cannot be changed", config.InfraApp)
	}
	if !pr.cfg.HasApp(name) {
		return refused("app %q is not registered; registered apps: %s", name, strings.Join(pr.cfg.AppNames(), ", "))
	}
	return nil
}

// writeConfig applies an edit to ajiya.toml, validates it and writes it.
func writeConfig(pr *project, edit func(string) (string, bool)) error {
	text, err := config.ReadText(pr.cfg.Root)
	if err != nil {
		return err
	}
	text, ok := edit(text)
	if !ok {
		return refused("could not find that entry in %s; edit it by hand", config.FileName)
	}
	cfg, err := config.WriteText(pr.cfg.Root, text)
	if err != nil {
		return usageErr("%v", err)
	}
	pr.cfg = cfg
	return nil
}

func warnMissingDir(e *env, pr *project, p string) {
	if st, err := os.Stat(filepath.Join(pr.cfg.Root, filepath.FromSlash(p))); err != nil || !st.IsDir() {
		fmt.Fprintf(e.stderr, "warning: %s is not a folder yet\n", p)
	}
}
