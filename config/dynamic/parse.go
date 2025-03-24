package dynamic

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/script"
	"mokapi/version"
	"net/url"
	"path/filepath"
	"reflect"
)

type dynamicObject struct {
	data any
}

type Reader interface {
	// Read reads data from the given url. If v is not nil then
	// read tries to store the data in the value pointed to by v
	Read(u *url.URL, v any) (*Config, error)
}

type Context struct {
	Config *Config
	Reader Reader
}

type Parser interface {
	Parse(config *Config, reader Reader) error
}

func Parse(c *Config, r Reader) error {
	var err error
	c.Data, err = parse(c)
	if err != nil {
		return err
	}

	c.Scope.SetName(c.Info.Path())

	if p, ok := c.Data.(Parser); ok {
		err = p.Parse(c, r)
		if err != nil {
			return errors.Wrapf(err, "parsing file %v", c.Info.Path())
		}
	}

	return nil
}

func parse(c *Config) (interface{}, error) {
	name := getFileName(c)
	reset(c)

	b := c.Raw
	var err error
	if filepath.Ext(name) == ".tmpl" {
		b, err = renderTemplate(b)
		if err != nil {
			return nil, fmt.Errorf("unable to render template %v: %v", c.Info.Path(), err)
		}
		name = name[0 : len(name)-len(filepath.Ext(name))]
	}

	result := c.Data
	switch filepath.Ext(name) {
	case ".yml", ".yaml":
		result, err = parseYaml(b, result)
	case ".json":
		result, err = parseJson(b, result)
	case ".lua", ".js", ".ts":
		if result == nil {
			result = script.New(name, b)
		} else {
			s := result.(*script.Script)
			s.Code = string(b)
			s.Filename = name
		}
	default:
		if result != nil {
			rv := reflect.ValueOf(result)
			if rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Struct || rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array || rv.Kind() == reflect.Map {
				// try parse from json and yaml
				var v interface{}
				v, err = parseJson(b, result)
				if err == nil {
					return v, nil
				}
				v, err = parseYaml(b, result)
				if err == nil {
					return v, nil
				}
				err = nil
			}
		}
		result = string(b)
	}

	return result, err
}

func parseJson(b []byte, v any) (interface{}, error) {
	d := &dynamicObject{data: v}
	err := UnmarshalJSON(b, d)
	if err != nil {
		return nil, err
	}
	return d.data, nil
}

func parseYaml(b []byte, v any) (interface{}, error) {
	d := &dynamicObject{data: v}
	err := yaml.Unmarshal(b, d)
	if err != nil {
		return nil, err
	}
	return d.data, nil
}

func (d *dynamicObject) UnmarshalJSON(b []byte) error {
	data := make(map[string]string)
	_ = UnmarshalJSON(b, &data)

	if ct := getConfigType(data); ct != nil {
		d.data = reflect.New(ct.configType).Interface()
		err := UnmarshalJSON(b, d.data)
		if err != nil {
			return formatError(b, err)
		}
		return nil
	}

	if d.data == nil {
		return nil
	}

	// resolve pointer of pointer for example: **schema.Schema
	rt := reflect.TypeOf(d.data)
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	rv := reflect.ValueOf(d.data)
	if rv.Elem().CanSet() {
		rv.Elem().Set(reflect.New(rt).Elem())
	} else {
		d.data = reflect.New(rt).Interface()
	}

	err := UnmarshalJSON(b, &d.data)
	if err != nil {
		return formatError(b, err)
	}

	return nil
}

func (d *dynamicObject) UnmarshalYAML(unmarshal func(interface{}) error) error {
	data := make(map[string]string)
	_ = unmarshal(data)

	if ct := getConfigType(data); ct != nil {
		d.data = reflect.New(ct.configType).Interface()
		return unmarshal(d.data)
	}

	if d.data == nil {
		return nil
	}

	// resolve pointer of pointer for example: **schema.Schema
	rt := reflect.TypeOf(d.data)
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	rv := reflect.ValueOf(d.data)
	if rv.Elem().CanSet() {
		rv.Elem().Set(reflect.New(rt).Elem())
	} else {
		d.data = reflect.New(rt).Interface()
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

	if info.Url == nil {
		return ""
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

func getConfigType(data map[string]string) *configType {
	for _, ct := range configTypes {
		if s, ok := data[ct.header]; ok {
			if ct.checkVersion(version.New(s)) {
				return ct
			}
		}
	}
	return nil
}

func reset(c *Config) {
	v := reflect.ValueOf(c.Data)
	if v.Kind() != reflect.Ptr || v.IsZero() {
		return
	}

	p := v.Elem()
	p.Set(reflect.Zero(p.Type()))
}

type EmptyReader struct{}

func (e *EmptyReader) Read(u *url.URL, v any) (*Config, error) {
	return nil, fmt.Errorf("not found %v", u.String())
}
