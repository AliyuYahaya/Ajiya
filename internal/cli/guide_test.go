package cli

import (
	"strings"
	"testing"

	"github.com/AliyuYahaya/Ajiya/internal/kit"
)

// Every command must appear in the agent guides, so a command added without
// updating them fails here.
func TestGuidesCoverCommands(t *testing.T) {
	s := kit.Source("guide/setup.md") + kit.Source("guide/daily.md")
	internal := map[string]bool{"hook run": true, "version": true} // not for agents
	for _, c := range commands {
		if internal[c.name] {
			continue
		}
		if !strings.Contains(s, "ajiya "+c.name) {
			t.Errorf("no guide mentions 'ajiya %s'", c.name)
		}
	}
}
