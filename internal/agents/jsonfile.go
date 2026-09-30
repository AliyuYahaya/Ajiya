package agents

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// obj is a JSON object that keeps the order of its keys and the exact text of
// its values, so an edit changes only what it means to.
type obj struct {
	keys []string
	vals map[string]json.RawMessage
}

func newObj() *obj { return &obj{vals: map[string]json.RawMessage{}} }

func parseObj(data []byte) (*obj, error) {
	o := newObj()
	if len(bytes.TrimSpace(data)) == 0 {
		return o, nil
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	if tok, err := dec.Token(); err != nil {
		return nil, err
	} else if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, errors.New("the top level is not a JSON object")
	}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key := tok.(string)
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		if _, dup := o.vals[key]; !dup {
			o.keys = append(o.keys, key)
		}
		o.vals[key] = raw
	}
	if _, err := dec.Token(); err != nil { // the closing brace
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, errors.New("unexpected text after the JSON object")
	}
	return o, nil
}

func (o *obj) set(k string, v json.RawMessage) {
	if _, ok := o.vals[k]; !ok {
		o.keys = append(o.keys, k)
	}
	o.vals[k] = v
}

func (o *obj) del(k string) {
	delete(o.vals, k)
	for i, x := range o.keys {
		if x == k {
			o.keys = append(o.keys[:i:i], o.keys[i+1:]...)
			break
		}
	}
}

func jsonString(s string) []byte {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.Encode(s)
	return bytes.TrimSpace(b.Bytes())
}

func (o *obj) compact() []byte {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, k := range o.keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.Write(jsonString(k))
		b.WriteByte(':')
		b.Write(o.vals[k])
	}
	b.WriteByte('}')
	return b.Bytes()
}

func indent(compact []byte) []byte {
	var out bytes.Buffer
	json.Indent(&out, compact, "", "  ")
	out.WriteByte('\n')
	return out.Bytes()
}

// servers returns the top-level object and its mcpServers object.
func servers(p string, data []byte) (*obj, *obj, error) {
	top, err := parseObj(data)
	if err != nil {
		return nil, nil, fmt.Errorf("%s is not valid JSON (%v); fix or remove it and run this again. Nothing was changed", p, err)
	}
	srv := newObj()
	if raw, ok := top.vals["mcpServers"]; ok {
		if srv, err = parseObj(raw); err != nil || bytes.HasPrefix(bytes.TrimSpace(raw), []byte("null")) {
			return nil, nil, fmt.Errorf("%s: \"mcpServers\" is not a JSON object; fix it and run this again. Nothing was changed", p)
		}
	}
	return top, srv, nil
}

func jsonLookup(p string, data []byte) (entry, bool, error) {
	_, srv, err := servers(p, data)
	if err != nil {
		return entry{}, false, err
	}
	raw, ok := srv.vals[ServerName]
	if !ok {
		return entry{}, false, nil
	}
	var en entry
	var v struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}
	if json.Unmarshal(raw, &v) == nil {
		en = entry{v.Command, v.Args}
	}
	return en, true, nil
}

func entryJSON(en entry) json.RawMessage {
	var b bytes.Buffer
	b.WriteString(`{"type":"stdio","command":`)
	b.Write(jsonString(en.Command))
	b.WriteString(`,"args":[`)
	for i, a := range en.Args {
		if i > 0 {
			b.WriteByte(',')
		}
		b.Write(jsonString(a))
	}
	b.WriteString(`]}`)
	return b.Bytes()
}

// previewEntry shows one entry as diff lines with a "+" or "-" mark.
func previewEntry(mark string, raw json.RawMessage) []string {
	var out bytes.Buffer
	json.Indent(&out, raw, "    ", "  ")
	ls := strings.Split(out.String(), "\n")
	ls[0] = `    "` + ServerName + `": ` + ls[0]
	for i := range ls {
		ls[i] = mark + ls[i]
	}
	return ls
}

// jsonEdit adds, replaces or removes the ajiya entry and returns the new file
// and a diff-style preview.
func jsonEdit(p string, old []byte, en entry, install, present bool) ([]byte, []string, error) {
	top, srv, err := servers(p, old)
	if err != nil {
		return nil, nil, err
	}
	pv := []string{"--- " + p, "+++ " + p, `  "mcpServers": {   (other entries are kept)`}
	if present {
		pv = append(pv, previewEntry("- ", srv.vals[ServerName])...)
	}
	if install {
		raw := entryJSON(en)
		srv.set(ServerName, raw)
		pv = append(pv, previewEntry("+ ", raw)...)
	} else {
		srv.del(ServerName)
	}
	top.set("mcpServers", srv.compact())
	return indent(top.compact()), pv, nil
}
