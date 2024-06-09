package decoders

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

var (
	searchPaths = []string{".", "/etc/mokapi"}
	fileNames   = []string{"mokapi.yaml", "mokapi.yml"}
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

func (f *FileDecoder) Decode(flags map[string][]string, element interface{}) error {
	if len(f.filename) == 0 {
		if val, ok := flags["configfile"]; ok {
			f.filename = val[0]
		} else if val, ok := flags["config-file"]; ok {
			f.filename = val[0]
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
			} else if !os.IsNotExist(err) {
				return err
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
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, element)
	case ".json":
		err = json.Unmarshal(data, element)
	default:
		err = fmt.Errorf("unsupported file extension: %v", filepath.Ext(path))
	}

	if err != nil {
		return fmt.Errorf("parse file '%v' failed: %w", path, err)
	}
	return nil
}
