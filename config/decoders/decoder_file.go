package decoders

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type FileDecoder struct {
	filenames []string
}

func (f *FileDecoder) Decode(flags map[string][]string, element interface{}) error {
	if len(f.filenames) == 0 {
		if val, ok := flags["configfile"]; ok {
			f.filenames = val
		}
	}

	if len(f.filenames) == 0 {
		return nil
	}

	for _, filename := range f.filenames {
		data, err := loadFile(filename)
		if err != nil {
			return err
		}

		err = parseYml(data, element)
		if err != nil {
			return errors.Wrapf(err, "parsing YAML file %s", filename)
		}
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
