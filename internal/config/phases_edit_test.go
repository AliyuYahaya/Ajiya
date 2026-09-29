package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestSetPhaseOrderText(t *testing.T) {
	const head = "[project]\nname = \"X\"\nprefix = \"X\"\n"
	order := []string{"api", "go-live"}
	tests := []struct{ name, in, want string }{
		{"replace, keeping comment",
			head + "\n[phases]\norder = ['old']  # how we read it\n",
			head + "\n[phases]\norder = [\"api\", \"go-live\"]  # how we read it\n"},
		{"replace a list over several lines",
			head + "\n[phases]\norder = [\n  \"a\",  # first ]\n  'b]',\n]\n\n[test]\ncommand = \"x\"\n",
			head + "\n[phases]\norder = [\"api\", \"go-live\"]\n\n[test]\ncommand = \"x\"\n"},
		{"table without order",
			head + "\n[phases]\n\n[test]\ncommand = \"go test\"\n",
			head + "\n[phases]\norder = [\"api\", \"go-live\"]\n\n[test]\ncommand = \"go test\"\n"},
		{"no table, array tables follow",
			head + "\n[[apps]]\nname = \"a\"\npath = \"a\"\n",
			head + "\n[phases]\norder = [\"api\", \"go-live\"]\n\n[[apps]]\nname = \"a\"\npath = \"a\"\n"},
		{"no table, nothing follows",
			head,
			head + "\n[phases]\norder = [\"api\", \"go-live\"]\n"},
		{"order in another table is not touched",
			head + "\n[test]\norder = [\"x\"]\n",
			head + "\n[test]\norder = [\"x\"]\n\n[phases]\norder = [\"api\", \"go-live\"]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SetPhaseOrderText(tt.in, order)
			if got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestSetPhaseOrderTextLong(t *testing.T) {
	const head = "[project]\nname = \"X\"\nprefix = \"X\"\n"
	order := []string{"accounts", "bookings", "admin-console", "reporting", "notifications", "go-live"}
	got := SetPhaseOrderText(head+"\n[phases]\norder = []\n", order)
	want := head + "\n[phases]\norder = [\n  \"accounts\",\n  \"bookings\",\n  \"admin-console\",\n  \"reporting\",\n  \"notifications\",\n  \"go-live\",\n]\n"
	if got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	c, err := Parse([]byte(got))
	if err != nil || !slices.Equal(c.Phases.Order, order) {
		t.Fatalf("parse: %v %v", c, err)
	}
	// Writing it again replaces the whole list.
	again := SetPhaseOrderText(got, []string{"a"})
	if again != head+"\n[phases]\norder = [\"a\"]\n" {
		t.Errorf("rewrite:\n%s", again)
	}
}

func TestSetPhaseOrder(t *testing.T) {
	root := t.TempDir()
	in := "[project]\nname = \"X\"\nprefix = \"X\"\n"
	os.WriteFile(filepath.Join(root, FileName), []byte(in), 0o644)
	c, err := SetPhaseOrder(root, []string{"b", "a"})
	if err != nil || !slices.Equal(c.Phases.Order, []string{"b", "a"}) {
		t.Fatalf("SetPhaseOrder: %v %v", c, err)
	}
	// An invalid list is refused and the file is left alone.
	if _, err := SetPhaseOrder(root, []string{"a", "a"}); err == nil || !strings.Contains(err.Error(), `lists "a" twice`) {
		t.Fatalf("duplicate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(root, FileName))
	if !strings.Contains(string(data), `order = ["b", "a"]`) {
		t.Errorf("file changed:\n%s", data)
	}
}

func TestPhaseOrderValidate(t *testing.T) {
	const head = "[project]\nname = \"X\"\nprefix = \"X\"\n\n[phases]\n"
	for in, want := range map[string]string{
		`order = ["a", "B"]`: `"B" is not a phase slug`,
		`order = ["a", "a"]`: `lists "a" twice`,
		`ordr = ["a"]`:       `unknown key "phases.ordr"`,
	} {
		_, err := Parse([]byte(head + in + "\n"))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got %v, want %q", in, err, want)
		}
	}
}
