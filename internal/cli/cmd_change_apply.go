package cli

import (
	"fmt"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/check"
	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/gitx"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
	"github.com/AliyuYahaya/Ajiya/internal/view"
)

// runChangeApprove applies every verdict of an open change in one write. Only a
// person runs it; it is not offered over MCP.
func runChangeApprove(e *env, args []string) error {
	fs := newFlags("change approve")
	by := fs.String("by", "", "the person who approves")
	newDec := fs.String("new-decision", "", "the decision that replaces the ones this change supersedes")
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	if *by == "" {
		return usageErr("--by is required: approving changes many tickets at once and is a person's decision")
	}
	if err := plan.CheckStatusText("--by", *by); err != nil {
		return usageErr("%v", err)
	}
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	c, err := pr.openChange(pos[0])
	if err != nil {
		return err
	}
	if probs := c.CandidateProblems(); len(probs) > 0 {
		return refused("%s cannot be approved: %s; assess every candidate with 'ajiya change assess'", c.ID, strings.Join(probs, "; "))
	}
	// A clean ajiya/ makes approving one diff on top of the committed plan
	// and the snapshot, and means the person approves what was committed.
	dirty, err := gitx.DirtyFiles(pr.cfg.Root, plan.Dir)
	if err != nil {
		return refused("%v", err)
	}
	if len(dirty) > 0 {
		if len(dirty) > 8 {
			dirty = append(dirty[:8], fmt.Sprintf("and %d more", len(dirty)-8))
		}
		return refused("%s/ has uncommitted changes: %s; commit the plan (with the verdicts) first so approving is one clean diff", plan.Dir, strings.Join(dirty, ", "))
	}

	baseline := map[string]bool{}
	for _, f := range check.Run(pr.cfg, pr.plan) {
		baseline[f.Code+f.Message] = true
	}
	before := map[string]string{}
	for _, ph := range pr.plan.Phases {
		before[ph.Slug] = string(ph.Format())
	}
	var ms []plan.Milestone
	for _, m := range pr.cfg.MilestoneList() {
		ms = append(ms, plan.Milestone{Name: m.Name, Targets: m.Targets})
	}
	opts := plan.ApplyOptions{Prefix: pr.cfg.Project.Prefix, NewDecision: *newDec, By: *by, Date: e.today(), Milestones: ms}
	retargets, err := pr.plan.ApplyChange(c, opts)
	if err != nil {
		if ae, ok := err.(*plan.ApplyError); ok {
			return refused("%s cannot be approved:\n  - %s", c.ID, strings.Join(ae.Problems, "\n  - "))
		}
		return refused("%v", err)
	}

	// Which files change. Phases left as they were are not written, so an old
	// six-column file stays as it is unless a ticket in it changed.
	var touched []*plan.Phase
	for _, ph := range pr.plan.Phases {
		if before[ph.Slug] != string(ph.Format()) {
			ph.SixColumn = false
			touched = append(touched, ph)
		}
	}

	// The ajiya.toml text with the milestones retargeted, computed in memory.
	newCfg, cfgText, err := retargetConfig(pr, retargets)
	if err != nil {
		return refused("%s cannot be approved: %v", c.ID, err)
	}

	// Validate the whole result before writing anything: check the plan as it
	// will be on disk. A copy holds the new file contents; the real plan keeps
	// the old ones so saving sees what changed.
	v := *pr.plan
	v.Raw = map[string][]byte{}
	for k, b := range pr.plan.Raw {
		v.Raw[k] = b
	}
	for _, ph := range touched {
		v.Raw[ph.Slug] = ph.Format()
	}
	v.RawDecisions = map[string][]byte{}
	for k, b := range pr.plan.RawDecisions {
		v.RawDecisions[k] = b
	}
	for _, id := range c.Decisions {
		v.RawDecisions[id] = pr.plan.Decision(id).Format()
	}
	v.RawChanges = map[string][]byte{}
	for k, b := range pr.plan.RawChanges {
		v.RawChanges[k] = b
	}
	v.RawChanges[c.ID] = c.Format()
	var bad []string
	for _, f := range check.Run(newCfg, &v) {
		if f.Level == check.Error && !baseline[f.Code+f.Message] {
			bad = append(bad, f.String())
		}
	}
	if len(bad) > 0 {
		return refused("approving %s would leave the plan with errors, so nothing was written:\n  - %s", c.ID, strings.Join(bad, "\n  - "))
	}

	for _, ph := range touched {
		if err := pr.plan.Save(ph); err != nil {
			return err
		}
	}
	for _, id := range c.Decisions {
		if err := pr.plan.SaveDecision(pr.plan.Decision(id)); err != nil {
			return err
		}
	}
	if cfgText != "" {
		if _, err := config.WriteText(pr.cfg.Root, cfgText); err != nil {
			return err
		}
	}
	if err := pr.plan.SaveChange(c); err != nil { // last: it is the record of what was done
		return err
	}
	if *asJSON {
		return writeJSON(e, view.ChangeOf(c))
	}
	fmt.Fprintf(e.stdout, "Approved %s by %s.\n%s\n", c.ID, *by, c.Results)
	if len(c.Decisions) > 0 && *newDec == "" {
		fmt.Fprintln(e.stdout, "No --new-decision was given, so the superseded decision(s) have no replacement.")
	}
	fmt.Fprintf(e.stdout, "Commit the plan change in ajiya/ with the trailer 'Ajiya: chore' (or the IDs it concerns).\n")
	return nil
}

