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
	Scope     Scope
}

type Refs struct {
	refs map[string]*Config
	m    sync.Mutex
}

type Listeners struct {
	list *sortedmap.LinkedHashMap[string, ConfigListener]
	m    sync.Mutex
}

func AddRef(parent, ref *Config) {
	if parent.Info.Key() == ref.Info.Key() {
		return
	}

	added := parent.Refs.Add(ref)
	if !added {
		return
	}
	ref.Listeners.Add(parent.Info.Url.String(), func(config *Config) {
		parent.Info.Time = ref.Info.Time
		parent.Listeners.Invoke(parent)
	})
}

func (l *Listeners) Add(key string, fn ConfigListener) {
	l.m.Lock()
	defer l.m.Unlock()

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
	i.Checksum = inner.Checksum
	c.Info = i
	c.Info.inner = &inner

}

func Validate(c *Config) error {
	if v, ok := c.Data.(Validator); ok {
		return v.Validate()
	}
	return nil
}

func (r *Refs) List(recursive bool) []*Config {
	max := 20
	if !recursive {
		max = 1
	}
	return r.list(max)
}

func (r *Refs) list(max int) []*Config {
	if max == 0 {
		return nil
	}

	var refs []*Config
	for _, v := range r.refs {
		refs = append(refs, v)
		refs = append(refs, v.Refs.list(max-1)...)
	}
	return refs
}

func (r *Refs) Add(ref *Config) bool {
	r.m.Lock()
	defer r.m.Unlock()

	if r.refs == nil {
		r.refs = make(map[string]*Config)
	}

	key := ref.Info.Path()
	if _, ok := r.refs[key]; ok {
		return false
	}
	r.refs[key] = ref
	return true
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

func (c *Config) OpenScope(name string) {
	c.Scope.Open(name)
}

func (c *Config) CloseScope() {
	c.Scope.Close()
}
