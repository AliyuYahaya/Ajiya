package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

func runInit(e *env, args []string) error {
	fs := newFlags("init")
	name := fs.String("name", filepath.Base(e.dir), "project name")
	prefix := fs.String("prefix", "", "ticket ID prefix")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	if *prefix == "" {
		*prefix = config.SuggestPrefix(*name)
	}
	path := filepath.Join(e.dir, config.FileName)
	if _, err := os.Stat(path); err == nil {
		return refused("%s already exists in this directory", config.FileName)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	text := config.Template(*name, *prefix)
	if _, err := config.Parse([]byte(text)); err != nil {
		return usageErr("%v", err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(e.dir, plan.Dir), 0o755); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "Set up %s with ticket IDs like %s.\nWrote %s and %s/.\nNext: register apps in %s, then 'ajiya phase add <slug> \"<goal>\"'.\n",
		*name, plan.FormatID(*prefix, 1), config.FileName, plan.Dir, config.FileName)
	return nil
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
