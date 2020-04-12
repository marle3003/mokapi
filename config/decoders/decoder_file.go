package decoders

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type FileDecoder struct {
	filename string
}

func NewFileDecoder(filename string) *FileDecoder {
	return &FileDecoder{filename: filename}
}

func (f *FileDecoder) Decode(flags map[string]string, element interface{}) error {
	if len(f.filename) == 0 {
		if val, ok := flags["configfile"]; ok {
			f.filename = val
		}
	}

	data, error := loadFile(f.filename)
	if error != nil {
		return error
	}

	err := parseYml(data, element)
	if err != nil {
		return errors.Wrapf(err, "parsing YAML file %s", f.filename)
	}
	return nil
}

func loadFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func parseYml(data []byte, element interface{}) error {
	err := yaml.Unmarshal(data, element)
	if err != nil {
		return err
	}
	return nil
}
