package serve

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/AliyuYahaya/Ajiya/internal/activity"
	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/gitx"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// stamp is what the watcher compares: a file's modification time and size.
// A missing file has the zero stamp.
type stamp struct {
	mod  time.Time
	size int64
}

// watcher finds the files a build depends on and their stamps.
type watcher struct {
	root    string
	head    string // path of the git HEAD file, "" outside a repository
	ref     string // the ref HEAD pointed to when last looked at
	refPath string // its file
	packed  string // packed-refs, where a ref lives once packed
}

func newWatcher(root string) *watcher {
	w := &watcher{root: root}
	if head, err := gitx.GitPath(root, "HEAD"); err == nil {
		w.head = w.abs(head)
		if packed, err := gitx.GitPath(root, "packed-refs"); err == nil {
			w.packed = w.abs(packed)
		}
	}
	return w
}

func (w *watcher) abs(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(w.root, p)
}

// snapshot returns the stamp of every watched file, keyed by a name for the
// log: a path relative to the root, or "git HEAD".
func (w *watcher) snapshot() map[string]stamp {
	snap := map[string]stamp{}
	add := func(name, path string) {
		if st, err := os.Stat(path); err == nil {
			snap[name] = stamp{st.ModTime(), st.Size()}
		} else {
			snap[name] = stamp{}
		}
	}
	add(config.FileName, filepath.Join(w.root, config.FileName))
	matches, _ := filepath.Glob(filepath.Join(w.root, plan.Dir, "*.md"))
	for _, m := range matches {
		name := plan.Dir + "/" + filepath.Base(m)
		if !slices.Contains(activity.Generated, name) {
			add(name, m)
		}
	}
	if w.head != "" {
		add("git HEAD", w.head)
		w.followHead()
		if w.refPath != "" {
			add("git "+w.ref, w.refPath)
		}
		if w.packed != "" {
			add("git packed-refs", w.packed)
		}
	}
	return snap
}

// followHead finds the file of the ref HEAD points to, asking git only when
// the ref changes. A detached HEAD has no ref file.
func (w *watcher) followHead() {
	data, err := os.ReadFile(w.head)
	if err != nil {
		return
	}
	ref, ok := strings.CutPrefix(strings.TrimSpace(string(data)), "ref: ")
	if !ok {
		w.ref, w.refPath = "", ""
		return
	}
	if ref == w.ref {
		return
	}
	w.ref, w.refPath = ref, ""
	if p, err := gitx.GitPath(w.root, ref); err == nil {
		w.refPath = w.abs(p)
	}
}

// changed lists the names whose stamps differ between two snapshots.
func changed(old, cur map[string]stamp) []string {
	var names []string
	for name, st := range cur {
		if o, ok := old[name]; !ok || o != st {
			names = append(names, name)
		}
	}
	for name := range old {
		if _, ok := cur[name]; !ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return slices.Compact(names)
}

// reason describes a change for the log.
func reason(names []string) string {
	if len(names) > 3 {
		return fmt.Sprintf("%s and %d more changed", strings.Join(names[:3], ", "), len(names)-3)
	}
	return strings.Join(names, ", ") + " changed"
}

// watch polls for changes until ctx is done and rebuilds on each one. Only
// one watch runs at a time, from Serve.
func (s *Server) watch(ctx context.Context) {
	t := time.NewTicker(s.opts.Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		cur := s.w.snapshot()
		names := changed(s.last, cur)
		s.last = cur
		if len(names) == 0 {
			continue
		}
		if err := s.rebuild(false); err != nil {
			fmt.Fprintf(s.opts.Log, "Rebuild failed: %s\n", firstLine(err.Error()))
			continue
		}
		fmt.Fprintf(s.opts.Log, "Rebuilt: %s\n", reason(names))
	}
}

func firstLine(s string) string {
	s, _, _ = strings.Cut(s, "\n")
	return s
}
