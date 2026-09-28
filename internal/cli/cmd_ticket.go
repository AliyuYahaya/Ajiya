package cli

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/gitx"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

func runTicketAdd(e *env, args []string) error {
	fs := newFlags("ticket add")
	phase := fs.String("phase", "", "phase slug")
	app := fs.String("app", "", "app name")
	doneWhen := fs.String("done-when", "", "the check that proves the ticket is done")
	depends := fs.String("depends", "", "IDs this ticket depends on, comma separated")
	human := fs.String("human", "", "why this ticket needs a person")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	title := pos[0]
	for _, f := range []struct{ name, v string }{{"--phase", *phase}, {"--app", *app}, {"--done-when", *doneWhen}} {
		if f.v == "" {
			return usageErr("%s is required", f.name)
		}
	}
	if title == "" {
		return usageErr("the title is empty")
	}
	if err := checkTexts("title", title, "--done-when", *doneWhen); err != nil {
		return err
	}
	if err := plan.CheckStatusText("--human", *human); err != nil {
		return usageErr("%v", err)
	}

	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	ph, err := pr.phase(*phase)
	if err != nil {
		return err
	}
	if err := pr.checkApp(*app); err != nil {
		return err
	}
	deps := plan.ParseIDList(*depends)
	if err := pr.checkDepends("", deps); err != nil {
		return err
	}
	t := &plan.Ticket{
		ID:       pr.plan.NextID(pr.cfg.Project.Prefix),
		App:      *app,
		Title:    title,
		DoneWhen: *doneWhen,
		Depends:  deps,
		Status:   plan.Status{State: plan.Pending, Human: *human},
		Phase:    ph.Slug,
	}
	ph.Tickets = append(ph.Tickets, t)
	if err := pr.plan.Save(ph); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "Added %s to %s: %s\n", t.ID, ph.Slug, t.Title)
	return nil
}

// checkTexts checks pairs of (name, value) in order.
func checkTexts(pairs ...string) error {
	for i := 0; i+1 < len(pairs); i += 2 {
		if err := plan.CheckText(pairs[i], pairs[i+1]); err != nil {
			return usageErr("%v", err)
		}
	}
	return nil
}

func runTicketStart(e *env, args []string) error {
	fs := newFlags("ticket start")
	note := fs.String("note", "", "what is left to do")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	if err := plan.CheckStatusText("--note", *note); err != nil {
		return usageErr("%v", err)
	}
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	t, err := pr.ticket(pos[0])
	if err != nil {
		return err
	}
	if t.Status.Closed() {
		return refused("%s is already closed: %s", t.ID, t.Status)
	}
	s := plan.Status{State: plan.InProgress, Human: t.Status.Human, Note: *note}
	if *note == "" && t.Status.State == plan.InProgress {
		s.Note = t.Status.Note
	}
	t.Status = s
	if err := pr.saveTicket(t); err != nil {
		return err
	}
	for _, d := range waitingOn(pr, t) {
		fmt.Fprintf(e.stderr, "warning: %s depends on %s, which is not done (%s)\n", t.ID, d.ID, d.Status)
	}
	fmt.Fprintf(e.stdout, "%s %s\n", t.ID, t.Status)
	return nil
}

// waitingOn returns the dependencies of t that are not closed.
func waitingOn(pr *project, t *plan.Ticket) []*plan.Ticket {
	var out []*plan.Ticket
	for _, id := range t.Depends {
		if d := pr.plan.Ticket(id); d != nil && !d.Status.Closed() {
			out = append(out, d)
		}
	}
	return out
}

