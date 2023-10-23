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
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"text/template"
)

var UnknownFile = errors.New("unknown file")

type ConfigListener func(*Config)

type Validator interface {
	Validate() error
}

type ConfigInfo struct {
	Provider string
	Url      *url.URL
	Parent   *ConfigInfo
}

func (ci *ConfigInfo) Path() string {
	if ci.Parent != nil {
		return ci.Parent.Path()
	}
	if len(ci.Url.Opaque) > 0 {
		return ci.Url.Opaque
	}
	u := ci.Url
	path, _ := url.PathUnescape(ci.Url.Path)
	query, _ := url.QueryUnescape(ci.Url.RawQuery)
	var sb strings.Builder
	if len(u.Scheme) > 0 {
		sb.WriteString(u.Scheme + ":")
	}
	if len(u.Scheme) > 0 || len(u.Host) > 0 {
		sb.WriteString("//")
	}
	if len(u.Host) > 0 {
		sb.WriteString(u.Host)
	}
	sb.WriteString(path)
	if len(query) > 0 {
		sb.WriteString("?" + query)
	}
	return sb.String()
}

type Config struct {
	Info      ConfigInfo
	Raw       []byte
	Data      interface{}
	listeners *sortedmap.LinkedHashMap[string, ConfigListener]
	Checksum  []byte
	Key       string

	parseMode string
	m         sync.Mutex
}

func NewConfig(u *url.URL, opts ...ConfigOptions) *Config {
	f := &Config{Info: ConfigInfo{Url: u}, listeners: &sortedmap.LinkedHashMap[string, ConfigListener]{}}
	f.Options(opts...)
	return f
}

func (f *Config) Options(opts ...ConfigOptions) {
	for _, opt := range opts {
		opt(f, f.Data == nil)
	}
}

func (f *Config) Changed() {
	if f.listeners == nil {
		return
	}
	for it := f.listeners.Iter(); it.Next(); {
		it.Value()(f)
	}
}

type ConfigOptions func(config *Config, init bool)

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
		file.AddListener(parent.Info.Url.String(), func(_ *Config) {
			parent.Changed()
		})
	}
}

func WithListener(key string, l ConfigListener) ConfigOptions {
	return func(file *Config, init bool) {
		file.AddListener(key, l)
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

func (f *Config) AddListener(key string, l ConfigListener) {
	if f.listeners == nil {
		f.listeners = &sortedmap.LinkedHashMap[string, ConfigListener]{}
	}
	if _, found := f.listeners.Get(key); !found {
		f.listeners.Set(key, l)
	}
}

func (f *Config) Parse(r Reader) error {
	if f.parseMode == "plaintext" {
		f.Data = string(f.Raw)
		return nil
	}

	f.m.Lock()
	defer f.m.Unlock()

	path := f.Info.Url.Path
	if len(f.Info.Url.Opaque) > 0 {
		path = f.Info.Url.Opaque
	}
	_, name := filepath.Split(path)

	var data []byte
	if filepath.Ext(name) == ".tmpl" {
		var err error
		data, err = renderTemplate(f.Raw)
		if err != nil {
			return fmt.Errorf("unable to render template %v: %v", f.Info.Path(), err)
		}
		name = name[0 : len(name)-len(filepath.Ext(name))]
	} else {
		data = f.Raw
	}

	switch filepath.Ext(name) {
	case ".yml", ".yaml":
		err := yaml.Unmarshal(data, f)
		if err != nil {
			return err
		}
	case ".json":
		err := json.Unmarshal(data, f)
		if err != nil {
			return err
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
			return errors.Wrapf(err, "parsing file %v", f.Info.Path())
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
		} else if f.Data != nil {
			v := reflect.ValueOf(f.Data)
			if v.Kind() != reflect.Ptr {
				continue
			}
			if v.Elem().Type() == ct.configType {
				return f.unmarshal(unmarshal, ct)
			}
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

func (f *Config) Validate() error {
	if v, ok := f.Data.(Validator); ok {
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
