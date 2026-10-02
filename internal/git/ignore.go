package git

import (
	"path"
	"strings"
)

type ignoreRule struct {
	domain   string   // slash-separated dir of the .gitignore, "" for the root one
	segs     []string // anchored glob split by "/"
	base     string   // unanchored glob matched against the base name
	negate   bool
	dirOnly  bool
	anchored bool
}

// Matcher applies .gitignore rules with git's "last matching pattern wins"
// semantics. Paths passed to Ignored are slash-separated and relative to the
// root. Callers should stop descending into ignored directories, which also
// gives git's "cannot re-include a file under an excluded dir" behaviour.
type Matcher struct {
	rules []ignoreRule
}

// Add parses the contents of a .gitignore located in domain (a slash-separated
// directory relative to the root, "" for the root) and appends its rules.
// Add nested files after their parents so deeper rules take precedence.
func (m *Matcher) Add(content, domain string) {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSuffix(line, "\r")
		for strings.HasSuffix(line, " ") && !strings.HasSuffix(line, `\ `) {
			line = line[:len(line)-1]
		}
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		r := ignoreRule{domain: domain}
		if strings.HasPrefix(line, "!") {
			r.negate = true
			line = line[1:]
		}
		if strings.HasSuffix(line, "/") {
			r.dirOnly = true
			line = strings.TrimSuffix(line, "/")
		}
		if line == "" {
			continue
		}
		if strings.Contains(line, "/") {
			r.anchored = true
			r.segs = strings.Split(strings.TrimPrefix(line, "/"), "/")
		} else {
			r.base = line
		}
		m.rules = append(m.rules, r)
	}
}

// Ignored reports whether rel (relative to the root, slash-separated) is ignored.
func (m *Matcher) Ignored(rel string, isDir bool) bool {
	for i := len(m.rules) - 1; i >= 0; i-- {
		if m.rules[i].match(rel, isDir) {
			return !m.rules[i].negate
		}
	}
	return false
}

func (r ignoreRule) match(rel string, isDir bool) bool {
	if r.dirOnly && !isDir {
		return false
	}
	if r.domain != "" {
		if !strings.HasPrefix(rel, r.domain+"/") {
			return false
		}
		rel = rel[len(r.domain)+1:]
	}
	if r.anchored {
		return matchSegs(r.segs, strings.Split(rel, "/"))
	}
	ok, _ := path.Match(r.base, path.Base(rel))
	return ok
}

// matchSegs matches glob segments against path segments; "**" spans any
// number of segments.
func matchSegs(pat, segs []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			for len(pat) > 0 && pat[0] == "**" {
				pat = pat[1:]
			}
			if len(pat) == 0 {
				return len(segs) > 0
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
