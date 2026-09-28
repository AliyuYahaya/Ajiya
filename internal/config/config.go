// Package config loads and validates ajiya.toml.
package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/BurntSushi/toml"
)

// FileName is the config file at the project root.
const FileName = "ajiya.toml"

// InfraApp is always a valid app name, for work outside app folders.
const InfraApp = "infra"

// ErrNotFound means no ajiya.toml exists in the directory or any parent.
var ErrNotFound = errors.New("no " + FileName + " found in this directory or any parent; run 'ajiya init'")

type Config struct {
	Project Project `toml:"project"`
	Launch  Launch  `toml:"launch"`
	Test    Test    `toml:"test"`
	Apps    []App   `toml:"apps"`

	// Root is the directory holding ajiya.toml. Set by Load.
	Root string `toml:"-"`
}

type Project struct {
	Name   string `toml:"name"`
	Prefix string `toml:"prefix"`
}

type Launch struct {
	Target string `toml:"target"` // a phase slug or a ticket ID
}

type Test struct {
	Command string `toml:"command"`
}

type App struct {
	Name string `toml:"name"`
	Path string `toml:"path"`
	Kind string `toml:"kind"` // "" or "library"
}

var (
	prefixRE  = regexp.MustCompile(`^[A-Z][A-Z0-9]{0,9}$`)
	appNameRE = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)
)

// Find walks up from dir to the nearest directory containing ajiya.toml.
func Find(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, FileName)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotFound
		}
		dir = parent
	}
}

// Load reads and validates root/ajiya.toml.
func Load(root string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(root, FileName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	c, err := Parse(data)
	if err != nil {
		return nil, err
	}
	c.Root = root
	return c, nil
}

// Parse decodes and validates the contents of an ajiya.toml file.
func Parse(data []byte) (*Config, error) {
	var c Config
	md, err := toml.Decode(string(data), &c)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", FileName, strings.TrimPrefix(err.Error(), "toml: "))
	}
	if und := md.Undecoded(); len(und) > 0 {
		return nil, fmt.Errorf("%s: unknown key %q; check the spelling", FileName, und[0].String())
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

// Validate reports the first problem in the config, in one line.
func (c *Config) Validate() error {
	fail := func(format string, a ...any) error {
		return fmt.Errorf(FileName+": "+format, a...)
	}
	if strings.TrimSpace(c.Project.Name) == "" {
		return fail("[project] name is missing; add name = \"<project name>\"")
	}
	if !prefixRE.MatchString(c.Project.Prefix) {
		return fail("[project] prefix %q must be 1 to 10 capital letters or digits, starting with a letter", c.Project.Prefix)
	}
	seen := map[string]bool{}
	for i, a := range c.Apps {
		switch {
		case !appNameRE.MatchString(a.Name):
			return fail("app %d: name %q must be lower case letters, digits, '.', '_' or '-'", i+1, a.Name)
		case a.Name == InfraApp:
			return fail("app name %q is reserved: it is always available; remove this app", InfraApp)
		case seen[a.Name]:
			return fail("app %q is registered twice; remove one", a.Name)
		case a.Kind != "" && a.Kind != "library":
			return fail("app %q: kind %q is not known; use kind = \"library\" or leave it out", a.Name, a.Kind)
		}
		if err := checkAppPath(a.Path); err != nil {
			return fail("app %q: %v", a.Name, err)
		}
		seen[a.Name] = true
	}
	return nil
}

func checkAppPath(p string) error {
	switch {
	case p == "":
		return errors.New("path is missing; use \".\" for the repository root")
	case strings.Contains(p, `\`) || path.IsAbs(p) || filepath.IsAbs(p):
		return fmt.Errorf("path %q must be relative to the project root, with / separators", p)
	case path.Clean(p) != p:
		return fmt.Errorf("path %q is not clean; write it as %q", p, path.Clean(p))
	case p == ".." || strings.HasPrefix(p, "../"):
		return fmt.Errorf("path %q is outside the project root", p)
	}
	return nil
}

// HasApp reports whether name is a registered app or infra.
func (c *Config) HasApp(name string) bool {
	if name == InfraApp {
		return true
	}
	for _, a := range c.Apps {
		if a.Name == name {
			return true
		}
	}
	return false
}

// AppNames returns every valid app name, including infra, sorted.
func (c *Config) AppNames() []string {
	names := []string{InfraApp}
	for _, a := range c.Apps {
		names = append(names, a.Name)
	}
	sort.Strings(names)
	return names
}

// SuggestPrefix derives an ID prefix from a project name: the initials of a
// multi-word name, or the first two letters of a single word.
func SuggestPrefix(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r)) || r > unicode.MaxASCII
	})
	var p string
	if len(words) >= 2 {
		for _, w := range words {
			if len(p) < 4 {
				p += w[:1]
			}
		}
	} else if len(words) == 1 {
		p = words[0]
		if len(p) > 2 {
			p = p[:2]
		}
	}
	p = strings.ToUpper(p)
	if p == "" || !unicode.IsLetter(rune(p[0])) {
		p = "T" + p
	}
	return p
}

// Template is the ajiya.toml written by init.
func Template(name, prefix string) string {
	return fmt.Sprintf(`# Ajiya configuration: https://github.com/AliyuYahaya/Ajiya

[project]
name = %q
prefix = %q

# [launch]
# target = "go-live"          # a phase slug or a ticket ID

# [test]
# command = "go test ./..."   # run by 'ajiya ticket done --test'

# [[apps]]
# name = "api"
# path = "apps/api"
`, name, prefix)
}
