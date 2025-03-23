package mokapi

import (
	"fmt"
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
	obj.Set("encoding", f.Marshal)
	obj.Set("date", f.Date)
	obj.Set("marshal", f.Marshal)
}

func (m *Module) Sleep(i interface{}) {
	switch t := i.(type) {
	case int64:
		time.Sleep(time.Duration(t) * time.Millisecond)
	case string:
		d, err := time.ParseDuration(t)
		if err != nil {
			panic(m.vm.ToValue(err.Error()))
		}
		time.Sleep(d)
	default:
		panic(m.vm.ToValue(fmt.Errorf("unexpected type for time: %v", i)))
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
