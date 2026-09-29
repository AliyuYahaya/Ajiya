package config

import (
	"strings"
	"testing"
)

func TestWithTestCommand(t *testing.T) {
	text := WithTestCommand(Template("Demo", "DE"), "npm test")
	if !strings.Contains(text, "\n[test]\ncommand = \"npm test\"   # run by") {
		t.Errorf("no test command written:\n%s", text)
	}
	c, err := Parse([]byte(text))
	if err != nil || c.Test.Command != "npm test" {
		t.Errorf("Parse = %+v, %v", c, err)
	}
	if WithTestCommand("no example here", "x") != "no example here" {
		t.Error("text without the example changed")
	}
}
