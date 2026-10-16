package plan

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Tag keys. All tags are optional; a ticket with none is valid.
const (
	TagTech = "tech" // a technology or service the ticket depends on
	TagDec  = "dec"  // a decision ID
	TagPath = "path" // a file glob
)

// TagKeys lists the keys in the order they are written.
var TagKeys = []string{TagTech, TagDec, TagPath}

var decRE = regexp.MustCompile(`^DEC-[0-9]{4,}$`)

// Tags is the Tags cell of a ticket, parsed. Each list is sorted and has no
// repeats, so a tag set has one written form.
type Tags struct {
	Tech []string
	Dec  []string
	Path []string
}

// Empty reports whether there are no tags.
func (t Tags) Empty() bool { return len(t.Tech)+len(t.Dec)+len(t.Path) == 0 }

// Values returns the values stored under key, or nil for an unknown key.
func (t Tags) Values(key string) []string {
	switch key {
	case TagTech:
		return t.Tech
	case TagDec:
		return t.Dec
	case TagPath:
		return t.Path
	}
	return nil
}

// With returns the tags with the values of key replaced. Values are checked,
// sorted and de-duplicated; an empty list clears the key.
func (t Tags) With(key string, values []string) (Tags, error) {
	vs, err := cleanTagValues(key, values)
	if err != nil {
		return t, err
	}
	switch key {
	case TagTech:
		t.Tech = vs
	case TagDec:
		t.Dec = vs
	case TagPath:
		t.Path = vs
	}
	return t, nil
}

// String is the Tags cell: "key:value" pairs, or "-" for none.
func (t Tags) String() string {
	var parts []string
	for _, k := range TagKeys {
		for _, v := range t.Values(k) {
			parts = append(parts, k+":"+v)
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " ")
}

// ParseTags reads a Tags cell. An unknown key or a bad value is an error.
func ParseTags(cell string) (Tags, error) {
	var t Tags
	if cell == "-" {
		return t, nil
	}
	for _, f := range strings.Fields(cell) {
		key, val, ok := strings.Cut(f, ":")
		if !ok {
			return Tags{}, fmt.Errorf("tag %q is not key:value", f)
		}
		var err error
		if t, err = t.With(key, append(append([]string(nil), t.Values(key)...), val)); err != nil {
			return Tags{}, err
		}
	}
	return t, nil
}

// ParseTagList splits a comma-separated flag value. Empty and "-" mean none.
func ParseTagList(s string) []string {
	var out []string
	for _, v := range strings.Split(s, ",") {
		if v = strings.TrimSpace(v); v != "" && v != "-" {
			out = append(out, v)
		}
	}
	return out
}

// CheckTag reports why value cannot be stored under key.
func CheckTag(key, value string) error {
	switch key {
	case TagTech:
		if !SlugRE.MatchString(value) {
			return fmt.Errorf("tech tag %q must be a lower-case slug such as aws-rds", value)
		}
	case TagDec:
		if !decRE.MatchString(value) {
			return fmt.Errorf("dec tag %q must be a decision ID such as DEC-0003", value)
		}
	case TagPath:
		if value == "" || strings.ContainsAny(value, " \t|") {
			return fmt.Errorf("path tag %q must be a file glob without spaces or |", value)
		}
	default:
		return fmt.Errorf("tag key %q is not one of %s", key, strings.Join(TagKeys, ", "))
	}
	return nil
}

func cleanTagValues(key string, values []string) ([]string, error) {
	var out []string
	for _, v := range values {
		if err := CheckTag(key, v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	sort.Strings(out)
	w := 0
	for i, v := range out {
		if i == 0 || v != out[i-1] {
			out[w] = v
			w++
		}
	}
	return out[:w], nil
}

// TagCount is one tag in use and how many tickets carry it.
type TagCount struct {
	Key     string
	Value   string
	Tickets int
}

// TagCounts lists every tag in use, by key order then value.
func (p *Plan) TagCounts() []TagCount {
	n := map[[2]string]int{}
	for _, t := range p.Tickets() {
		for _, k := range TagKeys {
			for _, v := range t.Tags.Values(k) {
				n[[2]string{k, v}]++
			}
		}
	}
	var out []TagCount
	for _, k := range TagKeys {
		var vs []string
		for kv := range n {
			if kv[0] == k {
				vs = append(vs, kv[1])
			}
		}
		sort.Strings(vs)
		for _, v := range vs {
			out = append(out, TagCount{k, v, n[[2]string{k, v}]})
		}
	}
	return out
}
