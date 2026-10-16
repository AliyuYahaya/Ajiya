package plan

import (
	"path"
	"strings"
)

// GlobMatch reports whether the file name (a path from the project root, with
// / separators) is matched by pattern. A pattern is matched segment by
// segment: * and ? work inside a segment as in path.Match, and a segment of **
// matches any number of segments, none included. A pattern that matches a
// folder the file is in also matches the file, so "infra/rds" and "infra/*"
// match "infra/rds/main.tf".
func GlobMatch(pattern, name string) bool {
	pat := strings.Split(strings.Trim(pattern, "/"), "/")
	segs := strings.Split(strings.Trim(name, "/"), "/")
	for n := 1; n <= len(segs); n++ { // the whole path, then each folder it is in
		if matchSegs(pat, segs[:n]) {
			return true
		}
	}
	return false
}

func matchSegs(pat, segs []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			for len(pat) > 0 && pat[0] == "**" {
				pat = pat[1:]
			}
			if len(pat) == 0 {
				return true
			}
			for i := 0; i <= len(segs); i++ {
				if matchSegs(pat, segs[i:]) {
					return true
				}
			}
			return false
		}
		if len(segs) == 0 {
			return false
		}
		if ok, err := path.Match(pat[0], segs[0]); err != nil || !ok {
			return false
		}
		pat, segs = pat[1:], segs[1:]
	}
	return len(segs) == 0
}

// staticDir is the folder part of a glob before its first wildcard.
func staticDir(glob string) string {
	g := strings.Trim(glob, "/")
	if i := strings.IndexAny(g, "*?["); i >= 0 {
		g = g[:i]
		if j := strings.LastIndex(g, "/"); j >= 0 {
			g = g[:j]
		} else {
			g = ""
		}
	}
	return g
}

// GlobsOverlap reports whether two globs may cover the same files: one matches
// the other written out, or one's fixed leading folder contains the other's.
// It errs towards yes, because a change only needs to know which tickets might
// be touched.
func GlobsOverlap(a, b string) bool {
	if a == b || GlobMatch(a, b) || GlobMatch(b, a) {
		return true
	}
	da, db := staticDir(a), staticDir(b)
	within := func(x, y string) bool { return y == "" || x == y || strings.HasPrefix(x, y+"/") }
	return within(da, db) || within(db, da)
}