func runTicketDone(e *env, args []string) error {
	fs := newFlags("ticket done")
	note := fs.String("note", "", "a note to keep with the evidence")
	by := fs.String("by", "", "the person who did it, for tickets that need a human")
	test := fs.Bool("test", false, "run the [test] command and record the result; a failure refuses")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	if err := plan.CheckStatusText("--note", *note); err != nil {
		return usageErr("%v", err)
	}
	if err := plan.CheckStatusText("--by", *by); err != nil {
		return usageErr("%v", err)
	}
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	t, err := pr.ticket(pos[0])
	if err != nil {
		return err
	}
	if t.Status.Closed() {
		return refused("%s is already closed: %s", t.ID, t.Status)
	}

	s := plan.Status{State: plan.Done, Note: *note}
	if *by != "" {
		if t.Status.Human == "" {
			return refused("--by is only for tickets marked 'Needs a human'; for %s, commit with the trailer 'Ajiya: %s' and run 'ajiya ticket done %s'", t.ID, t.ID, t.ID)
		}
		if *note == "" {
			return usageErr("--by needs --note \"<what was done>\"")
		}
		s.By, s.Date = *by, today()
	} else {
		commits, err := gitx.Referencing(pr.cfg.Root, t.ID)
		if err != nil {
			return refused("%v", err)
		}
		if len(commits) == 0 {
			hint := ""
			if t.Status.Human != "" {
				hint = `, or close it with --by "<name>" --note "<what was done>"`
			}
			return refused("no commit references %s; commit the work with the trailer 'Ajiya: %s' first%s", t.ID, t.ID, hint)
		}
		s.Commit, s.Date = commits[0].Short(), commits[0].Date
	}
	if *test {
		if err := runTests(e, pr, t.ID); err != nil {
			return err
		}
		s.Tests = true
	}
	for _, d := range waitingOn(pr, t) {
		fmt.Fprintf(e.stderr, "warning: %s depends on %s, which is not done (%s)\n", t.ID, d.ID, d.Status)
	}
	t.Status = s
	if err := pr.saveTicket(t); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "%s %s\n", t.ID, t.Status)
	return nil
}

func runTicketShow(e *env, args []string) error {
	fs := newFlags("ticket show")
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	t, err := pr.ticket(pos[0])
	if err != nil {
		return err
	}
	ph := pr.plan.Phase(t.Phase)
	dependants := plan.NewGraph(pr.plan.Tickets()).Dependants(t.ID)
	commits, err := gitx.Referencing(pr.cfg.Root, t.ID)
	if err != nil && err != gitx.ErrNotRepo {
		return err
	}
	if *asJSON {
		type link struct {
			ID    string `json:"id"`
			Title string `json:"title,omitempty"`
			State string `json:"state"` // missing when the ID does not exist
		}
		type commit struct {
			Hash    string `json:"hash"`
			Date    string `json:"date"`
			Subject string `json:"subject"`
		}
		links := func(ids []string) []link {
			out := []link{}
			for _, id := range ids {
				l := link{ID: id, State: "missing"}
				if d := pr.plan.Ticket(id); d != nil {
					l.Title, l.State = d.Title, stateNames[d.Status.State]
				}
				out = append(out, l)
			}
			return out
		}
		cs := []commit{}
		for _, c := range commits {
			cs = append(cs, commit{c.Hash, c.Date, c.Subject})
		}
		return writeJSON(e, struct {
			jsonTicket
			PhaseTitle   string   `json:"phase_title"`
			Dependencies []link   `json:"dependencies"`
			Dependants   []link   `json:"dependants"`
			Commits      []commit `json:"commits"`
		}{toJSON(t), ph.Title, links(t.Depends), links(dependants), cs})
	}
	w := e.stdout
	fmt.Fprintf(w, "%s  %s\n", t.ID, t.Title)
	fmt.Fprintf(w, "Phase:      %s (%s)\n", ph.Slug, ph.Title)
	fmt.Fprintf(w, "App:        %s\n", t.App)
	fmt.Fprintf(w, "Done when:  %s\n", t.DoneWhen)
	fmt.Fprintf(w, "Status:     %s\n", t.Status)
	list := func(label string, ids []string) {
		if len(ids) == 0 {
			fmt.Fprintf(w, "%-11s -\n", label)
			return
		}
		for i, id := range ids {
			if i > 0 {
				label = ""
			}
			status := "(does not exist)"
			if d := pr.plan.Ticket(id); d != nil {
				status = d.Status.Mark() + " " + d.Title
			}
			fmt.Fprintf(w, "%-11s %s %s\n", label, id, status)
		}
	}
	list("Depends:", t.Depends)
	list("Needed by:", dependants)
	if len(commits) == 0 {
		fmt.Fprintf(w, "%-11s -\n", "Commits:")
	}
	for i, c := range commits {
		label := "Commits:"
		if i > 0 {
			label = ""
		}
		fmt.Fprintf(w, "%-11s %s %s %s\n", label, c.Short(), c.Date, c.Subject)
	}
	return nil
}

