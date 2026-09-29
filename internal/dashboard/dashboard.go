// Package dashboard holds the dashboard page, embedded in the binary. The page
// reads its data from data.js next to it (window.AJIYA), so it works opened
// from disk as well as through 'ajiya serve'. It makes no network requests.
package dashboard

import _ "embed"

//go:embed index.html
var page []byte

// HTML returns the dashboard page.
func HTML() []byte { return page }
