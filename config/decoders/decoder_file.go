package decoders

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

var (
	searchPaths  = []string{".", "/etc/mokapi"}
	fileNames    = []string{"mokapi.yaml", "mokapi.yml"}
	fileNotFound = errors.New("file not found")
)

type ReadFileFS func(path string) ([]byte, error)

type FileDecoder struct {
	filename string
	readFile ReadFileFS
}

func NewDefaultFileDecoder() *FileDecoder {
	return NewFileDecoder(os.ReadFile)
}

func NewFileDecoder(readFile ReadFileFS) *FileDecoder {
	return &FileDecoder{readFile: readFile}
}

func (f *FileDecoder) Decode(flags map[string]string, element interface{}) error {
	if len(f.filename) == 0 {
		if val, ok := flags["configfile"]; ok {
			f.filename = val
		}
	}

	if len(f.filename) > 0 {
		return f.read(f.filename, element)
	}

	for _, dir := range searchPaths {
		for _, name := range fileNames {
			path := filepath.Join(dir, name)
			if err := f.read(path, element); err == nil {
				return nil
			}
		}
	}

	return nil
}

func (f *FileDecoder) read(path string, element interface{}) error {
	data, err := f.readFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, element)
	if err != nil {
		return errors.Wrapf(err, "parsing YAML file %s", f.filename)
	}
	return nil
}
