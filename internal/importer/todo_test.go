package importer

import (
	"reflect"
	"testing"
)

func TestParseTodo(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []TodoItem
	}{
		{"empty", "", nil},
		{"dash", "- Buy milk", []TodoItem{{1, "Buy milk", false}}},
		{"star and plus", "* One\n+ Two", []TodoItem{{1, "One", false}, {2, "Two", false}}},
		{"numbered", "1. First\n2) Second\n10. Tenth", []TodoItem{{1, "First", false}, {2, "Second", false}, {3, "Tenth", false}}},
		{"unticked", "- [ ] Write tests", []TodoItem{{1, "Write tests", false}}},
		{"ticked", "- [x] Ship it\n- [X] Tell people", []TodoItem{{1, "Ship it", true}, {2, "Tell people", true}}},
		{"numbered checkbox", "1. [x] Done", []TodoItem{{1, "Done", true}}},
		{"nested", "- Parent\n  - Child\n    1. [x] Grandchild", []TodoItem{{1, "Parent", false}, {2, "Child", false}, {3, "Grandchild", true}}},
		{"tab indented", "\t- Tabbed", []TodoItem{{1, "Tabbed", false}}},
		{"text kept as written", "-   Fix  the |pipe| and **bold**  ", []TodoItem{{1, "Fix  the |pipe| and **bold**", false}}},
		{"bracket later is text", "- Read [x] later", []TodoItem{{1, "Read [x] later", false}}},
		{"box without space is text", "- [x]done", []TodoItem{{1, "[x]done", false}}},
		{"empty items skipped", "-\n- [ ]\n- [x]   \n*", nil},
		{"headings and paragraphs ignored", "# Title\n\nSome text.\n-not an item\n1.5 kg", nil},
		{"thematic breaks ignored", "---\n* * *\n- - -\n___", nil},
		{"crlf", "- One\r\n- [x] Two\r\n", []TodoItem{{1, "One", false}, {2, "Two", true}}},
		{"backtick fence", "- Before\n```\n- Inside\n```\n- After", []TodoItem{{1, "Before", false}, {5, "After", false}}},
		{"tilde fence", "~~~go\n- Inside\n~~~\n- After", []TodoItem{{4, "After", false}}},
		{"fence closes only with same kind and length", "````\n```\n- Inside\n~~~~\n````\n- After", []TodoItem{{6, "After", false}}},
		{"indented fence", "- Item\n  ```\n  - code\n  ```", []TodoItem{{1, "Item", false}}},
		{"unclosed fence", "```\n- Inside", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTodo(tt.src)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTodo(%q)\n got %+v\nwant %+v", tt.src, got, tt.want)
			}
		})
	}
}
