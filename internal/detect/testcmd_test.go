package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTestCommand(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{"go", map[string]string{"go.mod": "module x\n"}, "go test ./..."},
		{"npm", map[string]string{"package.json": `{"scripts":{"test":"jest"}}`}, "npm test"},
		{"pnpm", map[string]string{"package.json": `{"scripts":{"test":"jest"}}`, "pnpm-lock.yaml": ""}, "pnpm test"},
		{"yarn", map[string]string{"package.json": `{"scripts":{"test":"jest"}}`, "yarn.lock": ""}, "yarn test"},
		{"bun", map[string]string{"package.json": `{"scripts":{"test":"jest"}}`, "bun.lockb": ""}, "bun test"},
		{"no test script", map[string]string{"package.json": `{"scripts":{"dev":"vite"}}`}, ""},
		{"npm placeholder", map[string]string{"package.json": `{"scripts":{"test":"echo \"Error: no test specified\" && exit 1"}}`}, ""},
		{"bad json", map[string]string{"package.json": `{`}, ""},
		{"pytest", map[string]string{"pyproject.toml": "[tool.pytest.ini_options]\n"}, "pytest"},
		{"python without pytest", map[string]string{"pyproject.toml": "[project]\nname = \"x\"\n"}, ""},
		{"nothing", map[string]string{}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, body := range c.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if got := TestCommand(dir); got != c.want {
				t.Errorf("TestCommand = %q, want %q", got, c.want)
			}
		})
	}
}

func TestAppTestCommand(t *testing.T) {
	write := func(t *testing.T, dir string, files map[string]string) {
		for name, body := range files {
			full := filepath.Join(dir, filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	cases := []struct {
		name  string
		files map[string]string
		paths []string
		want  string
	}{
		{"pnpm workspace, tests in apps only", map[string]string{
			"package.json":          `{"name":"root","private":true}`,
			"pnpm-lock.yaml":        "",
			"apps/web/package.json": `{"name":"web","scripts":{"test":"vitest"}}`,
			"apps/api/package.json": `{"name":"api","scripts":{"test":"jest"}}`,
		}, []string{"apps/web", "apps/api"}, "pnpm --filter web test && pnpm --filter api test"},
		{"one app has tests", map[string]string{
			"pnpm-lock.yaml":        "",
			"apps/web/package.json": `{"name":"web","scripts":{"test":"vitest"}}`,
			"apps/api/package.json": `{"name":"api","scripts":{"dev":"node ."}}`,
		}, []string{"apps/web", "apps/api"}, "pnpm --filter web test"},
		{"npm apps", map[string]string{
			"apps/web/package.json": `{"name":"web","scripts":{"test":"jest"}}`,
		}, []string{"apps/web"}, "cd apps/web && npm test"},
		{"go app", map[string]string{"svc/go.mod": "module x\n"}, []string{"svc"}, "cd svc && go test ./..."},
		{"no app tests", map[string]string{
			"pnpm-lock.yaml":        "",
			"apps/web/package.json": `{"name":"web"}`,
		}, []string{"apps/web"}, ""},
		{"no apps", map[string]string{}, nil, ""},
		{"root path skipped", map[string]string{"package.json": `{"scripts":{"test":"jest"}}`}, []string{"."}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			write(t, dir, c.files)
			got, steps := AppTestCommand(dir, c.paths)
			if got != c.want {
				t.Errorf("AppTestCommand = %q, want %q", got, c.want)
			}
			if (got == "") != (len(steps) == 0) {
				t.Errorf("steps = %v, command %q", steps, got)
			}
		})
	}
}
