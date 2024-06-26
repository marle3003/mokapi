package js

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	"os"
	"time"
)

type mokapi struct {
	host common.Host
	rt   *goja.Runtime
}

func newMokapi(host common.Host, rt *goja.Runtime) interface{} {
	return &mokapi{host: host, rt: rt}
}

func (m *mokapi) Sleep(i interface{}) error {
	switch t := i.(type) {
	case int64:
		time.Sleep(time.Duration(t) * time.Millisecond)
	case string:
		d, err := time.ParseDuration(t)
		if err != nil {
			panic(m.rt.ToValue(err.Error()))
		}
		time.Sleep(d)
	}
	return nil
}

func (m *mokapi) Every(every string, do func(), args goja.Value) (int, error) {
	options := common.NewJobOptions()

	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(m.rt)
		for _, k := range params.Keys() {
			switch k {
			case "tags":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				tags := tagsV.ToObject(m.rt)
				for _, key := range tags.Keys() {
					options.Tags[key] = tags.Get(key).String()
				}
			case "times":
				times := params.Get(k)
				options.Times = int(times.ToInteger())
			case "skipImmediateFirstRun":
				skip := params.Get(k)
				options.SkipImmediateFirstRun = skip.ToBoolean()
			}
		}
	}

	f := func() {
		m.host.Lock()
		defer m.host.Unlock()
		do()
	}

	return m.host.Every(every, f, options)
}

func (m *mokapi) Cron(expr string, do func(), args goja.Value) (int, error) {
	options := common.NewJobOptions()

	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(m.rt)
		for _, k := range params.Keys() {
			switch k {
			case "tags":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				tags := tagsV.ToObject(m.rt)
				for _, key := range tags.Keys() {
					options.Tags[key] = tags.Get(key).String()
				}
			case "times":
				times := params.Get(k)
				options.Times = int(times.ToInteger())
			case "skipImmediateFirstRun":
				skip := params.Get(k)
				options.SkipImmediateFirstRun = skip.ToBoolean()
			}
		}
	}

	f := func() {
		m.host.Lock()
		defer m.host.Unlock()
		do()
	}

	return m.host.Cron(expr, f, options)
}

func (m *mokapi) On(event string, do goja.Value, args goja.Value) {
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
		m.host.Lock()
		defer m.host.Unlock()

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

func (m *mokapi) Env(name string) string {
	return os.Getenv(name)
}

func (m *mokapi) Open(file string) (string, error) {
	f, err := m.host.OpenFile(file, "")
	if err != nil {
		return "", err
	}
	return string(f.Raw), nil
}

type MarshalArg struct {
	Schema      *jsonSchema `json:"schema"`
	ContentType string      `json:"contentType"`
}

func (m *mokapi) Marshal(i interface{}, encoding *MarshalArg) string {
	ct := media.ContentType{}
	r := &schema.Ref{}
	if encoding != nil {
		ct = media.ParseContentType(encoding.ContentType)
		r.Value = toSchema(encoding.Schema)
	}
	if ct.IsEmpty() {
		ct = media.ParseContentType("application/json")
	}

	b, err := r.Marshal(i, ct)
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}
	return string(b)
}

type DateArg struct {
	Layout    string `json:"layout"`
	Timestamp int64  `json:"timestamp"`
}

func (m *mokapi) Date(args DateArg) string {
	var layout string
	switch args.Layout {
	case "DateTime":
		layout = time.DateTime
	case "DateOnly":
		layout = time.DateOnly
	case "TimeOnly":
		layout = time.TimeOnly
	case "UnixDate":
		layout = time.UnixDate
	case "RFC882":
		layout = time.RFC822
	case "RFC822Z":
		layout = time.RFC822Z
	case "RFC850":
		layout = time.RFC850
	case "RFC1123":
		layout = time.RFC1123
	case "RFC1123Z":
		layout = time.RFC1123Z
	case "RFC3339":
		layout = time.RFC3339
	case "RFC3339Nano":
		layout = time.RFC3339Nano
	default:
		layout = time.RFC3339
	}

	var t time.Time
	if args.Timestamp == 0 {
		t = time.Now().UTC()
	} else {
		t = time.UnixMilli(args.Timestamp).UTC()
	}
	return t.Format(layout)
}
