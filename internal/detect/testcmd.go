package detect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// TestCommand returns the command that runs the project's tests when the
// files at root show it, else "". go.mod gives "go test ./...", a package.json
// "test" script gives "<manager> test" (the lockfile picks pnpm, yarn or bun,
// else npm), and a pyproject.toml that mentions pytest gives "pytest".
func TestCommand(root string) string {
	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(root, name))
		return err == nil
	}
	if exists("go.mod") {
		return "go test ./..."
	}
	if data, err := os.ReadFile(filepath.Join(root, "package.json")); err == nil {
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
		}
		script := ""
		if json.Unmarshal(data, &pkg) == nil {
			script = strings.TrimSpace(pkg.Scripts["test"])
		}
		// npm init writes a placeholder test script that only fails.
		if script != "" && !strings.Contains(script, "no test specified") {
			switch {
			case exists("pnpm-lock.yaml"):
				return "pnpm test"
			case exists("yarn.lock"):
				return "yarn test"
			case exists("bun.lockb"), exists("bun.lock"):
				return "bun test"
			}
			return "npm test"
		}
	}
	if data, err := os.ReadFile(filepath.Join(root, "pyproject.toml")); err == nil && strings.Contains(string(data), "pytest") {
		return "pytest"
	}
	return ""
}
