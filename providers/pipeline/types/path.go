package types

import (
	"regexp"
)

type Path struct {
	full     string
	segments []string
	index    int
}

func ParsePath(s string) *Path {
	path := &Path{segments: make([]string, 0), index: 0, full: s}
	pat := regexp.MustCompile(`(?:\["?'?)?([\w\*]+)(?:\.|"?'?])?(.*)`)

	for {
		matches := pat.FindAllStringSubmatch(s, -1) // matches is [][]string
		if len(matches) == 0 {
			break
		}

		path.segments = append(path.segments, matches[0][1])
		s = matches[0][2]
	}

	return path
}

func (p *Path) Head() string {
	if p.index >= len(p.segments) {
		return ""
	}
	return p.segments[p.index]
}

// Returns true if the path successfully moved to the next segment;
// false if the path has passed the end of segments
func (p *Path) MoveNext() bool {
	p.index++
	if p.index == len(p.segments) {
		return false
	}
	return true
}

func (p *Path) String() string {
	return p.full
}

func (p *Path) Copy() *Path {
	return &Path{full: p.full, segments: p.segments, index: p.index}
}
