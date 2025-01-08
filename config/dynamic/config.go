package dynamic

import (
	"bytes"
	"fmt"
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
	Scope     *Scope
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

	if ref.Scope == nil {
		ref.OpenScope("")
	}
	if parent.Scope != nil {
		ref.Scope.dynamic = parent.Scope.dynamic
	}
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

type Scope struct {
	Name string

	stack   []map[string]interface{}
	dynamic *Scope
}

func NewScope(name string) *Scope {
	return &Scope{Name: name, stack: []map[string]interface{}{
		{},
	}}
}

func (s *Scope) Get(name string) (interface{}, error) {
	if s == nil {
		return nil, fmt.Errorf("anchor '%s' not found: no scope present", name)
	}

	if len(s.stack) != 0 {
		lexical := s.stack[len(s.stack)-1]
		if v, ok := lexical[name]; ok {
			return v, nil
		}
	}

	return nil, fmt.Errorf("anchor '%s' not found in scope '%s", name, s.Name)
}

func (s *Scope) Set(name string, value interface{}) error {
	if s == nil || len(s.stack) == 0 {
		return fmt.Errorf("set anchor '%s' failed: no scope present", name)
	}

	lexical := s.stack[len(s.stack)-1]
	if _, ok := lexical[name]; ok {
		return fmt.Errorf("anchor '%s' already defined in scope '%s'", name, s.Name)
	}
	lexical[name] = value
	return nil
}

func (s *Scope) GetDynamic(name string) (interface{}, error) {
	return s.dynamic.Get(name)
}

func (s *Scope) SetDynamic(name string, value interface{}) error {
	return s.dynamic.Set(name, value)
}

func (s *Scope) openScope(name string) {
	s.stack = append(s.stack, map[string]interface{}{})
}

func (s *Scope) close() {
	s.stack = s.stack[:len(s.stack)-1]
}

func (s *Scope) IsEmpty() bool {
	return len(s.stack) == 0
}

func (c *Config) OpenScope(name string) {
	if c.Scope == nil {
		c.Scope = NewScope(name)
	} else {
		c.Scope.openScope(name)
	}
}

func (c *Config) Close() {
	if c.Scope != nil {
		if c.Scope.IsEmpty() {
			c.Scope = nil
		} else {
			c.Scope.close()
		}
	}
}
