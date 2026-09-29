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

func readRepoFile(t *testing.T, parts ...string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(append([]string{"..", ".."}, parts...)...))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// The skill follows the Agent Skills specification and carries the rules and
// the pointers to both guides.
func TestSkillFile(t *testing.T) {
	s := readRepoFile(t, ".claude", "skills", "ajiya", "SKILL.md")
	if !strings.HasPrefix(s, "---\n") {
		t.Fatal("SKILL.md must start with the --- frontmatter line")
	}
	end := strings.Index(s[4:], "\n---\n")
	if end < 0 {
		t.Fatal("SKILL.md frontmatter is not closed")
	}
	fields := map[string]string{}
	for _, line := range strings.Split(s[4:4+end], "\n") {
		if k, v, ok := strings.Cut(line, ": "); ok {
			fields[k] = v
		}
	}
	if fields["name"] != "ajiya" {
		t.Errorf("name = %q, want the folder name ajiya", fields["name"])
	}
	if d := fields["description"]; d == "" || len(d) > 1024 || !strings.Contains(d, "Use ") {
		t.Errorf("description must be 1 to 1024 characters and say when to use the skill: %q", d)
	}
	body := s[4+end:]
	for i := 1; i <= 5; i++ {
		if !strings.Contains(body, "\n"+string(rune('0'+i))+". ") {
			t.Errorf("SKILL.md lacks rule %d", i)
		}
	}
	for _, g := range []string{".ajiya/guide/daily.md", ".ajiya/guide/setup.md"} {
		if !strings.Contains(body, g) {
			t.Errorf("SKILL.md does not point to %s", g)
		}
	}
}

// AGENTS.md (Codex) and CLAUDE.md carry the same marked block, which points to
// the guides.
func TestAgentBlocks(t *testing.T) {
	const start, end = "<!-- ajiya:start -->", "<!-- ajiya:end -->"
	var blocks []string
	for _, name := range []string{"AGENTS.md", "CLAUDE.md"} {
		s := readRepoFile(t, name)
		i, j := strings.Index(s, start), strings.Index(s, end)
		if i < 0 || j < i || strings.Count(s, start) != 1 || strings.Count(s, end) != 1 {
			t.Fatalf("%s needs exactly one %s ... %s block", name, start, end)
		}
		b := s[i : j+len(end)]
		for _, want := range []string{"ajiya next", "Ajiya: <ID>", "ajiya ticket done <ID> --test", ".ajiya/guide/daily.md", ".ajiya/guide/setup.md"} {
			if !strings.Contains(b, want) {
				t.Errorf("%s block lacks %q", name, want)
			}
		}
		blocks = append(blocks, b)
	}
	if blocks[0] != blocks[1] {
		t.Error("the AGENTS.md and CLAUDE.md blocks differ")
	}
}
