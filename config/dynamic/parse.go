package dynamic

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/script"
	"net/url"
	"path/filepath"
	"reflect"
)

type dynamicObject struct {
	data any
}

type Reader interface {
	// Read reads data from the given url. If v is not nil
	// Read tries to store the data in the value pointed to by v
	Read(u *url.URL, v any) (*Config, error)
}

type Parser interface {
	Parse(config *Config, reader Reader) error
}

func Parse(c *Config, r Reader) error {
	name := getFileName(c)

	var data []byte
	if filepath.Ext(name) == ".tmpl" {
		var err error
		data, err = renderTemplate(c.Raw)
		if err != nil {
			return fmt.Errorf("unable to render template %v: %v", c.Info.Path(), err)
		}
		name = name[0 : len(name)-len(filepath.Ext(name))]
	} else {
		data = c.Raw
	}

	switch filepath.Ext(name) {
	case ".yml", ".yaml":
		d := &dynamicObject{data: c.Data}
		err := yaml.Unmarshal(data, d)
		if err != nil {
			return err
		}
		c.Data = d.data
	case ".json":
		d := &dynamicObject{data: c.Data}
		err := UnmarshalJSON(data, d)
		if err != nil {
			return err
		}
		c.Data = d.data
	case ".lua", ".js", ".ts":
		if c.Data == nil {
			c.Data = script.New(name, data)
		} else {
			script := c.Data.(*script.Script)
			script.Code = string(data)
		}
	default:
		c.Data = string(data)
	}

	if p, ok := c.Data.(Parser); ok {
		err := p.Parse(c, r)
		if err != nil {
			return errors.Wrapf(err, "parsing file %v", c.Info.Path())
		}
	}

	return nil
}

func (d *dynamicObject) UnmarshalJSON(b []byte) error {
	data := make(map[string]string)
	_ = UnmarshalJSON(b, &data)

	for _, ct := range configTypes {
		if _, ok := data[ct.header]; ok {
			d.data = reflect.New(ct.configType).Interface()
			err := UnmarshalJSON(b, d.data)
			if err != nil {
				return formatError(b, err)
			}
			return nil
		}
	}

	if d.data == nil {
		return nil
	}

	err := UnmarshalJSON(b, d.data)
	if err != nil {
		return formatError(b, err)
	}

	return nil
}

func (d *dynamicObject) UnmarshalYAML(unmarshal func(interface{}) error) error {
	data := make(map[string]string)
	_ = unmarshal(data)

	for _, ct := range configTypes {
		if _, ok := data[ct.header]; ok {
			d.data = reflect.New(ct.configType).Interface()
			return unmarshal(d.data)
		}
	}

	if d.data == nil {
		return nil
	}

	err := unmarshal(d.data)
	if err != nil {
		return err
	}

	return nil
}

func getFileName(c *Config) string {
	info := c.Info
	for {
		inner := info.Inner()
		if inner == nil {
			break
		}
		info = *inner
	}

	path := info.Url.Path
	if len(info.Url.Opaque) > 0 {
		path = info.Url.Opaque
	}
	_, name := filepath.Split(path)
	return name
}

func formatError(input []byte, err error) error {
	var structErr *StructuralError
	if !errors.As(err, &structErr) {
		return err
	}

	newLine := byte(0x0A)
	offset := int(structErr.Offset)

	if offset > len(input) || offset < 0 {
		return err
	}

	line := 1
	column := 0
	for i, b := range input {
		if i == offset {
			break
		}
		if b == newLine {
			line++
			column = 0
		} else {
			column++
		}
	}

	return fmt.Errorf("%w at line %d, column %d", err, line, column)
}
