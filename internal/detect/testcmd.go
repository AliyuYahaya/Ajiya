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
	cmd, _ := testCommandIn(root, root)
	return cmd
}

// testCommandIn is TestCommand for the folder dir, with the lockfile looked
// for in lockDir (a workspace keeps one at its root). It also returns the
// package.json name when the command comes from a package.json script.
func testCommandIn(dir, lockDir string) (cmd, pkgName string) {
	root := dir
	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(root, name))
		return err == nil
	}
	lockExists := func(name string) bool {
		_, err := os.Stat(filepath.Join(lockDir, name))
		return err == nil
	}
	if exists("go.mod") {
		return "go test ./...", ""
	}
	if data, err := os.ReadFile(filepath.Join(root, "package.json")); err == nil {
		var pkg struct {
			Name    string            `json:"name"`
			Scripts map[string]string `json:"scripts"`
		}
		script := ""
		if json.Unmarshal(data, &pkg) == nil {
			script = strings.TrimSpace(pkg.Scripts["test"])
		}
		// npm init writes a placeholder test script that only fails.
		if script != "" && !strings.Contains(script, "no test specified") {
			switch {
			case lockExists("pnpm-lock.yaml"):
				return "pnpm test", pkg.Name
			case lockExists("yarn.lock"):
				return "yarn test", pkg.Name
			case lockExists("bun.lockb"), lockExists("bun.lock"):
				return "bun test", pkg.Name
			}
			return "npm test", pkg.Name
		}
	}
	if data, err := os.ReadFile(filepath.Join(root, "pyproject.toml")); err == nil && strings.Contains(string(data), "pytest") {
		return "pytest", ""
	}
	return "", ""
}

// AppTestCommand suggests a test command from the registered apps' manifests,
// for a root that has none. appPaths are slash paths relative to root. Each
// app with its own test command contributes one step: "pnpm --filter <name>
// test" when the root has a pnpm lockfile and the app's package.json has a
// name, else "cd <path> && <the app's command>". Several steps are joined
// with " && " so one command runs them all. It returns the steps too, for
// showing one per app, and "" when no app has tests.
func AppTestCommand(root string, appPaths []string) (command string, steps []string) {
	for _, p := range appPaths {
		if p == "" || p == "." {
			continue
		}
		cmd, name := testCommandIn(filepath.Join(root, filepath.FromSlash(p)), root)
		if cmd == "" {
			continue
		}
		if name != "" && cmd == "pnpm test" {
			cmd = "pnpm --filter " + name + " test"
		} else {
			cmd = "cd " + p + " && " + cmd
		}
		steps = append(steps, cmd)
	}
	return strings.Join(steps, " && "), steps
}
