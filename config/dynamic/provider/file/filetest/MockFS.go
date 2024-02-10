package filetest

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
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
	Entries    map[string]*Entry
	WorkingDir string
}

type fileInfo struct {
	entry *Entry
}

func (m *MockFS) ReadFile(path string) ([]byte, error) {
	path = strings.ReplaceAll(path, string(filepath.Separator), "/")
	path = strings.ReplaceAll(path, "C:/", "/")
	if f, ok := m.Entries[path]; ok {
		return f.Data, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockFS) Walk(root string, visit fs.WalkDirFunc) error {
	var ignoreDirs []string

	// loop order in a map is not safe, order path to ensure SkipDir
	keys := make([]string, 0, len(m.Entries))
	for k := range m.Entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
Walk:
	for _, path := range keys {
		tmp := strings.ReplaceAll(path, "/", string(filepath.Separator))
		rel, _ := filepath.Rel(root, tmp)
		if strings.HasPrefix(rel, "..") {
			continue
		}
		for _, dir := range ignoreDirs {
			if strings.HasPrefix(path, dir) {
				continue Walk
			}
		}
		f := &fileInfo{entry: m.Entries[path]}
		if err := visit(path, f, nil); err == fs.SkipDir {
			ignoreDirs = append(ignoreDirs, path)
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
	if e, ok := m.Entries[path]; ok {
		return &fileInfo{entry: e}, nil
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
