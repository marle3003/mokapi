package clitest

import (
	"os"
	"strings"
)

type TestFileReader struct {
	Files map[string][]byte
}

func (t *TestFileReader) ReadFile(name string) ([]byte, error) {
	if b, ok := t.Files[name]; ok {
		return b, nil
	}
	// if test is executed on windows we get second path
	name = strings.ReplaceAll(name, "\\", "/")
	if b, ok := t.Files[name]; ok {
		return b, nil
	}
	return nil, os.ErrNotExist
}

func (t *TestFileReader) FileExists(name string) bool {
	if _, ok := t.Files[name]; ok {
		return true
	}
	name = strings.ReplaceAll(name, "\\", "/")
	_, ok := t.Files[name]
	return ok
}
