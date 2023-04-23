package js

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
	"mokapi/lib/mustache"
	"reflect"
)

type mustacheModule struct {
	rt *goja.Runtime
}

func newMustache(_ common.Host, rt *goja.Runtime) interface{} {
	return &mustacheModule{rt: rt}
}

func (m *mustacheModule) Render(template string, scopeValue goja.Value) string {
	scope := m.parseScope(scopeValue)

	s, err := mustache.Render(template, scope)
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}
	return s
}

func (m *mustacheModule) parseScope(scopeValue goja.Value) interface{} {
	if scopeValue == nil && goja.IsUndefined(scopeValue) && goja.IsNull(scopeValue) {
		return nil
	}

	if t := scopeValue.ExportType(); t.Kind() != reflect.Map {
		return scopeValue.Export()
	}

	scope := make(map[string]interface{})
	o := scopeValue.ToObject(m.rt)
	for _, k := range o.Keys() {
		v := o.Get(k)
		if c, ok := goja.AssertFunction(v); ok {
			v, err := c(goja.Undefined())
			if err != nil {
				panic(m.rt.ToValue(err.Error()))
			}
			scope[k] = m.parseScope(v)
		} else {
			scope[k] = m.parseScope(v)
		}

	}
	return scope
}
