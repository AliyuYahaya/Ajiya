package check

import (
	"strings"
	"testing"
	"time"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// W015: a change open for more than 14 days; closed changes and fresh ones are fine.
func TestOpenChangeWarning(t *testing.T) {
	mk := func(id, date string, state plan.ChangeState) *plan.Change {
		return &plan.Change{ID: id, Status: plan.ChangeStatus{State: state, Date: date}}
	}
	p := &plan.Plan{Changes: []*plan.Change{
		mk("CHG-0001", "2026-10-01", plan.ChgOpen),     // 15 days old on the 16th
		mk("CHG-0002", "2026-10-02", plan.ChgOpen),     // 14 days: not yet
		mk("CHG-0003", "2026-09-01", plan.ChgApproved), // closed
	}}
	now := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	fs := openChangeWarnings(p, now)
	if len(fs) != 1 || fs[0].Code != "W015" || fs[0].Level != Warning || !strings.Contains(fs[0].Message, "CHG-0001 has been open for 15 days") ||
		fs[0].Location != "ajiya/changes/CHG-0001.md:1" {
		t.Errorf("findings = %v", fs)
	}
}
