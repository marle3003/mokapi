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

func Match(pattern, value string) bool {
	if len(pattern) == 0 {
		return false
	}
	value = filepath.ToSlash(value)

	pattern = strings.TrimSuffix(pattern, "/")

	if strings.HasPrefix(value, "/") {
		value = value[1:]
	}

	if !strings.Contains(pattern, "/") {
		return matchGlobal(pattern, value)
	}

	// If there is a separator at the beginning or middle (or both) of the pattern, then the pattern is relative
	if strings.HasPrefix(pattern, "/") {
		pattern = pattern[1:]
	}

	return match(pattern, value)
}

func matchGlobal(pattern, s string) bool {
	segments := strings.Split(s, "/")
	chunks := strings.Split(pattern, "/")
Segments:
	for i := range segments {
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
Pattern:
	for len(pattern) > 0 {
		var star starMode
		var chunk string
		star, chunk, pattern = scanChunk(pattern)
		if star != none && len(chunk) == 0 {
			return true
		}
		ok, rest := matchChunk(chunk, s)
		if ok {
			if len(rest) == 0 || rest[0] == '/' || len(pattern) > 0 {
				s = rest
				continue
			}
		}
		if star == doubleStar {
			pattern = strings.TrimPrefix(chunk+pattern, "/")
			segments := strings.Split(s, "/")
			chunks := strings.Split(pattern, "/")
		Segments:
			for i := range segments {
				for j, chunk := range chunks {
					seg := ""
					if i+j < len(segments) {
						seg = segments[i+j]
					}
					if chunk == "**" {
						path := strings.Join(segments[i+j:], "/")
						if len(path) == 0 {
							return false
						}
						return Match(strings.Join(chunks[j:], "/"), path)
					} else if !match(chunk, seg) {
						continue Segments
					}
				}
				return true
			}
			return false
		}
		if star == singleStar {
			for i := 0; i < len(s) && s[i] != '/'; i++ {
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

	return true
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
			//case '/':
			//	// '*/foo' requires a parent element
			//	if mode == singleStar || mode == none {
			//		break Scan
			//	}
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
