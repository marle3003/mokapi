package dynamic

import (
	"bytes"
	"github.com/Masterminds/sprig"
	"mokapi/sortedmap"
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
	Refs      Refs
	Listeners Listeners
}

type Refs struct {
	refs map[string]*Config
	m    sync.Mutex
}

type Listeners struct {
	list *sortedmap.LinkedHashMap[string, ConfigListener]
	m    sync.Mutex
}

func AddRef(parent *Config, ref *Config) {
	parent.Refs.Add(ref)
	ref.Listeners.Add(parent.Info.Url.String(), func(config *Config) {
		parent.Listeners.Invoke(parent)
	})
}

func (l *Listeners) Add(key string, fn ConfigListener) {
	l.m.Lock()
	l.m.Unlock()

	if l.list == nil {
		l.list = &sortedmap.LinkedHashMap[string, ConfigListener]{}
	}
	if _, found := l.list.Get(key); !found {
		l.list.Set(key, fn)
	}
}

func (l *Listeners) Invoke(c *Config) {
	if l.list == nil {
		return
	}
	for it := l.list.Iter(); it.Next(); {
		it.Value()(c)
	}
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

func Validate(c *Config) error {
	if v, ok := c.Data.(Validator); ok {
		return v.Validate()
	}
	return nil
}

func (r *Refs) List() []*Config {
	var refs []*Config
	for _, v := range r.refs {
		refs = append(refs, v)
		refs = append(refs, v.Refs.List()...)
	}
	return refs
}

func (r *Refs) Add(ref *Config) {
	r.m.Lock()
	defer r.m.Unlock()

	if r.refs == nil {
		r.refs = make(map[string]*Config)
	}

	key := ref.Info.Path()
	if _, ok := r.refs[key]; ok {
		return
	}
	r.refs[key] = ref
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
