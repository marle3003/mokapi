package file

import (
	"io/fs"
	"os"
	"path/filepath"
)

type FSReader interface {
	Walk(string, fs.WalkDirFunc) error
	ReadFile(name string) ([]byte, error)
	Stat(name string) (os.FileInfo, error)
	GetWorkingDir() (string, error)
}

type Reader struct {
}

func (r *Reader) Walk(root string, f fs.WalkDirFunc) error {
	return filepath.WalkDir(root, f)
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
