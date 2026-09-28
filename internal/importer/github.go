package importer

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// GitHubFields are the fields asked of 'gh issue list --json'.
const GitHubFields = "number,title,state,url"

// GitHubIssue is one issue as printed by 'gh issue list --json number,title,state,url'.
type GitHubIssue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"` // OPEN or CLOSED
	URL    string `json:"url"`
}

// GitHubDecode reads the JSON array printed by gh and returns the issues in
// ascending number order, so imported IDs do not depend on gh's order.
func GitHubDecode(data []byte) ([]GitHubIssue, error) {
	var issues []GitHubIssue
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, fmt.Errorf("could not read the issue list from gh: %v", err)
	}
	for _, is := range issues {
		if is.Number <= 0 {
			return nil, errors.New("gh listed an issue without a number")
		}
		if is.URL == "" {
			return nil, fmt.Errorf("gh listed issue #%d without a URL", is.Number)
		}
		if is.State != "OPEN" && is.State != "CLOSED" {
			return nil, fmt.Errorf("gh listed issue #%d with state %q; expected OPEN or CLOSED", is.Number, is.State)
		}
	}
	sort.SliceStable(issues, func(i, j int) bool { return issues[i].Number < issues[j].Number })
	return issues, nil
}

// GitHubStatus returns the status an imported issue starts with: pending if
// open, done with the issue number as evidence if closed. Both keep the URL.
func GitHubStatus(is GitHubIssue) plan.Status {
	if is.State == "CLOSED" {
		return plan.Status{State: plan.Done, Issue: fmt.Sprint(is.Number), Link: is.URL}
	}
	return plan.Status{State: plan.Pending, Link: is.URL}
}

// GitHubCheck reports why an issue cannot become a ticket as it is, or nil.
// Imports never reword titles, so the caller skips such issues instead.
func GitHubCheck(is GitHubIssue) error {
	if is.Title == "" {
		return errors.New("the title is empty")
	}
	if err := plan.CheckText("the title", is.Title); err != nil {
		return err
	}
	return plan.CheckStatusText("the URL", is.URL)
}
