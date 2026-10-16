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
	internal["tags"] = true                                        // until AJ-0056 writes the tags section of the guide; remove this line then
	internal["snapshot"], internal["snapshot list"] = true, true   // until AJ-0056 writes the change section of the guide; remove this line then
	for _, c := range commands {
		if internal[c.name] {
			continue
		}
		if !strings.Contains(s, "ajiya "+c.name) {
			t.Errorf("no guide mentions 'ajiya %s'", c.name)
		}
	}
}
