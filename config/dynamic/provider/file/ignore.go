package file

import (
	"path/filepath"
	"strings"
)

type IgnoreFiles map[string]*IgnoreFile

type IgnoreFile struct {
	entries []string
}

func newIgnoreFile(b []byte) (*IgnoreFile, error) {
	return &IgnoreFile{entries: strings.Split(string(b), "\n")}, nil
}

func (i *IgnoreFile) Match(path string) bool {
	return i.MatchWithReInclude(path, false)
}

func (i *IgnoreFile) MatchWithReInclude(path string, skipReInclude bool) bool {
	result := false
	for _, e := range i.entries {
		e = strings.TrimSpace(e)
		if strings.HasPrefix(e, "#") || len(e) == 0 {
			continue
		}
		if e[0] == '!' && !skipReInclude {
			if Match(e[1:], path) {
				return i.MatchWithReInclude(filepath.Dir(path), true)
			}
		}
		if Match(e, path) {
			result = true
		}
	}
	return result
}

func (m IgnoreFiles) Match(path string) bool {
	var file *IgnoreFile
	precedence := -1
	for k, v := range m {
		p := len(k)
		if k == "." || strings.HasPrefix(path, k) && p > precedence {
			file = v
			precedence = len(k)
		}
	}
	if file != nil {
		return file.Match(path)
	}
	return false
}
