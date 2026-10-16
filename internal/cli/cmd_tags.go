package cli

import (
	"fmt"
	"text/tabwriter"
)

// runTags lists every tag in use with the number of tickets carrying it, so
// that new tickets reuse an existing value instead of a near-duplicate.
func runTags(e *env, args []string) error {
	fs := newFlags("tags")
	asJSON := fs.Bool("json", false, "print JSON")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	counts := pr.plan.TagCounts()
	if *asJSON {
		type tag struct {
			Key     string `json:"key"`
			Value   string `json:"value"`
			Tickets int    `json:"tickets"`
		}
		items := []tag{}
		for _, c := range counts {
			items = append(items, tag{c.Key, c.Value, c.Tickets})
		}
		return writeJSON(e, items)
	}
	if len(counts) == 0 {
		fmt.Fprintln(e.stdout, "No tags in use.")
		return nil
	}
	w := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
	for _, c := range counts {
		fmt.Fprintf(w, "%s:%s\t%d\n", c.Key, c.Value, c.Tickets)
	}
	return w.Flush()
}
