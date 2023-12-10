package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/script"
	"mokapi/sortedmap"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"text/template"
)

type ConfigListener func(*Config)

type Validator interface {
	Validate() error
}

type Config struct {
	Info      ConfigInfo
	Raw       []byte
	Data      interface{}
	listeners *sortedmap.LinkedHashMap[string, ConfigListener]
	Checksum  []byte

	m sync.Mutex
}

func NewConfig(info ConfigInfo, opts ...ConfigOptions) *Config {
	f := &Config{Info: info, listeners: &sortedmap.LinkedHashMap[string, ConfigListener]{}}
	f.Options(opts...)
	return f
}

func Wrap(i ConfigInfo, c *Config) {
	inner := c.Info
	c.Info = i
	c.Info.inner = &inner

}

func (c *Config) Options(opts ...ConfigOptions) {
	for _, opt := range opts {
		opt(c, c.Data == nil)
	}
}

func (c *Config) Changed() {
	if c.listeners == nil {
		return
	}
	for it := c.listeners.Iter(); it.Next(); {
		it.Value()(c)
	}
}

func (c *Config) AddListener(key string, l ConfigListener) {
	if c.listeners == nil {
		c.listeners = &sortedmap.LinkedHashMap[string, ConfigListener]{}
	}
	if _, found := c.listeners.Get(key); !found {
		c.listeners.Set(key, l)
	}
}

func (c *Config) Parse(r Reader) error {
	c.m.Lock()
	defer c.m.Unlock()

	path := c.Info.Url.Path
	if len(c.Info.Url.Opaque) > 0 {
		path = c.Info.Url.Opaque
	}
	_, name := filepath.Split(path)

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
		err := yaml.Unmarshal(data, c)
		if err != nil {
			return err
		}
	case ".json":
		err := json.Unmarshal(data, c)
		if err != nil {
			return err
		}
	case ".lua", ".js":
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

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	data := make(map[string]string)
	_ = unmarshal(data)

	for _, ct := range configTypes {
		if _, ok := data[ct.header]; ok {
			return c.unmarshal(unmarshal, ct)
		} else if c.Data != nil {
			v := reflect.ValueOf(c.Data)
			if v.Kind() != reflect.Ptr {
				continue
			}
			if v.Elem().Type() == ct.configType {
				return c.unmarshal(unmarshal, ct)
			}
		}
	}

	if c.Data == nil {
		return nil
	}

	err := unmarshal(c.Data)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) unmarshal(fn func(interface{}) error, ct configType) error {
	if c.Data != nil {
		i := reflect.New(ct.configType).Interface()
		err := fn(i)
		v := reflect.ValueOf(c.Data).Elem()
		v.Set(reflect.ValueOf(i).Elem())
		return err
	} else {
		c.Data = reflect.New(ct.configType).Interface()
		return fn(c.Data)
	}
}

func (c *Config) UnmarshalJSON(b []byte) error {
	data := make(map[string]string)
	_ = json.Unmarshal(b, &data)

	for _, ct := range configTypes {
		if _, ok := data[ct.header]; ok {
			c.Data = reflect.New(ct.configType).Interface()
			return json.Unmarshal(b, c.Data)
		}
	}

	if c.Data == nil {
		return nil
	}

	err := json.Unmarshal(b, c.Data)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) Validate() error {
	if v, ok := c.Data.(Validator); ok {
		return v.Validate()
	}
	return nil
}

func renderTemplate(b []byte) ([]byte, error) {
	content := string(b)

	funcMap := sprig.TxtFuncMap()
	funcMap["extractUsername"] = extractUsername
	tmpl := template.New("").Funcs(funcMap)

	tmpl, err := tmpl.Parse(content)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, false)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func extractUsername(s string) string {
	slice := strings.Split(s, "\\")
	return slice[len(slice)-1]
}
