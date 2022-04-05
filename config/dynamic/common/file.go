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

type File struct {
	Url *url.URL

	Data interface{}

	Listeners []func(*File)

	parseMode string
	m         sync.Mutex
}

func NewFile(u *url.URL, opts ...FileOptions) *File {
	f := &File{Url: u}
	for _, opt := range opts {
		opt(f, true)
	}
	return f
}

func (f *File) Options(opts ...FileOptions) {
	for _, opt := range opts {
		opt(f, f.Data == nil)
	}
}

func (f *File) Changed() {
	for _, l := range f.Listeners {
		l(f)
	}
}

type FileOptions func(file *File, init bool)

func WithListener(f func(file *File)) FileOptions {
	return func(file *File, init bool) {
		file.Listeners = append(file.Listeners, f)
	}
}

func WithData(data interface{}) FileOptions {
	return func(file *File, init bool) {
		if !init {
			return
		}
		file.Data = data
	}
}

func WithParent(parent *File) FileOptions {
	return func(file *File, init bool) {
		file.Listeners = append(file.Listeners, func(_ *File) {
			parent.Changed()
		})
	}
}

func AllowParsingAny() FileOptions {
	return func(file *File, init bool) {
		if !init {
			return
		}
		file.parseMode = "any"
	}
}

func AsPlaintext() FileOptions {
	return func(file *File, init bool) {
		if !init {
			return
		}
		file.parseMode = "plaintext"
	}
}

func (f *File) Parse(c *Config, r Reader) error {
	f.m.Lock()
	defer f.m.Unlock()

	path := f.Url.Path
	if len(f.Url.Opaque) > 0 {
		path = f.Url.Opaque
	}
	_, name := filepath.Split(path)

	if filepath.Ext(name) == ".tmpl" {
		var err error
		c.Data, err = renderTemplate(c.Data)
		if err != nil {
			return fmt.Errorf("unable to render template %v: %v", c.Url, err)
		}
		name = name[0 : len(name)-len(filepath.Ext(name))]
	}

	if f.parseMode == "plaintext" {
		f.Data = string(c.Data)
		return nil
	}

	switch filepath.Ext(name) {
	case ".yml", ".yaml":
		err := yaml.Unmarshal(c.Data, f)
		if err != nil {
			f.Data = string(c.Data)
		}
	case ".json":
		err := json.Unmarshal(c.Data, f)
		if err != nil {
			f.Data = string(c.Data)
		}
	case ".lua", ".js":
		if f.Data == nil {
			f.Data = script.New(name, c.Data)
		} else {
			script := f.Data.(*script.Script)
			script.Code = string(c.Data)
		}
	default:
		f.Data = string(c.Data)
	}

	if p, ok := f.Data.(Parser); ok {
		err := p.Parse(f, r)
		if err != nil {
			return errors.Wrapf(err, "parsing file %v", f.Url)
		}
	}

	return nil
}

func (f *File) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

func (f *File) unmarshal(fn func(interface{}) error, ct configType) error {
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

func (f *File) UnmarshalJSON(b []byte) error {
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
