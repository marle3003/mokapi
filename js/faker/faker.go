package faker

import (
	"fmt"
	"github.com/dop251/goja"
	"mokapi/engine/common"
	"mokapi/js/util"
	"mokapi/providers/openapi/schema"
	"mokapi/schema/json/generator"
	"reflect"
)

type Module struct {
	vm   *goja.Runtime
	host common.Host
}

func Require(rt *goja.Runtime, module *goja.Object) {
	o := rt.Get("mokapi/internal").(*goja.Object)
	host := o.Get("host").Export().(common.Host)
	f := &Module{
		vm:   rt,
		host: host,
	}
	obj := module.Get("exports").(*goja.Object)
	obj.Set("fake", f.Fake)
	obj.Set("findByName", f.FindByName)
}

func (m *Module) Fake(v goja.Value) interface{} {
	if v == nil {
		return nil
	}

	t := v.ExportType()
	if t.Kind() != reflect.Map {
		panic(m.vm.ToValue(fmt.Errorf("expect object parameter but got: %v", util.JsType(v.Export()))))
	}

	r := &generator.Request{}
	if isOpenApiSchema(v.ToObject(m.vm)) {
		s, err := ToOpenAPISchema(v, m.vm)
		if err != nil {
			panic(m.vm.ToValue(err.Error()))
		}
		r.Schema = schema.ConvertToJsonSchema(s)
	} else {
		s, err := ToJsonSchema(v, m.vm)
		if err != nil {
			panic(m.vm.ToValue(err.Error()))
		}
		r.Schema = s
	}

	i, err := generator.New(r)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}
	return i
}
