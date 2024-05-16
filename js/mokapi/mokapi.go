package mokapi

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
	"mokapi/js/eventloop"
	"mokapi/js/faker"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	"os"
	"time"
)

type Module struct {
	host common.Host
	vm   *goja.Runtime
	loop *eventloop.EventLoop
}

func Require(vm *goja.Runtime, module *goja.Object) {
	o := vm.Get("mokapi/internal").(*goja.Object)
	host := o.Get("host").Export().(common.Host)
	loop := o.Get("loop").Export().(*eventloop.EventLoop)
	f := &Module{
		vm:   vm,
		host: host,
		loop: loop,
	}
	obj := module.Get("exports").(*goja.Object)
	obj.Set("sleep", f.Sleep)
	obj.Set("every", f.Every)
	obj.Set("cron", f.Cron)
	obj.Set("on", f.On)
	obj.Set("env", f.Env)
	obj.Set("marshal", f.Marshal)
	obj.Set("date", f.Date)
}

func (m *Module) Sleep(i interface{}) error {
	switch t := i.(type) {
	case int64:
		time.Sleep(time.Duration(t) * time.Millisecond)
	case string:
		d, err := time.ParseDuration(t)
		if err != nil {
			panic(m.vm.ToValue(err.Error()))
		}
		time.Sleep(d)
	}
	return nil
}

func (m *Module) Every(every string, do func(), args goja.Value) (int, error) {
	options := common.NewJobOptions()

	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(m.vm)
		for _, k := range params.Keys() {
			switch k {
			case "tags":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				tags := tagsV.ToObject(m.vm)
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

func (m *Module) Cron(expr string, do func(), args goja.Value) (int, error) {
	options := common.NewJobOptions()

	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(m.vm)
		for _, k := range params.Keys() {
			switch k {
			case "tags":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				tags := tagsV.ToObject(m.vm)
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

func (m *Module) On(event string, do goja.Value, args goja.Value) {
	tags := make(map[string]string)

	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(m.vm)
		for _, k := range params.Keys() {
			switch k {
			case "tags":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				tagsO := tagsV.ToObject(m.vm)
				for _, key := range tagsO.Keys() {
					tags[key] = tagsO.Get(key).String()
				}
			}
		}
	}

	f := func(args ...interface{}) (bool, error) {
		m.host.Lock()
		defer m.host.Unlock()

		r, err := m.loop.RunAsync(func(vm *goja.Runtime) (goja.Value, error) {
			call, _ := goja.AssertFunction(do)
			var params []goja.Value
			for _, v := range args {
				params = append(params, vm.ToValue(v))
			}
			v, err := call(goja.Undefined(), params...)
			if err != nil {
				return nil, err
			}
			return v, nil
		})

		if err != nil {
			return false, err
		}

		return r.ToBoolean(), nil
	}

	m.host.On(event, f, tags)
}

func (m *Module) Env(name string) string {
	return os.Getenv(name)
}

func (m *Module) Open(file string) (string, error) {
	f, err := m.host.OpenFile(file, "")
	if err != nil {
		return "", err
	}
	return string(f.Raw), nil
}

type MarshalArg struct {
	Schema      *faker.JsonSchema `json:"schema"`
	ContentType string            `json:"contentType"`
}

func (m *Module) Marshal(i interface{}, encoding *MarshalArg) string {
	ct := media.ContentType{}
	r := &schema.Ref{}
	if encoding != nil {
		ct = media.ParseContentType(encoding.ContentType)
		v, err := faker.ConvertToSchema(encoding.Schema)
		if err != nil {
			panic(m.vm.ToValue(err.Error()))
		}
		r.Value = v
	}
	if ct.IsEmpty() {
		ct = media.ParseContentType("application/json")
	}

	b, err := r.Marshal(i, ct)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}
	return string(b)
}

type DateArg struct {
	Layout    string `json:"layout"`
	Timestamp int64  `json:"timestamp"`
}

func (m *Module) Date(args DateArg) string {
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
