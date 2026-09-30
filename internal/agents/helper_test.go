package agents

import (
	"fmt"
	"os"
	"testing"
)

// TestMain lets the test binary act as a stand-in child process that prints
// its CLAUDE_CONFIG_DIR, for TestRunnerPassesClaudeConfigDir.
func TestMain(m *testing.M) {
	if os.Getenv("AJIYA_TEST_PRINT_ENV") == "1" && len(os.Args) > 1 && os.Args[1] == "-test.run=^$" {
		fmt.Printf("CLAUDE_CONFIG_DIR=%s\n", os.Getenv("CLAUDE_CONFIG_DIR"))
		return
	}
	os.Exit(m.Run())
}
