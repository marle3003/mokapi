package dynamic

import (
	"fmt"
	"mokapi/version"
	"net/url"
	"path/filepath"
	"reflect"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type dynamicObject struct {
	data any
}

type Reader interface {
	// Read reads data from the given url. If v is not nil then
	// read uses the type of v as a hint for the first parse. v itself is not filled,
	// the parsed data is available by Config.Data().
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
	data, err := parse(c)
	if err != nil {
		// clears data and scope, anchors of the previous parse must not resolve anymore
		c.data.Store(nil)
		return err
	}

	// The hook fills the fresh object on a view of the config and the result is published
	// as a whole afterwards, so readers that still hold the previous object never see it change.
	view := c.viewOf(data)
	view.Scope().SetName(c.Info.Path())
	if p, ok := data.(Parser); ok {
		if err = p.Parse(view, r); err != nil {
			return errors.Wrapf(err, "parsing file %v", c.Info.Path())
		}
	}
	c.data.Store(view.data.Load())

	return nil
}

func parse(c *Config) (any, error) {
	name := getFileName(c)

	b := c.Raw
	var err error
	if filepath.Ext(name) == ".tmpl" {
		b, err = renderTemplate(b)
		if err != nil {
			return nil, fmt.Errorf("unable to render template %v: %v", c.Info.Path(), err)
		}
		name = name[0 : len(name)-len(filepath.Ext(name))]
	}

	// the previous data is only a type hint, it is never filled
	result := c.Data()
	switch filepath.Ext(name) {
	case ".yml", ".yaml":
		result, err = parseYaml(b, result)
	case ".json":
		result, err = parseJson(b, result)
	case ".lua", ".js", ".cjs", ".mjs", ".ts", ".ldif":
		result = string(b)
	default:
		// try parse from JSON and YAML
		var v any
		v, err = parseJson(b, result)
		if err == nil {
			c.Info.ContentType = "application/json"
			return v, nil
		}
		v, err = parseYaml(b, result)
		if v != nil && err == nil {
			c.Info.ContentType = "application/yaml"
			return v, nil
		}
		err = nil
		result = string(b)
	}

	return result, err
}

func parseJson(b []byte, v any) (any, error) {
	d := &dynamicObject{data: v}
	err := UnmarshalJSON(b, d)
	if err != nil {
		return nil, FormatError(b, err)
	}
	return d.data, nil
}

func parseYaml(b []byte, v any) (any, error) {
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
			return FormatError(b, err)
		}
		return nil
	}

	if d.data == nil {
		return nil
	}

	if err := d.newTarget(); err != nil {
		return err
	}

	err := UnmarshalJSON(b, &d.data)
	if err != nil {
		return err
	}

	return nil
}

func (d *dynamicObject) UnmarshalYAML(unmarshal func(any) error) error {
	data := make(map[string]string)
	_ = unmarshal(data)

	if ct := getConfigType(data); ct != nil {
		d.data = reflect.New(ct.configType).Interface()
		return unmarshal(d.data)
	}

	if d.data == nil {
		return nil
	}

	if err := d.newTarget(); err != nil {
		return err
	}

	err := unmarshal(d.data)
	if err != nil {
		return err
	}

	return nil
}

// newTarget replaces the type hint with a new, empty value of the same type,
// so that the hint itself is never modified.
func (d *dynamicObject) newTarget() error {
	rt := reflect.TypeOf(d.data)
	if rt.Kind() != reflect.Pointer {
		return fmt.Errorf("dynamic data must have a pointer to a struct")
	}
	d.data = reflect.New(rt.Elem()).Interface()
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

func getConfigType(data map[string]string) *configType {
	configTypesMu.RLock()
	defer configTypesMu.RUnlock()
	for _, ct := range configTypes {
		if s, ok := data[ct.header]; ok {
			if ct.checkVersion == nil || ct.checkVersion(version.New(s)) {
				return ct
			}
		}
	}
	return nil
}

type EmptyReader struct{}

func (e *EmptyReader) Read(u *url.URL, _ any) (*Config, error) {
	return nil, fmt.Errorf("not found %v", u.String())
}
