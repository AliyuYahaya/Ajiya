package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitSuggestsPerAppTestCommand(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	files := map[string]string{
		"package.json":          `{"name":"root","private":true}`,
		"pnpm-workspace.yaml":   "packages:\n  - apps/*\n",
		"pnpm-lock.yaml":        "",
		"apps/web/package.json": `{"name":"web","scripts":{"test":"vitest"}}`,
		"apps/api/package.json": `{"name":"api","scripts":{"test":"jest"}}`,
	}
	for name, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	out := string(ajiya(t, dir, 0, "init", "--name", "Demo", "--prefix", "DE", "--yes"))
	cfg, err := os.ReadFile(filepath.Join(dir, "ajiya.toml"))
	if err != nil {
		t.Fatal(err)
	}
	want := `command = "pnpm --filter api test && pnpm --filter web test"`
	if !strings.Contains(string(cfg), want) {
		t.Errorf("config lacks %s:\n%s", want, cfg)
	}
	if !strings.Contains(out, "Tests found per app") || strings.Contains(out, "No test command found") {
		t.Errorf("output = %s", out)
	}
}
