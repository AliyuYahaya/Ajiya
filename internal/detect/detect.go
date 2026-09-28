// Package detect finds folders that look like apps: folders with their own
// manifest file.
package detect

import (
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Manifests are the files that mark a folder as a package or app.
var Manifests = []string{"package.json", "composer.json", "pyproject.toml", "go.mod", "Cargo.toml", "Gemfile", "pubspec.yaml"}

// skipDirs are never searched: dependencies, build output, fixtures, examples.
var skipDirs = map[string]bool{
	"node_modules": true, "vendor": true, ".venv": true, "venv": true, "dist": true, "build": true,
	"fixtures": true, "examples": true, "example": true, "testdata": true, "target": true,
}

// IsManifest reports whether a file name is a manifest.
func IsManifest(name string) bool {
	for _, m := range Manifests {
		if name == m {
			return true
		}
	}
	return strings.HasSuffix(name, ".csproj")
}

// ManifestDirs returns every folder under root, as a slash path relative to
// root ("." for root itself), that holds a manifest, sorted.
func ManifestDirs(root string) ([]string, error) {
	seen := map[string]bool{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // unreadable folders are skipped, not fatal
		}
		name := d.Name()
		if d.IsDir() {
			if p != root && (skipDirs[name] || strings.HasPrefix(name, ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if IsManifest(name) {
			rel, err := filepath.Rel(root, filepath.Dir(p))
			if err == nil {
				seen[filepath.ToSlash(rel)] = true
			}
		}
		return nil
	})
	dirs := make([]string, 0, len(seen))
	for d := range seen {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	return dirs, err
}

// Under reports whether the slash path p is at or inside the folder dir.
func Under(p, dir string) bool {
	if dir == "." {
		return true
	}
	p, dir = path.Clean(p), path.Clean(dir)
	return p == dir || strings.HasPrefix(p, dir+"/")
}
