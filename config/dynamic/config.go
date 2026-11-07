package dynamic

import (
	"bytes"
	"mokapi/sortedmap"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/go-sprout/sprout/sprigin"
)

type Event int

const (
	Create Event = iota + 1
	Update
	Delete
	Chmod
)

var EventText = map[Event]string{
	Create: "Create",
	Update: "Update",
	Delete: "Delete",
	Chmod:  "Chmod",
}

type ConfigListener func(event ConfigEvent)

type Validator interface {
	Validate() error
}

type SourceType int

const (
	SourceMain SourceType = iota
	SourceReference
)

type Config struct {
	Info       ConfigInfo
	Raw        []byte
	Data       interface{}
	Refs       Refs
	Listeners  Listeners
	Scope      Scope
	SourceType SourceType
	resolving  map[string]bool
}

type Refs struct {
	refs map[string]*Config
	m    sync.Mutex
}

type Listeners struct {
	list *sortedmap.LinkedHashMap[string, ConfigListener]
	m    sync.Mutex
}

type ConfigEvent struct {
	Name   string
	Config *Config
	Event  Event
	// Checksum and Time are a snapshot of Config.Info taken when the event was created.
	// Listeners run concurrently with the next update of Config.Info, so they must
	// not read these values from Config.
	Checksum []byte
	Time     time.Time
}

func AddRef(parent, ref *Config) {
	if parent.Info.Key() == ref.Info.Key() {
		return
	}

	added := parent.Refs.Add(ref)
	if !added {
		return
	}
	if parent.Info.Url == nil {
		return
	}
	checksum := ref.Info.Checksum
	ref.Listeners.Add(parent.Info.Url.String(), func(e ConfigEvent) {
		// event Create is used for reading first time
		if bytes.Equal(e.Checksum, checksum) && e.Event == Create {
			return
		}

		parent.Listeners.Invoke(ConfigEvent{Event: Update, Config: parent, Name: parent.Info.Path(), Time: e.Time})
		if e.Event == Delete {
			parent.Refs.Remove(ref)
		}
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

func (l *Listeners) Remove(key string) {
	l.m.Lock()
	defer l.m.Unlock()

	if l.list == nil {
		return
	}
	l.list.Del(key)
}

func (l *Listeners) Invoke(e ConfigEvent) {
	for _, fn := range l.List() {
		fn(e)
	}
}

// List returns a snapshot of the listeners in insertion order. Callers invoke them
// without holding the lock, because a listener may add or invoke other listeners.
func (l *Listeners) List() []ConfigListener {
	l.m.Lock()
	defer l.m.Unlock()

	if l.list == nil {
		return nil
	}
	var fns []ConfigListener
	for it := l.list.Iter(); it.Next(); {
		fns = append(fns, it.Value())
	}
	return fns
}

func Wrap(i ConfigInfo, c *Config) {
	inner := c.Info
	i.Checksum = inner.Checksum
	c.Info = i
	c.Info.inner = &inner

}

func Validate(c *Config) error {
	if c.Data == nil {
		return nil
	}
	if v, ok := c.Data.(Validator); ok {
		return v.Validate()
	}
	return nil
}

func (r *Refs) List(recursive bool) []*Config {
	m := 20
	if !recursive {
		m = 1
	}
	return r.list(m)
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

func (r *Refs) Remove(ref *Config) {
	r.m.Lock()
	defer r.m.Unlock()

	if r.refs == nil {
		return
	}

	key := ref.Info.Path()
	delete(r.refs, key)
}

func renderTemplate(b []byte) ([]byte, error) {
	content := string(b)

	funcMap := sprigin.FuncMap()
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

func (e Event) String() string {
	return EventText[e]
}

func (c *Config) EnterRef(ref string) bool {
	if c.resolving == nil {
		c.resolving = map[string]bool{}
	}

	if c.resolving[ref] {
		return false
	}
	c.resolving[ref] = true
	return true
}

func (c *Config) LeaveRef(ref string) {
	delete(c.resolving, ref)
}
