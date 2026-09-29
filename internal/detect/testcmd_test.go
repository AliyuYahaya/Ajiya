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
