package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AliyuYahaya/Ajiya/internal/importer"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

const (
	inboxPhase = "inbox"
	inboxGoal  = "Imported items waiting to be organised"
)

// runImportTodo adds the items of a markdown to-do list to the inbox phase,
// word for word. Each ticket links back to <file>:<line>, so importing the
// same file again skips the items already imported.
func runImportTodo(e *env, args []string) error {
	fs := newFlags("import todo")
	app := fs.String("app", "infra", "app name")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	file := pos[0]
	if file == "" {
		return usageErr("the file name is empty")
	}
	name := filepath.ToSlash(file)
	if err := plan.CheckStatusText("the file name", name); err != nil {
		return usageErr("%v", err)
	}
	path := file
	if !filepath.IsAbs(path) {
		path = filepath.Join(e.dir, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return usageErr("cannot read %s: %v", file, err)
	}

	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	if err := pr.checkApp(*app); err != nil {
		return err
	}
	linked := map[string]bool{}
	for _, t := range pr.plan.Tickets() {
		if t.Status.Link != "" {
			linked[t.Status.Link] = true
		}
	}

	var ph *plan.Phase
	imported, before, skipped := 0, 0, 0
	for _, it := range importer.ParseTodo(string(data)) {
		link := fmt.Sprintf("%s:%d", name, it.Line)
		if linked[link] {
			skipped++
			continue
		}
		if err := plan.CheckText("the title", it.Text); err != nil {
			fmt.Fprintf(e.stderr, "warning: %s: skipped: %v\n", link, err)
			continue
		}
		st := plan.Status{State: plan.Pending, Link: link}
		if it.Ticked {
			st = plan.Status{State: plan.Done, Before: true, Link: link}
			before++
		}
		if ph == nil {
			ph = pr.ensurePhase(inboxPhase, inboxGoal)
		}
		pr.newTicket(ph, *app, it.Text, "-", nil, st)
		linked[link] = true
		imported++
	}
	if ph != nil {
		if err := pr.plan.Save(ph); err != nil {
			return err
		}
	}
	fmt.Fprintf(e.stdout, "Imported %d item(s) into %s (%d done before Ajiya); skipped %d already imported.\n",
		imported, inboxPhase, before, skipped)
	return nil
}
