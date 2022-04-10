package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/script"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"text/template"
)

var UnknownFile = errors.New("unknown file")

type Config struct {
	Url          *url.URL
	Raw          []byte
	Data         interface{}
	Listeners    []func(*Config)
	ProviderName string

	parseMode string
	m         sync.Mutex
}

func NewConfig(u *url.URL, opts ...ConfigOptions) *Config {
	f := &Config{Url: u}
	for _, opt := range opts {
		opt(f, true)
	}
	return f
}

func (f *Config) Options(opts ...ConfigOptions) {
	for _, opt := range opts {
		opt(f, f.Data == nil)
	}
}

func (f *Config) Changed() {
	for _, l := range f.Listeners {
		l(f)
	}
}

type ConfigOptions func(config *Config, init bool)

func WithListener(f func(file *Config)) ConfigOptions {
	return func(file *Config, init bool) {
		file.Listeners = append(file.Listeners, f)
	}
}

func WithData(data interface{}) ConfigOptions {
	return func(file *Config, init bool) {
		if !init {
			return
		}
		file.Data = data
	}
}

func WithParent(parent *Config) ConfigOptions {
	return func(file *Config, init bool) {
		file.Listeners = append(file.Listeners, func(_ *Config) {
			parent.Changed()
		})
	}
}

func AllowParsingAny() ConfigOptions {
	return func(file *Config, init bool) {
		if !init {
			return
		}
		file.parseMode = "any"
	}
}

func AsPlaintext() ConfigOptions {
	return func(file *Config, init bool) {
		if !init {
			return
		}
		file.parseMode = "plaintext"
	}
}

func (f *Config) Parse(r Reader) error {
	if f.parseMode == "plaintext" {
		f.Data = string(f.Raw)
		return nil
	}

	f.m.Lock()
	defer f.m.Unlock()

	path := f.Url.Path
	if len(f.Url.Opaque) > 0 {
		path = f.Url.Opaque
	}
	_, name := filepath.Split(path)

	var data []byte
	if filepath.Ext(name) == ".tmpl" {
		var err error
		data, err = renderTemplate(f.Raw)
		if err != nil {
			return fmt.Errorf("unable to render template %v: %v", f.Url, err)
		}
		name = name[0 : len(name)-len(filepath.Ext(name))]
	} else {
		data = f.Raw
	}

	switch filepath.Ext(name) {
	case ".yml", ".yaml":
		err := yaml.Unmarshal(data, f)
		if err != nil {
			f.Data = string(data)
		}
	case ".json":
		err := json.Unmarshal(data, f)
		if err != nil {
			f.Data = string(data)
		}
	case ".lua", ".js":
		if f.Data == nil {
			f.Data = script.New(name, data)
		} else {
			script := f.Data.(*script.Script)
			script.Code = string(data)
		}
	default:
		f.Data = string(data)
	}

	if p, ok := f.Data.(Parser); ok {
		err := p.Parse(f, r)
		if err != nil {
			return errors.Wrapf(err, "parsing file %v", f.Url)
		}
	}

	return nil
}

func (f *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	data := make(map[string]string)
	_ = unmarshal(data)

	for _, ct := range configTypes {
		if _, ok := data[ct.header]; ok {
			return f.unmarshal(unmarshal, ct)
		}
	}

	if f.Data == nil {
		if f.parseMode == "any" {
			f.Data = make(map[string]interface{})
		} else {
			return nil
		}
	}

	err := unmarshal(f.Data)
	if err != nil {
		return err
	}

	return nil
}

func (f *Config) unmarshal(fn func(interface{}) error, ct configType) error {
	if f.Data != nil {
		i := reflect.New(ct.configType).Interface()
		err := fn(i)
		v := reflect.ValueOf(f.Data).Elem()
		v.Set(reflect.ValueOf(i).Elem())
		return err
	} else {
		f.Data = reflect.New(ct.configType).Interface()
		return fn(f.Data)
	}
}

func (f *Config) UnmarshalJSON(b []byte) error {
	data := make(map[string]string)
	_ = json.Unmarshal(b, &data)

	for _, ct := range configTypes {
		if _, ok := data[ct.header]; ok {
			f.Data = reflect.New(ct.configType).Interface()
			return json.Unmarshal(b, f.Data)
		}
	}

	if f.Data == nil {
		if f.parseMode == "any" {
			f.Data = make(map[string]interface{})
		} else {
			return nil
		}
	}

	err := json.Unmarshal(b, f.Data)
	if err != nil {
		return err
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
