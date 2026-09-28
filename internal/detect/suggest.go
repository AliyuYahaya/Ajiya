package detect

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Candidate is a folder that looks like an app, with the reasons why.
type Candidate struct {
	Name   string
	Path   string // slash path relative to the root, "." for the root
	Kind   string // "" or "library"
	Source string // where it was found: a workspace file, or "manifest"
	Signs  []string
}

// Suggest lists the apps in root, most reliable first: folders named by a
// workspace file (pnpm, npm or yarn workspaces, go.work, Nx), then other
// folders with their own manifest. Within each group, folders with more signs
// of running on their own come first. project names the root app when the
// only manifest is at the root.
func Suggest(root, project string) ([]Candidate, error) {
	manifests, err := ManifestDirs(root)
	if err != nil {
		return nil, err
	}
	workspace := workspaceDirs(root, manifests)
	skipRoot := len(workspace) > 0 || len(manifests) > 1

	byPath := map[string]*Candidate{}
	for _, w := range workspace {
		if _, ok := byPath[w.path]; !ok {
			byPath[w.path] = &Candidate{Path: w.path, Source: w.source}
		}
	}
	for _, d := range manifests {
		if d == "." && skipRoot {
			continue // a workspace or monorepo root, not an app of its own
		}
		if _, ok := byPath[d]; !ok && !insideAny(d, byPath) {
			byPath[d] = &Candidate{Path: d, Source: "manifest"}
		}
	}

	var out []Candidate
	for _, c := range byPath {
		c.Signs = runSigns(filepath.Join(root, filepath.FromSlash(c.Path)))
		if len(c.Signs) == 0 && isLibraryDir(c.Path) {
			c.Kind = "library"
		}
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if (a.Source == "manifest") != (b.Source == "manifest") {
			return b.Source == "manifest"
		}
		if len(a.Signs) != len(b.Signs) {
			return len(a.Signs) > len(b.Signs)
		}
		return a.Path < b.Path
	})
	names := map[string]bool{}
	for i := range out {
		out[i].Name = uniqueName(out[i].Path, project, names)
	}
	return out, nil
}

// AppManifestDirs is ManifestDirs without the root when the root is a
// workspace or monorepo root rather than an app of its own.
func AppManifestDirs(root string) ([]string, error) {
	dirs, err := ManifestDirs(root)
	if err != nil || len(dirs) == 0 || dirs[0] != "." {
		return dirs, err
	}
	if len(dirs) > 1 || len(workspaceDirs(root, dirs)) > 0 {
		return dirs[1:], nil
	}
	return dirs, nil
}

func insideAny(d string, cands map[string]*Candidate) bool {
	for p := range cands {
		if p != "." && Under(d, p) {
			return true
		}
	}
	return false
}

type wsDir struct{ path, source string }

// workspaceDirs reads the workspace files at the root and returns the folders
// they name that exist.
func workspaceDirs(root string, manifests []string) []wsDir {
	var out []wsDir
	add := func(source string, patterns []string) {
		for _, d := range expand(root, patterns, manifests) {
			out = append(out, wsDir{d, source})
		}
	}
	if data, err := os.ReadFile(filepath.Join(root, "pnpm-workspace.yaml")); err == nil {
		add("pnpm-workspace.yaml", yamlList(string(data), "packages"))
	}
	if data, err := os.ReadFile(filepath.Join(root, "package.json")); err == nil {
		add("package.json workspaces", jsonWorkspaces(data))
	}
	if data, err := os.ReadFile(filepath.Join(root, "go.work")); err == nil {
		add("go.work", goWorkUses(string(data)))
	}
	// Nx: every project.json marks a project.
	var nx []string
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && p != root && (skipDirs[d.Name()] || strings.HasPrefix(d.Name(), ".")) {
			return filepath.SkipDir
		}
		if !d.IsDir() && d.Name() == "project.json" {
			if rel, err := filepath.Rel(root, filepath.Dir(p)); err == nil && rel != "." {
				nx = append(nx, filepath.ToSlash(rel))
			}
		}
		return nil
	})
	for _, d := range nx {
		out = append(out, wsDir{d, "nx project.json"})
	}
	return out
}

