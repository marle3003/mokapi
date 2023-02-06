package modules

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
	"os"
	"sync"
	"time"
)

type Mokapi struct {
	host common.Host
	rt   *goja.Runtime
	// goja scripts are not thread-safe
	m sync.Mutex
}

func NewMokapi(host common.Host, rt *goja.Runtime) interface{} {
	return &Mokapi{host: host, rt: rt}
}

func (*Mokapi) Sleep(milliseconds float64) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}

func (m *Mokapi) Every(every string, do func(), args goja.Value) (int, error) {
	options := common.NewJobOptions()

	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(m.rt)
		for _, k := range params.Keys() {
			switch k {
			case "tags":
				options.Tags = make(map[string]string)
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				tagsO := tagsV.ToObject(m.rt)
				for _, key := range tagsO.Keys() {
					options.Tags[key] = tagsO.Get(key).String()
				}
			case "times":
				tagsV := params.Get(k)
				options.Times = int(tagsV.ToInteger())
			}
		}
	}

	f := func() {
		m.m.Lock()
		defer m.m.Unlock()
		do()
	}

	return m.host.Every(every, f, options)
}

func (m *Mokapi) Cron(expr string, do func(), args goja.Value) (int, error) {
	options := common.NewJobOptions()

	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(m.rt)
		for _, k := range params.Keys() {
			switch k {
			case "tags":
				options.Tags = make(map[string]string)
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				tagsO := tagsV.ToObject(m.rt)
				for _, key := range tagsO.Keys() {
					options.Tags[key] = tagsO.Get(key).String()
				}
			case "times":
				tagsV := params.Get(k)
				options.Times = int(tagsV.ToInteger())
			}
		}
	}

	f := func() {
		m.m.Lock()
		defer m.m.Unlock()
		do()
	}

	return m.host.Cron(expr, f, options)
}

func (m *Mokapi) On(event string, do goja.Value, args goja.Value) {
	tags := make(map[string]string)

	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(m.rt)
		for _, k := range params.Keys() {
			switch k {
			case "tags":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				tagsO := tagsV.ToObject(m.rt)
				for _, key := range tagsO.Keys() {
					tags[key] = tagsO.Get(key).String()
				}
			}
		}
	}

	f := func(args ...interface{}) (bool, error) {
		m.m.Lock()
		defer m.m.Unlock()

		call, _ := goja.AssertFunction(do)
		var params []goja.Value
		for _, v := range args {
			params = append(params, m.rt.ToValue(v))
		}
		r, err := call(goja.Undefined(), params...)
		if err != nil {
			return false, err
		}
		return r.ToBoolean(), nil
	}

	m.host.On(event, f, tags)
}

func (m *Mokapi) Env(name string) string {
	return os.Getenv(name)
}

func (m *Mokapi) Open(file string) (string, error) {
	_, s, err := m.host.OpenFile(file, "")
	if err != nil {
		return "", err
	}
	return s, nil
}

type DateArg struct {
	Layout    string `json:"layout"`
	Timestamp int64  `json:"timestamp"`
}

func (m *Mokapi) Date(args DateArg) string {
	if len(args.Layout) == 0 {
		args.Layout = time.RFC3339
	}
	var t time.Time
	if args.Timestamp == 0 {
		t = time.Now().UTC()
	} else {
		t = time.UnixMilli(args.Timestamp).UTC()
	}
	return t.Format(args.Layout)
}
