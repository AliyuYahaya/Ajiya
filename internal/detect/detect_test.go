package detect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestDirs(t *testing.T) {
	root := t.TempDir()
	for _, f := range []string{
		"go.mod", "apps/api/package.json", "apps/web/pyproject.toml", "apps/web/src/x.py",
		"services/Pay.Api/Pay.Api.csproj", "node_modules/left-pad/package.json",
		"apps/api/node_modules/x/package.json", ".cache/go.mod", "testdata/go.mod", "examples/demo/package.json",
	} {
		p := filepath.Join(root, filepath.FromSlash(f))
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, nil, 0o644)
	}
	dirs, err := ManifestDirs(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(dirs, " "); got != ". apps/api apps/web services/Pay.Api" {
		t.Errorf("ManifestDirs = %s", got)
	}
}

func TestUnder(t *testing.T) {
	for _, tt := range []struct {
		p, dir string
		want   bool
	}{
		{"apps/api/main.go", "apps/api", true},
		{"apps/api", "apps/api", true},
		{"apps/api2/x", "apps/api", false},
		{"README.md", ".", true},
		{"apps", "apps/api", false},
	} {
		if got := Under(tt.p, tt.dir); got != tt.want {
			t.Errorf("Under(%q, %q) = %v", tt.p, tt.dir, got)
		}
	}
}