// expand turns workspace patterns such as apps/* or packages/** into folders.
func expand(root string, patterns, manifests []string) []string {
	var dirs []string
	for _, pat := range patterns {
		pat = strings.TrimPrefix(strings.Trim(pat, `"' `), "./")
		if pat == "" || strings.HasPrefix(pat, "!") {
			continue
		}
		if prefix, ok := strings.CutSuffix(pat, "/**"); ok {
			for _, m := range manifests {
				if m != "." && Under(m, prefix) && m != prefix {
					dirs = append(dirs, m)
				}
			}
			continue
		}
		matches, _ := filepath.Glob(filepath.Join(root, filepath.FromSlash(pat)))
		for _, m := range matches {
			if st, err := os.Stat(m); err == nil && st.IsDir() {
				if rel, err := filepath.Rel(root, m); err == nil {
					dirs = append(dirs, filepath.ToSlash(rel))
				}
			}
		}
	}
	return dirs
}

// yamlList reads a top-level list such as "packages:\n  - 'apps/*'".
func yamlList(text, key string) []string {
	var items []string
	in := false
	for line := range strings.SplitSeq(text, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, key+":"):
			in = true
		case in && strings.HasPrefix(trimmed, "- "):
			item, _, _ := strings.Cut(strings.TrimPrefix(trimmed, "- "), " #")
			items = append(items, strings.TrimSpace(item))
		case in && trimmed != "" && !strings.HasPrefix(trimmed, "#"):
			in = false
		}
	}
	return items
}

func jsonWorkspaces(data []byte) []string {
	var pkg struct {
		Workspaces json.RawMessage `json:"workspaces"`
	}
	if json.Unmarshal(data, &pkg) != nil || len(pkg.Workspaces) == 0 {
		return nil
	}
	var list []string
	if json.Unmarshal(pkg.Workspaces, &list) == nil {
		return list
	}
	var obj struct {
		Packages []string `json:"packages"`
	}
	json.Unmarshal(pkg.Workspaces, &obj)
	return obj.Packages
}

var goWorkUseRE = regexp.MustCompile(`(?m)^\s*use\s+(\S+)\s*$`)

func goWorkUses(text string) []string {
	var dirs []string
	for _, m := range goWorkUseRE.FindAllStringSubmatch(text, -1) {
		if m[1] != "(" {
			dirs = append(dirs, m[1])
		}
	}
	if i := strings.Index(text, "use ("); i >= 0 {
		block, _, _ := strings.Cut(text[i+len("use ("):], ")")
		for line := range strings.SplitSeq(block, "\n") {
			line, _, _ = strings.Cut(line, "//")
			if line = strings.TrimSpace(line); line != "" {
				dirs = append(dirs, line)
			}
		}
	}
	return dirs
}

// runSigns lists the signs that a folder runs on its own.
func runSigns(dir string) []string {
	var signs []string
	exists := func(name string) bool { _, err := os.Stat(filepath.Join(dir, name)); return err == nil }
	if exists("Dockerfile") {
		signs = append(signs, "Dockerfile")
	}
	if data, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
		}
		if json.Unmarshal(data, &pkg) == nil {
			for _, s := range []string{"start", "dev", "serve"} {
				if pkg.Scripts[s] != "" {
					signs = append(signs, s+" script")
					break
				}
			}
		}
	}
	for _, f := range []string{"artisan", "manage.py"} {
		if exists(f) {
			signs = append(signs, f)
		}
	}
	if m, _ := filepath.Glob(filepath.Join(dir, "next.config.*")); len(m) > 0 {
		signs = append(signs, "next.config")
	}
	return signs
}

func isLibraryDir(p string) bool {
	first, _, _ := strings.Cut(p, "/")
	switch first {
	case "packages", "libs", "lib", "shared":
		return true
	}
	return false
}

var nameCleanRE = regexp.MustCompile(`[^a-z0-9._-]+`)

func cleanName(s string) string {
	s = strings.Trim(nameCleanRE.ReplaceAllString(strings.ToLower(s), "-"), "-._")
	if s == "" || s == "infra" {
		s = "app"
	}
	return s
}

// uniqueName names an app after its folder, adding the parent folder when two
// folders share a name.
func uniqueName(p, project string, used map[string]bool) string {
	name := cleanName(path.Base(p))
	if p == "." {
		name = cleanName(project)
	}
	if used[name] && p != "." {
		name = cleanName(path.Base(path.Dir(p)) + "-" + path.Base(p))
	}
	for base, i := name, 2; used[name]; i++ {
		name = base + "-" + strconv.Itoa(i)
	}
	used[name] = true
	return name
}
