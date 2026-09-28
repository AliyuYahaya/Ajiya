package plan

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var idRE = regexp.MustCompile(`^([A-Z][A-Z0-9]*)-(\d{4,})$`)

// SplitID splits an ID such as CB-0142 into its prefix and number.
func SplitID(id string) (prefix string, n int, ok bool) {
	m := idRE.FindStringSubmatch(id)
	if m == nil {
		return "", 0, false
	}
	n, err := strconv.Atoi(m[2])
	if err != nil {
		return "", 0, false
	}
	return m[1], n, true
}

// ValidID reports whether id has the given prefix and the <PREFIX>-<4 digits> form.
func ValidID(prefix, id string) bool {
	p, n, ok := SplitID(id)
	return ok && p == prefix && FormatID(prefix, n) == id
}

// FormatID formats an ID with at least four digits.
func FormatID(prefix string, n int) string {
	return fmt.Sprintf("%s-%04d", prefix, n)
}

// LessID orders IDs by prefix, then number, with malformed IDs last.
func LessID(a, b string) bool {
	pa, na, oka := SplitID(a)
	pb, nb, okb := SplitID(b)
	switch {
	case oka != okb:
		return oka
	case !oka:
		return a < b
	case pa != pb:
		return pa < pb
	case na != nb:
		return na < nb
	}
	return a < b
}

// SortIDs sorts ids in place with LessID.
func SortIDs(ids []string) {
	sort.SliceStable(ids, func(i, j int) bool { return LessID(ids[i], ids[j]) })
}

// ParseIDList splits a list of IDs separated by commas or spaces. "-" and ""
// mean none. The result is sorted and has no duplicates.
func ParseIDList(s string) []string {
	seen := map[string]bool{}
	var ids []string
	for _, f := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' }) {
		if f == "-" || seen[f] {
			continue
		}
		seen[f] = true
		ids = append(ids, f)
	}
	SortIDs(ids)
	return ids
}
