package tls

import (
	"os"
	"strings"
)

type FileOrContent string

func (f FileOrContent) Read(currentDir string) ([]byte, error) {
	path := f.String()
	if strings.HasPrefix(path, "./") {
		path = strings.Replace(path, ".", currentDir, 1)
	}

	if isPath(path) {
		return os.ReadFile(path)
	}
	return []byte(f.String()), nil
}

func isPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (f FileOrContent) String() string {
	return string(f)
}
