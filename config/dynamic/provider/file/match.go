package file

import (
	"path/filepath"
	"strings"
)

type starMode int

const (
	none       starMode = iota
	singleStar starMode = 2
	doubleStar starMode = 3
)

func Match(pattern, v string) bool {
	v = filepath.ToSlash(v)
	if !strings.HasPrefix(v, "/") {
		v = "/" + v
	}

	p := strings.TrimSuffix(pattern, "/")
	if !strings.Contains(p, "/") {
		// without a slash it means in any folders
		return matchGlobal(p, v)
	}

	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	if strings.HasPrefix(pattern, "**/") {
		pattern = strings.TrimPrefix(pattern, "**/")
		return matchGlobal(pattern, v)
	}

	if strings.Contains(pattern, "*") {
		return match(pattern, v)
	}

	pattern = strings.TrimPrefix(pattern, "/")
	v = strings.TrimPrefix(v, "/")

	return strings.HasPrefix(v, pattern)
}

func matchGlobal(pattern, s string) bool {
	s = strings.TrimPrefix(s, "/")
	segments := strings.Split(s, "/")
Segments:
	for i := range segments {
		chunks := strings.Split(strings.TrimPrefix(pattern, "/"), "/")
		for j, chunk := range chunks {
			if !match(chunk, segments[i+j]) {
				continue Segments
			}
		}
		return true

	}
	return false
}

func match(pattern, s string) bool {
	if strings.HasSuffix(pattern, "/") {
		// trailing slashing means a folder with any containing files
		pattern += "*"
	}
Pattern:
	for len(pattern) > 0 {
		var star starMode
		var chunk string
		star, chunk, pattern = scanChunk(pattern)
		ok, rest := matchChunk(chunk, s)
		if ok && (len(rest) == 0 || len(pattern) > 0) {
			s = rest
			continue
		}
		if star == doubleStar {
			p := chunk + pattern
			if len(p) == 0 {
				return true
			}
			return matchGlobal(p, s)
		}
		if star == singleStar {
			for i := 0; i < len(s); i++ {
				if i > 0 && s[i] == '/' && star == singleStar {
					if len(chunk) > 0 {
						break
					}
					s = s[i:]
					continue Pattern
				}
				ok, rest = matchChunk(chunk, s[i+1:])
				if ok {
					s = rest
					continue Pattern
				}
			}
			if len(pattern) == 0 && len(chunk) == 0 {
				return true
			}
		}

		return false
	}

	return len(s) == 0
}

func scanChunk(pattern string) (mode starMode, chunk, rest string) {
	for len(pattern) > 0 && pattern[0] == '*' {
		pattern = pattern[1:]
		if mode == singleStar {
			mode = doubleStar
		} else if mode == none {
			mode = singleStar
		}
	}
	var i int
Scan:
	for i = 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			break Scan
		case '/':
			// '*/foo' requires a parent element
			if mode == singleStar {
				break Scan
			}
		}
	}
	return mode, pattern[0:i], pattern[i:]
}

func matchChunk(chunk, s string) (ok bool, rest string) {
	if len(chunk) == 0 {
		return false, s
	}
	for len(chunk) > 0 {
		if len(s) == 0 {
			return false, s
		}
		switch chunk[0] {
		case '[':
			r := s[0]
			match := false
			chunk = chunk[1:]
			invert := false
			if len(chunk) > 0 && chunk[0] == '!' {
				chunk = chunk[1:]
				invert = true
			}
			for {
				if len(chunk) > 0 && chunk[0] == ']' {
					break
				}
				var start uint8
				start, chunk = consumeChar(chunk)
				end := start
				if len(chunk) > 0 && chunk[0] == '-' {
					chunk = chunk[1:]
					end, chunk = consumeChar(chunk)
				}
				if start <= r && r <= end {
					match = true
				}
			}
			if match == invert {
				return false, s
			}
			s = s[1:]
		case '?':
			s = s[1:]
		default:
			if chunk[0] != s[0] {
				return false, s
			}
			s = s[1:]
		}

		chunk = chunk[1:]
	}
	return true, s
}

func consumeChar(s string) (uint8, string) {
	if len(s) == 0 {
		return ' ', s
	}
	r := s[0]
	return r, s[1:]
}
