package urlpath

import "strings"

type wildcard int

const (
	none wildcard = iota
	star
	doubleStar
)

const Separator = '/'

func Match(pattern string, value string) bool {
Pattern:
	for len(pattern) > 0 {

		var tok wildcard
		var chunk string
		tok, chunk, pattern = scanChunk(pattern)

		if tok == star && len(chunk) == 0 {
			return !strings.Contains(value, "/")
		}

		rest, ok := matchChunk(chunk, value)

		if ok && (len(rest) == 0 || len(pattern) > 0) {
			value = rest
			continue
		}

		switch tok {
		case star, doubleStar:
			for len(value) > 0 && (tok == doubleStar || value[0] != Separator) {
				value = value[1:]
				rest, ok := matchChunk(chunk, value)
				if ok {
					value = rest
					continue Pattern
				}
			}
		}

		return false
	}

	return len(value) == 0
}

func scanChunk(pattern string) (token wildcard, chunk, rest string) {
	for len(pattern) > 0 && pattern[0] == '*' {
		switch token {
		case star:
			token = doubleStar
		case doubleStar:
		case none:
			token = star
		}
		pattern = pattern[1:]
	}

	var i int
	for i = 0; i < len(pattern); i++ {
		if pattern[i] == '*' {
			break
		}
	}

	return token, pattern[0:i], pattern[i:]
}

func matchChunk(chunk, s string) (rest string, ok bool) {
	for len(chunk) > 0 {
		if len(s) == 0 {
			return
		}

		if chunk[0] != s[0] {
			return
		}

		chunk = chunk[1:]
		s = s[1:]
	}

	return s, true
}
