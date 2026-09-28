package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/AliyuYahaya/Ajiya/internal/importer"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// Imported items land in this phase until a person moves them.
const (
	githubInboxSlug = "inbox"
	githubInboxGoal = "Imported items waiting to be organised"
)

// runImportGitHub imports a repository's issues into the inbox phase: open
// issues as pending tickets, closed ones as done with the issue number as
// evidence, each keeping its URL. Issues already imported (their URL is a
// ticket's link) are skipped, so running it again only adds new issues.
//
// This is the only command that reaches the network, and it does so only
// through the GitHub CLI (gh), with the user's own login, when asked to.
func runImportGitHub(e *env, args []string) error {
	fs := newFlags("import github")
	repo := fs.String("repo", "", "the repository as owner/name; default: the one gh finds from the current directory")
	app := fs.String("app", "infra", "the app the tickets belong to")
	limit := fs.Int("limit", 1000, "the most issues to ask gh for")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	if *limit <= 0 {
		return usageErr("--limit must be a positive number")
	}
	pr, err := loadForChange(e)
	if err != nil {
		return err
	}
	if err := pr.checkApp(*app); err != nil {
		return err
	}
	gh, err := exec.LookPath("gh")
	if err != nil {
		return refused("gh is not installed; install GitHub CLI (https://cli.github.com) and run 'gh auth login'")
	}

	ghArgs := []string{"issue", "list", "--state", "all", "--limit", fmt.Sprint(*limit), "--json", importer.GitHubFields}
	if *repo != "" {
		ghArgs = append(ghArgs, "--repo", *repo)
	}
	cmd := exec.Command(gh, ghArgs...)
	cmd.Dir = pr.cfg.Root
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		msg := ghFirstLine(stderr.String())
		var exit *exec.ExitError
		if msg == "" && errors.As(err, &exit) {
			msg = fmt.Sprintf("exit %d", exit.ExitCode())
		} else if msg == "" {
			msg = err.Error()
		}
		return refused("gh issue list failed: %s", msg)
	}
	issues, err := importer.GitHubDecode(stdout.Bytes())
	if err != nil {
		return refused("%v", err)
	}

	linked := map[string]bool{}
	for _, t := range pr.plan.Tickets() {
		if t.Status.Link != "" {
			linked[t.Status.Link] = true
		}
	}
	var ph *plan.Phase
	added, closed, skipped := 0, 0, 0
	for _, is := range issues {
		if linked[is.URL] {
			skipped++
			continue
		}
		if err := importer.GitHubCheck(is); err != nil {
			fmt.Fprintf(e.stderr, "warning: skipped issue #%d: %v\n", is.Number, err)
			continue
		}
		if ph == nil {
			ph = pr.ensurePhase(githubInboxSlug, githubInboxGoal)
		}
		t := pr.newTicket(ph, *app, is.Title, "-", nil, importer.GitHubStatus(is))
		linked[is.URL] = true
		added++
		if t.Status.State == plan.Done {
			closed++
		}
	}
	if ph != nil {
		if err := pr.plan.Save(ph); err != nil {
			return err
		}
	}
	fmt.Fprintf(e.stdout, "Imported %d issue(s) into %s (%d closed); skipped %d already imported.\n", added, githubInboxSlug, closed, skipped)
	return nil
}

// ghFirstLine returns the first non-blank line of s, trimmed.
func ghFirstLine(s string) string {
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			return l
		}
	}
	return ""
}
