package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every command must appear in the agent guides, so a command added without
// updating them fails here. The guides are in the repository's .ajiya/guide.
func TestGuidesCoverCommands(t *testing.T) {
	var text strings.Builder
	for _, name := range []string{"setup.md", "daily.md"} {
		b, err := os.ReadFile(filepath.Join("..", "..", ".ajiya", "guide", name))
		if err != nil {
			t.Fatal(err)
		}
		text.Write(b)
	}
	s := text.String()
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
