package detect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tree writes files ("path" or "path=content") under a new temp dir.
func tree(t *testing.T, files ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, f := range files {
		name, content, _ := strings.Cut(f, "=")
		p := filepath.Join(root, filepath.FromSlash(name))
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func show(cs []Candidate) string {
	var lines []string
	for _, c := range cs {
		kind := ""
		if c.Kind != "" {
			kind = " " + c.Kind
		}
		lines = append(lines, fmt.Sprintf("%s=%s%s [%s] %s", c.Name, c.Path, kind, c.Source, strings.Join(c.Signs, ",")))
	}
	return strings.Join(lines, "\n")
}

func TestSuggest(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  string
	}{
		{"single root manifest",
			[]string{"go.mod", "main.go"},
			"clinic-booking=. [manifest] "},
		{"pnpm workspace, root excluded, apps ranked by signs",
			[]string{
				"package.json={}",
				"pnpm-workspace.yaml=packages:\n  - 'apps/*'\n  - \"packages/*\" # shared\n  - '!apps/old'\nother: x\n",
				`apps/web/package.json={"scripts":{"dev":"next dev"}}`, "apps/web/next.config.js",
				"apps/api/package.json={}", "apps/api/Dockerfile",
				"apps/docs/package.json={}",
				"packages/ui/package.json={}",
				"node_modules/x/package.json={}",
			},
			"web=apps/web [pnpm-workspace.yaml] dev script,next.config\n" +
				"api=apps/api [pnpm-workspace.yaml] Dockerfile\n" +
				"docs=apps/docs [pnpm-workspace.yaml] \n" +
				"ui=packages/ui library [pnpm-workspace.yaml] "},
		{"npm workspaces object form, and a stray manifest",
			[]string{
				`package.json={"workspaces":{"packages":["services/**"]}}`,
				"services/billing/api/package.json={}", "services/billing/api/Dockerfile",
				"tools/cli/go.mod",
			},
			"api=services/billing/api [package.json workspaces] Dockerfile\n" +
				"cli=tools/cli [manifest] "},
		{"go.work",
			[]string{"go.work=go 1.26\n\nuse (\n\t./cmd/server // the API\n\t./lib\n)\nuse ./tools\n",
				"cmd/server/go.mod", "cmd/server/Dockerfile", "lib/go.mod", "tools/go.mod"},
			"server=cmd/server [go.work] Dockerfile\n" +
				"lib=lib library [go.work] \n" +
				"tools=tools [go.work] "},
		{"nx projects and python and php apps",
			[]string{"nx.json={}", "apps/shop/project.json", "apps/shop/package.json={}",
				"backend/pyproject.toml", "backend/manage.py", "portal/composer.json", "portal/artisan"},
			"shop=apps/shop [nx project.json] \n" +
				"backend=backend [manifest] manage.py\n" +
				"portal=portal [manifest] artisan"},
		{"same folder name twice",
			[]string{"apps/web/package.json={}", "admin/web/package.json={}"},
			"web=admin/web [manifest] \napps-web=apps/web [manifest] "},
		{"nothing to find", []string{"README.md"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Suggest(tree(t, tt.files...), "Clinic Booking")
			if err != nil {
				t.Fatal(err)
			}
			if show(got) != tt.want {
				t.Errorf("got\n%s\nwant\n%s", show(got), tt.want)
			}
		})
	}
}