// retargetConfig applies the retargets to the text of ajiya.toml, keeping the
// user's comments and layout, and returns the config it parses to and the new
// text. With no retargets it returns the current config and "".
func retargetConfig(pr *project, retargets []plan.Retarget) (*config.Config, string, error) {
	if len(retargets) == 0 {
		return pr.cfg, "", nil
	}
	text, err := config.ReadText(pr.cfg.Root)
	if err != nil {
		return nil, "", err
	}
	for _, r := range retargets {
		if pr.cfg.Launch.Target != "" { // the one-target [launch] form
			if len(r.Targets) != 1 {
				return nil, "", fmt.Errorf("the [launch] target would become %s, but it takes one; run 'ajiya launch set' for one of them first", strings.Join(r.Targets, ", "))
			}
			text = config.SetLaunchText(text, r.Targets[0])
			continue
		}
		if text, err = config.SetMilestoneTargetsText(text, r.Name, r.Targets); err != nil {
			return nil, "", err
		}
	}
	cfg, err := config.Parse([]byte(text))
	if err != nil {
		return nil, "", fmt.Errorf("ajiya.toml would not load: %v", err)
	}
	cfg.Root = pr.cfg.Root
	return cfg, text, nil
}

// runChangeReject closes an open change without touching the plan.
func runChangeReject(e *env, args []string) error {
	fs := newFlags("change reject")
	by := fs.String("by", "", "the person who rejects")
	reason := fs.String("reason", "", "why the change is rejected")
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	if *by == "" || *reason == "" {
		return usageErr("--by and --reason are both required: rejecting a change is a person's decision")
	}
	for what, v := range map[string]string{"--by": *by, "--reason": *reason} {
		if err := plan.CheckStatusText(what, v); err != nil {
			return usageErr("%v", err)
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
	c.Status = plan.ChangeStatus{State: plan.ChgRejected, Date: e.today(), By: *by, Reason: *reason}
	c.Results = "- Rejected: the plan was not changed"
	if err := pr.plan.SaveChange(c); err != nil {
		return err
	}
	if *asJSON {
		return writeJSON(e, view.ChangeOf(c))
	}
	fmt.Fprintf(e.stdout, "Rejected %s by %s: %s. The plan was not changed.\n", c.ID, *by, *reason)
	return nil
}
