package dashboard

import (
	"regexp"
	"sort"
	"strings"
	"testing"
)

func pageText(t *testing.T) string {
	t.Helper()
	s := string(HTML())
	if len(s) == 0 {
		t.Fatal("embedded page is empty")
	}
	return s
}

// The page must make no network requests, so it holds no web URLs at all.
func TestNoURLs(t *testing.T) {
	s := strings.ToLower(pageText(t))
	for _, u := range []string{"http://", "https://", "//fonts.", "@import"} {
		if strings.Contains(s, u) {
			t.Errorf("page contains %q; it must not reach the network", u)
		}
	}
}

func TestLoadsDataJS(t *testing.T) {
	s := pageText(t)
	if !strings.Contains(s, `<script src="data.js"></script>`) {
		t.Error("page does not load data.js")
	}
	if !strings.Contains(s, "window.AJIYA") {
		t.Error("page does not read window.AJIYA")
	}
	if !strings.Contains(s, "ajiya build") {
		t.Error("page does not tell the user to run ajiya build when data is missing")
	}
}

var varDecl = regexp.MustCompile(`(--[a-z0-9-]+)\s*:`)

// block returns the text between the '{' after marker and its matching '}'.
func block(t *testing.T, s, marker string) string {
	t.Helper()
	i := strings.Index(s, marker)
	if i < 0 {
		t.Fatalf("no %q in page", marker)
	}
	start := i + strings.Index(s[i:], "{")
	depth := 0
	for j := start; j < len(s); j++ {
		switch s[j] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start+1 : j]
			}
		}
	}
	t.Fatalf("unclosed block after %q", marker)
	return ""
}

func vars(b string) []string {
	var out []string
	for _, m := range varDecl.FindAllStringSubmatch(b, -1) {
		out = append(out, m[1])
	}
	sort.Strings(out)
	return out
}

// Light variables on :root, and the same set for dark, both for the system
// preference and for the toggle.
func TestThemeVariables(t *testing.T) {
	s := pageText(t)
	light := vars(block(t, s, ":root {"))
	if len(light) < 10 {
		t.Fatalf("only %d colour variables on :root", len(light))
	}
	system := vars(block(t, block(t, s, "@media (prefers-color-scheme: dark)"), `:root:not([data-theme="light"])`))
	toggle := vars(block(t, s, `:root[data-theme="dark"]`))
	for name, dark := range map[string][]string{"prefers-color-scheme": system, "data-theme=dark": toggle} {
		if strings.Join(dark, " ") != strings.Join(light, " ") {
			t.Errorf("dark (%s) variables differ from light:\n light %v\n dark  %v", name, light, dark)
		}
	}
}

// Every view registered in VIEWS has a nav button and a renderer, and every
// nav button has a view.
func TestViewsInNav(t *testing.T) {
	s := pageText(t)
	reg := block(t, s, "const VIEWS =")
	var views []string
	for _, m := range regexp.MustCompile(`(?m)^\s*(\w+): \{ label:`).FindAllStringSubmatch(reg, -1) {
		views = append(views, m[1])
	}
	want := []string{"overview", "next", "board", "phases", "tickets", "apps", "checks", "activity"}
	if strings.Join(views, " ") != strings.Join(want, " ") {
		t.Errorf("views = %v, want %v", views, want)
	}
	nav := map[string]bool{}
	for _, m := range regexp.MustCompile(`class="nav-item" type="button" data-tab="(\w+)"`).FindAllStringSubmatch(s, -1) {
		nav[m[1]] = true
	}
	renderers := block(t, s, "const renderers =")
	for _, v := range views {
		if !nav[v] {
			t.Errorf("view %q has no nav button", v)
		}
		if !regexp.MustCompile(`\b` + v + `: render\w+`).MatchString(renderers) {
			t.Errorf("view %q has no renderer", v)
		}
		delete(nav, v)
	}
	for v := range nav {
		t.Errorf("nav button %q has no view", v)
	}
}
