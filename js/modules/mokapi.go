package modules

import (
	"github.com/dop251/goja"
	"mokapi/js/common"
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

func (m *Mokapi) Every(every string, do func(), times int) (int, error) {
	return m.host.Every(every, do, times)
}

func (m *Mokapi) Cron(expr string, do func(), times int) (int, error) {
	return m.host.Cron(expr, do, times)
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
