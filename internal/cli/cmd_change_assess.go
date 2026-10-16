package cli

import (
	"fmt"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
	"github.com/AliyuYahaya/Ajiya/internal/view"
)

// runChangeAssess records the agent's verdict on one candidate of an open change.
func runChangeAssess(e *env, args []string) error {
	fs := newFlags("change assess")
	verdict := fs.String("verdict", "", "unaffected, modify, replace, invalidate or investigate")
	reason := fs.String("reason", "", "why this verdict")
	title := fs.String("title", "", "title of the new ticket (replace, modify of a done ticket) or the edited title (modify of an open ticket)")
	doneWhen := fs.String("done-when", "", "done-when of the new ticket, or the edited done-when")
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parse(fs, args, 2)
	if err != nil {
		return err
	}
	if *verdict == "" {
		return usageErr("--verdict is required: one of %s", verdictNames())
	}
	v, err := plan.ParseVerdict(*verdict)
	if err != nil {
		return usageErr("--verdict: %v", err)
	}
	if *reason == "" {
		return usageErr("--reason is required: say why the ticket is %s", v)
	}
	if isSet(fs, "title") && *title == "" {
		return usageErr("--title cannot be empty")
	}
	if isSet(fs, "done-when") && *doneWhen == "" {
		return usageErr("--done-when cannot be empty")
	}
	if err := checkTexts("--reason", *reason, "--title", *title, "--done-when", *doneWhen); err != nil {
		return err
	}
	given := *title != "" || *doneWhen != ""
	switch v {
	case plan.VerdictUnaffected, plan.VerdictInvalidate, plan.VerdictInvestigate:
		if given {
			return usageErr("--title and --done-when are for replace and modify; a %s verdict creates or edits no ticket", v)
		}
	case plan.VerdictReplace:
		if *title == "" || *doneWhen == "" {
			return usageErr("replace needs --title and --done-when for the new ticket")
		}
	}
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	c, err := pr.openChange(pos[0])
	if err != nil {
		return err
	}
	k := c.Candidate(pos[1])
	if k == nil {
		return refused("%s is not a candidate of %s; run 'ajiya change impact %s' to find candidates", pos[1], c.ID, c.ID)
	}
	t := pr.plan.Ticket(k.ID)
	if t == nil {
		return refused("%s is no longer in the plan", k.ID)
	}
	if t.Status.Excluded() && v != plan.VerdictUnaffected && v != plan.VerdictInvestigate {
		return refused("%s is already %s; only unaffected or investigate make sense for it", t.ID, strings.ToLower(t.Status.String()))
	}
	if v == plan.VerdictModify {
		switch {
		case t.Status.State == plan.Done && (*title == "" || *doneWhen == ""):
			return usageErr("%s is done, so modify adds a follow-up ticket: give --title and --done-when for it", t.ID)
		case !t.Status.Closed() && !given:
			return usageErr("%s is open, so modify edits it in place: give --title, --done-when or both (tags are not edited here; use 'ajiya ticket edit')", t.ID)
		}
	}
	if err := c.SetVerdict(k.ID, v, *reason, *title, *doneWhen); err != nil {
		return refused("%v", err)
	}
	if err := pr.plan.SaveChange(c); err != nil {
		return err
	}
	if *asJSON {
		return writeJSON(e, view.ChangeOf(c))
	}
	fmt.Fprintf(e.stdout, "%s: %s is %s (%s)\n", c.ID, k.ID, v, *reason)
	left := 0
	for _, x := range c.Candidates {
		if x.Verdict == "" || x.Verdict == plan.VerdictInvestigate {
			left++
		}
	}
	if left > 0 {
		fmt.Fprintf(e.stdout, "%d candidate(s) still need a verdict other than investigate.\n", left)
	} else {
		fmt.Fprintf(e.stdout, "Every candidate has a verdict. Show the person 'ajiya change show %s'; only they run approve or reject.\n", c.ID)
	}
	return nil
}

func verdictNames() string {
	var n []string
	for _, v := range plan.Verdicts {
		n = append(n, string(v))
	}
	return strings.Join(n, ", ")
}
