package agents

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

// Codex's config.toml is edited as text so that comments and layout survive: a
// TOML round trip drops them. The table is appended, or its lines are cut out,
// and the result is parsed to prove nothing else changed.

const tomlMarker = "# Added by ajiya mcp install"

var headerRE = regexp.MustCompile(`^\s*\[\[?\s*([^\[\]]+?)\s*\]\]?\s*(#.*)?$`)

func tomlParse(p string, data []byte) (map[string]any, error) {
	m := map[string]any{}
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%s is not valid TOML (%v); fix or remove it and run this again. Nothing was changed", p, err)
	}
	return m, nil
}

func tomlEntry(p string, m map[string]any) (entry, bool, error) {
	raw, ok := m["mcp_servers"]
	if !ok {
		return entry{}, false, nil
	}
	tbl, ok := raw.(map[string]any)
	if !ok {
		return entry{}, false, fmt.Errorf("%s: mcp_servers is not a table; fix it and run this again. Nothing was changed", p)
	}
	srv, ok := tbl[ServerName]
	if !ok {
		return entry{}, false, nil
	}
	var en entry
	if t, ok := srv.(map[string]any); ok {
		en.Command, _ = t["command"].(string)
		if as, ok := t["args"].([]any); ok {
			for _, a := range as {
				s, _ := a.(string)
				en.Args = append(en.Args, s)
			}
		}
	}
	return en, true, nil
}

func tomlLookup(p string, data []byte) (entry, bool, error) {
	m, err := tomlParse(p, data)
	if err != nil {
		return entry{}, false, err
	}
	return tomlEntry(p, m)
}

func tomlQuote(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}

func tomlBlock(en entry) []string {
	args := make([]string, len(en.Args))
	for i, a := range en.Args {
		args[i] = tomlQuote(a)
	}
	return []string{tomlMarker, "[mcp_servers." + ServerName + "]", "command = " + tomlQuote(en.Command), "args = [" + strings.Join(args, ", ") + "]"}
}

// isAjiyaHeader reports whether a line is a table header, and whether it opens
// [mcp_servers.ajiya] or a table below it.
func isAjiyaHeader(l string) (header, ours bool) {
	m := headerRE.FindStringSubmatch(l)
	if m == nil {
		return false, false
	}
	name := strings.NewReplacer(`"`, "", `'`, "", " ", "", "\t", "").Replace(m[1])
	pre := "mcp_servers." + ServerName
	return true, name == pre || strings.HasPrefix(name, pre+".")
}

// cutAjiya removes the ajiya table (and its sub-tables) from the lines.
func cutAjiya(ls []string) (out, removed []string) {
	i := 0
	toEnd := false
	for i < len(ls) {
		h, ours := isAjiyaHeader(ls[i])
		if !(h && ours) {
			out = append(out, ls[i])
			i++
			continue
		}
		if n := len(out); n > 0 && strings.TrimSpace(out[n-1]) == tomlMarker { // our own comment goes too
			removed = append(removed, out[n-1])
			out = out[:n-1]
		}
		for i < len(ls) {
			if h, ours := isAjiyaHeader(ls[i]); h && !ours {
				break
			}
			removed = append(removed, ls[i])
			i++
		}
		// comments right above the next table belong to that table
		for i < len(ls) && len(removed) > 0 && strings.HasPrefix(strings.TrimSpace(removed[len(removed)-1]), "#") {
			removed = removed[:len(removed)-1]
			i--
		}
		toEnd = i >= len(ls)
	}
	if toEnd { // the table ran to the end of the file: drop the blank lines before it
		for len(out) > 0 && strings.TrimSpace(out[len(out)-1]) == "" {
			out = out[:len(out)-1]
		}
	}
	return out, removed
}

func tomlEdit(p string, old []byte, en entry, install, present bool) ([]byte, []string, error) {
	before, err := tomlParse(p, old)
	if err != nil {
		return nil, nil, err
	}
	pv := []string{"--- " + p, "+++ " + p}
	ls := lines(old)
	if present {
		var removed []string
		ls, removed = cutAjiya(ls)
		for _, l := range removed {
			if strings.TrimSpace(l) != "" {
				pv = append(pv, "- "+l)
			}
		}
	}
	if install {
		if len(ls) > 0 && strings.TrimSpace(ls[len(ls)-1]) != "" {
			ls = append(ls, "")
		}
		for _, l := range tomlBlock(en) {
			ls = append(ls, l)
			pv = append(pv, "+ "+l)
		}
	}
	text := strings.Join(ls, "\n")
	if text != "" {
		text += "\n"
	}
	if strings.Contains(string(old), "\r\n") {
		text = strings.ReplaceAll(text, "\n", "\r\n")
	}
	hand := fmt.Sprintf("add or remove [mcp_servers.%s] by hand. Nothing was changed", ServerName)
	after, err := tomlParse(p, []byte(text))
	if err != nil {
		return nil, nil, fmt.Errorf("could not edit %s safely; %s", p, hand)
	}
	if err := verifyTOML(before, after, en, install); err != nil {
		return nil, nil, fmt.Errorf("could not edit %s safely (%v); %s", p, err, hand)
	}
	return []byte(text), pv, nil
}

// verifyTOML checks that everything but the ajiya entry is as it was, and that
// the entry is as wanted.
func verifyTOML(before, after map[string]any, en entry, install bool) error {
	strip := func(m map[string]any) map[string]any {
		out := map[string]any{}
		for k, v := range m {
			out[k] = v
		}
		if t, ok := m["mcp_servers"].(map[string]any); ok {
			c := map[string]any{}
			for k, v := range t {
				if k != ServerName {
					c[k] = v
				}
			}
			if len(c) > 0 {
				out["mcp_servers"] = c
			} else {
				delete(out, "mcp_servers")
			}
		}
		return out
	}
	if !reflect.DeepEqual(strip(before), strip(after)) {
		return fmt.Errorf("the edit would change other settings")
	}
	got, ok, err := tomlEntry("", after)
	if err != nil {
		return err
	}
	if install && (!ok || !reflect.DeepEqual(got, en)) {
		return fmt.Errorf("the new entry did not read back")
	}
	if !install && ok {
		return fmt.Errorf("the entry is still there")
	}
	return nil
}
