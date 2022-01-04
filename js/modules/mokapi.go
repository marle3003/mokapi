package modules

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
	"time"
)

type Mokapi struct {
	host common.Host
	rt   *goja.Runtime
}

func NewMokapi(host common.Host, rt *goja.Runtime) interface{} {
	return &Mokapi{host: host, rt: rt}
}

func (*Mokapi) Sleep(milliseconds float64) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}

func (m *Mokapi) Every(every string, do func(), args goja.Value) (int, error) {
	times := -1
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
			case "times":
				tagsV := params.Get(k)
				times = int(tagsV.ToInteger())
			}
		}
	}
	return m.host.Every(every, do, times, tags)
}

func (m *Mokapi) Cron(expr string, do func(), args goja.Value) (int, error) {
	times := -1
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
			case "times":
				tagsV := params.Get(k)
				times = int(tagsV.ToInteger())
			}
		}
	}

	return m.host.Cron(expr, do, times, tags)
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
