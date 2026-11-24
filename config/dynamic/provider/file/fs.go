package file

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	log "github.com/sirupsen/logrus"
)

type FSReader interface {
	Walk(string, fs.WalkDirFunc) error
	ReadFile(name string) ([]byte, error)
	Stat(name string) (os.FileInfo, error)
	GetWorkingDir() (string, error)
	Abs(string) (string, error)
}

type Reader struct {
}

// Walk is like filepath.WalkDir but follows symbolic links, so that linked
// directories are read like normal ones. Paths passed to f keep the link path.
// Links that lead back to a directory being walked (cycles) and broken links are skipped.
func (r *Reader) Walk(root string, f fs.WalkDirFunc) error {
	return walkFollowingLinks(root, f, nil)
}

// walkFollowingLinks walks the resolved root. followed holds the resolved locations of
// the links taken so far; a directory reachable via several links is read once per link.
func walkFollowingLinks(root string, f fs.WalkDirFunc, followed []string) error {
	real, err := filepath.EvalSymlinks(root)
	if err != nil {
		return f(root, nil, err)
	}

	return filepath.WalkDir(real, func(realPath string, d fs.DirEntry, err error) error {
		// cannot fail: WalkDir only passes paths below real
		rel, _ := filepath.Rel(real, realPath)
		path := filepath.Join(root, rel)
		if err != nil || d.Type()&fs.ModeSymlink == 0 {
			return f(path, d, err)
		}

		info, err := os.Stat(path)
		if err != nil {
			log.Debugf("skip broken symlink: %v", path)
			return nil
		}
		if !info.IsDir() {
			return f(path, fs.FileInfoToDirEntry(info), nil)
		}

		target, err := filepath.EvalSymlinks(path)
		if err != nil {
			log.Debugf("skip broken symlink: %v", path)
			return nil
		}
		followed := append(slices.Clone(followed), realPath)
		for _, location := range followed {
			if isSub(target, location) {
				log.Debugf("skip symlink cycle: %v -> %v", path, target)
				return nil
			}
		}
		return walkFollowingLinks(path, f, followed)
	})
}

func (r *Reader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (r *Reader) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (r *Reader) GetWorkingDir() (string, error) {
	return os.Getwd()
}

func (r *Reader) Abs(path string) (string, error) { return filepath.Abs(path) }
