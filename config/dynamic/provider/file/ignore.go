package file

import (
	"path/filepath"
	"strings"
)

const StartPrecedence = -1

type IgnoreFiles map[string]*IgnoreFile

type IgnoreFile struct {
	path    string
	entries []string
}

func newIgnoreFile(path string, b []byte) (*IgnoreFile, error) {
	return &IgnoreFile{
		path:    path,
		entries: strings.Split(string(b), "\n"),
	}, nil
}

func (i *IgnoreFile) Match(path string) bool {
	base := filepath.Dir(i.path)
	if base != "." {
		path = strings.TrimPrefix(path, base)
	}

	result := false
	for _, e := range i.entries {
		e = strings.TrimSpace(e)
		if strings.HasPrefix(e, "#") || len(e) == 0 {
			continue
		}
		if e[0] == '!' {
			if Match(e[1:], path) {
				// include file again
				result = false
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
	precedence := StartPrecedence
	for k, v := range m {
		if path == k {
			// path is folder containing this ignore file, skip
			continue
		}
		p := len(strings.Split(k, string(filepath.Separator)))
		if k == "." && precedence == StartPrecedence {
			file = v
			precedence = 0
		}
		if strings.HasPrefix(path, k) && p > precedence {
			file = v
			precedence = len(strings.Split(k, string(filepath.Separator)))
		}
	}
	if file != nil {
		return file.Match(path)
	}
	return false
}
