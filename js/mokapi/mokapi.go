package mokapi

import (
	"fmt"
	"mokapi/engine/common"
	"mokapi/js/eventloop"
	"mokapi/js/faker"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	"os"
	"time"

	"github.com/dop251/goja"
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
	m := &Module{
		vm:   vm,
		host: host,
		loop: loop,
	}
	obj := module.Get("exports").(*goja.Object)
	_ = obj.Set("sleep", m.Sleep)
	_ = obj.Set("delay", m.Delay)
	_ = obj.Set("every", m.Every)
	_ = obj.Set("cron", m.Cron)
	_ = obj.Set("on", m.On)
	_ = obj.Set("env", m.Env)
	_ = obj.Set("encoding", m.Marshal)
	_ = obj.Set("date", m.Date)
	_ = obj.Set("marshal", m.Marshal)
	_ = obj.Set("patch", patch)
	_ = obj.Set("Delete", Delete)
	_ = obj.Set("shared", NewSharedMemory(host.Store(), vm))
	_ = obj.Set("webhook", m.Webhook)
	_ = obj.Set("app", &App{m: m})
}

func (m *Module) Sleep(i interface{}) {
	time.Sleep(m.timeOrThrow(i))
}

// Delay returns a Promise that resolves after the given time. Unlike Sleep it does not block
// the event loop, so other handlers of the script keep running while the Promise is pending.
func (m *Module) Delay(i interface{}) *goja.Promise {
	// like setTimeout, a float is truncated to whole milliseconds
	if f, ok := i.(float64); ok {
		i = m.vm.ToValue(f).ToInteger()
	}
	d := m.timeOrThrow(i)

	p, resolve, _ := m.vm.NewPromise()
	time.AfterFunc(d, func() {
		m.loop.Run(func(*goja.Runtime) { _ = resolve(goja.Undefined()) })
	})
	return p
}

// timeOrThrow throws a JavaScript error for an invalid time argument.
func (m *Module) timeOrThrow(i interface{}) time.Duration {
	d, err := parseTime(i)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}
	return d
}

// parseTime reads a time argument of sleep and delay: milliseconds or a Go duration string.
func parseTime(i interface{}) (time.Duration, error) {
	switch t := i.(type) {
	case int64:
		return time.Duration(t) * time.Millisecond, nil
	case string:
		return time.ParseDuration(t)
	default:
		return 0, fmt.Errorf("unexpected type for time: %v", i)
	}
}

func (m *Module) Env(name string) string {
	return os.Getenv(name)
}

type MarshalArg struct {
	Schema      goja.Value `json:"schema"`
	ContentType string     `json:"contentType"`
}

func (m *Module) Marshal(i interface{}, args *MarshalArg) string {
	ct := media.ContentType{}
	var s *schema.Schema
	if args != nil {
		ct = media.ParseContentType(args.ContentType)
		var err error
		s, err = faker.ToOpenAPISchema(args.Schema, m.vm)
		if err != nil {
			panic(m.vm.ToValue(err.Error()))
		}
	}
	if ct.IsEmpty() {
		ct = media.ParseContentType("application/json")
	}

	b, err := s.Marshal(i, ct)
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
