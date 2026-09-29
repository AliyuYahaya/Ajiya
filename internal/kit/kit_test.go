package kit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWithBlock(t *testing.T) {
	b := Block()
	tests := []struct {
		name, in, want, err string
	}{
		{"empty file", "", b + "\n", ""},
		{"text without newline", "# Notes", "# Notes\n\n" + b + "\n", ""},
		{"text with newline", "# Notes\n", "# Notes\n\n" + b + "\n", ""},
		{"text with blank line", "# Notes\n\n", "# Notes\n\n" + b + "\n", ""},
		{"old block replaced, text kept", "# Mine\n\n" + Start + "\nold\n" + End + "\n\nMore of mine.\n", "# Mine\n\n" + b + "\n\nMore of mine.\n", ""},
		{"current block unchanged", "Top\n" + b + "\nBottom", "Top\n" + b + "\nBottom", ""},
		{"start only", "x " + Start, "", "found 1"},
		{"two blocks", Start + End + Start + End, "", "found 2"},
		{"markers reversed", End + " " + Start, "", "comes before"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WithBlock(tt.in)
			if tt.err != "" {
				if err == nil || !strings.Contains(err.Error(), tt.err) {
					t.Fatalf("err = %v, want one containing %q", err, tt.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

func actions(rs []Result) map[string]string {
	m := map[string]string{}
	for _, r := range rs {
		m[r.Path] = r.Action
	}
	return m
}

func TestInstall(t *testing.T) {
	dir := t.TempDir()
	mine := "# House rules\n\nUse tabs.\n"
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(mine), 0o644); err != nil {
		t.Fatal(err)
	}

	// First run: every owned file created, AGENTS.md created, CLAUDE.md gains
	// the block after the user's text.
	rs, err := Install(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := actions(rs)
	for _, o := range Owned {
		if got[o.Path] != "created" {
			t.Errorf("%s: %s, want created", o.Path, got[o.Path])
		}
		b, err := os.ReadFile(filepath.Join(dir, o.Path))
		if err != nil || string(b) != Source(o.Source) {
			t.Errorf("%s does not hold the kit's copy (err %v)", o.Path, err)
		}
	}
	if got["AGENTS.md"] != "created" || got["CLAUDE.md"] != "updated" {
		t.Errorf("AGENTS.md %s, CLAUDE.md %s; want created, updated", got["AGENTS.md"], got["CLAUDE.md"])
	}
	claude, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if string(claude) != mine+"\n"+Block()+"\n" {
		t.Errorf("CLAUDE.md = %q", claude)
	}

	// Second run: nothing to do.
	rs, err = Install(dir)
	if err != nil {
		t.Fatal(err)
	}
	for p, a := range actions(rs) {
		if a != "unchanged" {
			t.Errorf("second run: %s %s, want unchanged", p, a)
		}
	}

	// The user edits around the block and changes a guide; the next run
	// restores the guide and the block and keeps the user's edits.
	edited := "Before.\n" + strings.Replace(string(claude), "Use tabs.", "Use spaces.", 1) + "After.\n"
	edited = strings.Replace(edited, "ajiya next", "something else", 1)
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(edited), 0o644)
	os.WriteFile(filepath.Join(dir, ".ajiya/guide/daily.md"), []byte("changed"), 0o644)
	rs, err = Install(dir)
	if err != nil {
		t.Fatal(err)
	}
	got = actions(rs)
	if got["CLAUDE.md"] != "updated" || got[".ajiya/guide/daily.md"] != "updated" {
		t.Errorf("third run: %v", got)
	}
	claude, _ = os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	want := "Before.\n# House rules\n\nUse spaces.\n\n" + Block() + "\nAfter.\n"
	if string(claude) != want {
		t.Errorf("CLAUDE.md = %q\nwant %q", claude, want)
	}
}

// Broken markers in one file stop the whole install before anything is
// written.
func TestInstallRefusesBrokenMarkers(t *testing.T) {
	dir := t.TempDir()
	broken := "Mine\n" + Start + "\nno end\n"
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(broken), 0o644)
	_, err := Install(dir)
	if err == nil || !strings.Contains(err.Error(), "CLAUDE.md") {
		t.Fatalf("err = %v, want a refusal naming CLAUDE.md", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("files were written despite the refusal: %v", entries)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md")); string(b) != broken {
		t.Error("CLAUDE.md was changed")
	}
}

// The skill follows the Agent Skills specification and carries the rules and
// the pointers to both guides.
func TestSkill(t *testing.T) {
	s := Source("SKILL.md")
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
	if fields["name"] != "ajiya" || filepath.Base(filepath.Dir(Owned[2].Path)) != "ajiya" {
		t.Errorf("name = %q, want ajiya, the name of the folder it installs into", fields["name"])
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
	for _, o := range Owned[:2] {
		if !strings.Contains(body, o.Path) {
			t.Errorf("SKILL.md does not point to %s", o.Path)
		}
	}
}

// The block for AGENTS.md and CLAUDE.md sits between the markers, lists the
// loop and points to both guides.
func TestBlock(t *testing.T) {
	b := Block()
	if !strings.HasPrefix(b, Start) || !strings.HasSuffix(b, End) || strings.Count(b, Start) != 1 || strings.Count(b, End) != 1 {
		t.Fatalf("block must run from %s to %s", Start, End)
	}
	for _, want := range []string{"ajiya next", "Ajiya: <ID>", "ajiya ticket done <ID> --test", Owned[0].Path, Owned[1].Path} {
		if !strings.Contains(b, want) {
			t.Errorf("block lacks %q", want)
		}
	}
}

// This repository runs the kit it ships: 'ajiya init' here changes nothing.
// When the kit changes, run 'ajiya init' in the repository and commit the result.
func TestRepositoryKitIsCurrent(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, o := range Owned {
		b, err := os.ReadFile(filepath.Join(root, o.Path))
		if err != nil || string(b) != Source(o.Source) {
			t.Errorf("%s differs from the kit (err %v); run 'ajiya init' and commit", o.Path, err)
		}
	}
	for _, name := range Blocks {
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if got, err := WithBlock(string(b)); err != nil || got != string(b) {
			t.Errorf("%s's block differs from the kit (err %v); run 'ajiya init' and commit", name, err)
		}
	}
}