func runTicketEdit(e *env, args []string) error {
	fs := newFlags("ticket edit")
	title := fs.String("title", "", "new title")
	doneWhen := fs.String("done-when", "", "new done-when check")
	depends := fs.String("depends", "", `new dependency list; "-" for none`)
	app := fs.String("app", "", "new app")
	phase := fs.String("phase", "", "move to this phase")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	if fs.NFlag() == 0 {
		return usageErr("nothing to change; give at least one of --title, --done-when, --depends, --app, --phase")
	}
	for _, f := range []struct {
		name string
		v    string
	}{{"title", *title}, {"done-when", *doneWhen}, {"app", *app}, {"phase", *phase}} {
		if isSet(fs, f.name) && f.v == "" {
			return usageErr("--%s cannot be empty", f.name)
		}
	}
	if err := checkTexts("--title", *title, "--done-when", *doneWhen); err != nil {
		return err
	}

	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	t, err := pr.ticket(pos[0])
	if err != nil {
		return err
	}
	if isSet(fs, "app") {
		if err := pr.checkApp(*app); err != nil {
			return err
		}
		t.App = *app
	}
	if isSet(fs, "depends") {
		deps := plan.ParseIDList(*depends)
		if err := pr.checkDepends(t.ID, deps); err != nil {
			return err
		}
		t.Depends = deps
	}
	if isSet(fs, "title") {
		t.Title = *title
	}
	if isSet(fs, "done-when") {
		t.DoneWhen = *doneWhen
	}
	from := pr.plan.Phase(t.Phase)
	if isSet(fs, "phase") && *phase != t.Phase {
		to, err := pr.phase(*phase)
		if err != nil {
			return err
		}
		pr.plan.Move(t, to)
		if err := pr.plan.Save(to); err != nil {
			return err
		}
	}
	if err := pr.plan.Save(from); err != nil {
		return err
	}
	deps := "-"
	if len(t.Depends) > 0 {
		deps = strings.Join(t.Depends, ", ")
	}
	fmt.Fprintf(e.stdout, "%s  %s\n  phase %s · app %s · depends %s\n  done when: %s\n", t.ID, t.Title, t.Phase, t.App, deps, t.DoneWhen)
	return nil
}

func runTicketBlock(e *env, args []string) error {
	pos, err := parse(newFlags("ticket block"), args, 2)
	if err != nil {
		return err
	}
	reason := pos[1]
	if reason == "" {
		return usageErr("say what the ticket is waiting on")
	}
	if err := plan.CheckStatusText("reason", reason); err != nil {
		return usageErr("%v", err)
	}
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	t, err := pr.ticket(pos[0])
	if err != nil {
		return err
	}
	if t.Status.Closed() {
		return refused("%s is already closed: %s", t.ID, t.Status)
	}
	t.Status = plan.Status{State: plan.Pending, Human: t.Status.Human, Blocked: reason}
	if err := pr.saveTicket(t); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "%s %s\n", t.ID, t.Status)
	return nil
}

// runTicketDrop closes a ticket without doing it. Only a person decides that.
func runTicketDrop(e *env, args []string) error {
	fs := newFlags("ticket drop")
	reason := fs.String("reason", "", "why the ticket is not needed")
	by := fs.String("by", "", "the person who decided")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	if *reason == "" || *by == "" {
		return usageErr("--reason and --by are both required: dropping work is a person's decision")
	}
	for what, v := range map[string]string{"--reason": *reason, "--by": *by} {
		if err := plan.CheckStatusText(what, v); err != nil {
			return usageErr("%v", err)
		}
	}
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	t, err := pr.ticket(pos[0])
	if err != nil {
		return err
	}
	if t.Status.Closed() {
		return refused("%s is already closed: %s", t.ID, t.Status)
	}
	t.Status = plan.Status{State: plan.Dropped, Reason: *reason, By: *by}
	if err := pr.saveTicket(t); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "%s %s\n", t.ID, t.Status)
	return nil
}

// runTests runs the configured test command in the project root, with its
// output on stderr so stdout stays for ajiya's own result.
func runTests(e *env, pr *project, id string) error {
	command := pr.cfg.Test.Command
	if command == "" {
		return usageErr("no [test] command in %s; add one, or leave out --test", config.FileName)
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Dir = pr.cfg.Root
	cmd.Stdout, cmd.Stderr = e.stderr, e.stderr
	fmt.Fprintf(e.stderr, "ajiya: running %s\n", command)
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return refused("tests failed (exit %d); %s stays open until they pass", exit.ExitCode(), id)
		}
		return refused("could not run the test command: %v", err)
	}
	return nil
}
