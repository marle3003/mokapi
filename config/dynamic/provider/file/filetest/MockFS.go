package filetest

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Entry struct {
	Name    string
	IsDir   bool
	Data    []byte
	ModTime time.Time
}

type MockFS struct {
	Entries    []*Entry
	WorkingDir string
}

type fileInfo struct {
	entry *Entry
}

func (m *MockFS) ReadFile(path string) ([]byte, error) {
	path = strings.ReplaceAll(path, string(filepath.Separator), "/")
	path = strings.ReplaceAll(path, "C:/", "/")
	for _, entry := range m.Entries {
		if entry.Name == path {
			return entry.Data, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockFS) Walk(root string, visit fs.WalkDirFunc) error {
	var ignoreDirs []string

Walk:
	for _, entry := range m.Entries {
		tmp := strings.ReplaceAll(entry.Name, "/", string(filepath.Separator))
		rel, _ := filepath.Rel(root, tmp)
		if strings.HasPrefix(rel, "..") {
			continue
		}
		for _, dir := range ignoreDirs {
			if strings.HasPrefix(entry.Name, dir) {
				continue Walk
			}
		}
		f := &fileInfo{entry: entry}
		if err := visit(entry.Name, f, nil); err == fs.SkipDir {
			ignoreDirs = append(ignoreDirs, entry.Name)
		}
	}
	return nil
}

func (m *MockFS) GetWorkingDir() (string, error) {
	return m.WorkingDir, nil
}

func (m *MockFS) Stat(name string) (fs.FileInfo, error) {
	path := strings.ReplaceAll(name, string(filepath.Separator), "/")
	path = strings.ReplaceAll(path, "C:/", "/")
	for _, entry := range m.Entries {
		if entry.Name == path {
			return &fileInfo{entry: entry}, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (f *fileInfo) Name() string {
	return f.entry.Name
}

func (f *fileInfo) IsDir() bool {
	return f.entry.IsDir
}

func (f *fileInfo) Type() fs.FileMode {
	if f.IsDir() {
		return fs.ModeDir
	}
	return 0
}

func (f *fileInfo) Info() (fs.FileInfo, error) {
	return nil, nil
}

func (f *fileInfo) Size() int64 {
	return int64(len(f.entry.Data))
}

func (f *fileInfo) Mode() os.FileMode {
	return os.FileMode(0)
}

func (f *fileInfo) ModTime() time.Time {
	return f.entry.ModTime
}

func (f *fileInfo) Sys() any {
	return 0
}
