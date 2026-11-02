package require

import (
	"sort"
	"strings"
)

type aliasEntry struct {
	prefix   string
	suffix   string
	wildcard bool
	target   string
}

type aliasTable []aliasEntry

func newAliasTable(aliases map[string]string) aliasTable {
	t := make(aliasTable, 0, len(aliases))
	for key, target := range aliases {
		e := aliasEntry{prefix: key, target: target}
		if before, after, ok := strings.Cut(key, "*"); ok {
			e.prefix, e.suffix, e.wildcard = before, after, true
		}
		t = append(t, e)
	}
	// exact keys first, then longest prefix, then longest suffix
	sort.SliceStable(t, func(i, j int) bool {
		a, b := t[i], t[j]
		if a.wildcard != b.wildcard {
			return !a.wildcard
		}
		if len(a.prefix) != len(b.prefix) {
			return len(a.prefix) > len(b.prefix)
		}
		if len(a.suffix) != len(b.suffix) {
			return len(a.suffix) > len(b.suffix)
		}
		return a.prefix+a.suffix < b.prefix+b.suffix
	})
	return t
}

func (t aliasTable) resolve(modPath string) string {
	for _, e := range t {
		if !e.wildcard {
			if modPath == e.prefix {
				return e.target
			}
			continue
		}
		if len(modPath) < len(e.prefix)+len(e.suffix) ||
			!strings.HasPrefix(modPath, e.prefix) ||
			!strings.HasSuffix(modPath, e.suffix) {
			continue
		}
		matched := modPath[len(e.prefix) : len(modPath)-len(e.suffix)]
		return strings.Replace(e.target, "*", matched, 1)
	}
	return modPath
}
