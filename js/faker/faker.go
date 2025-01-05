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

type Faker struct {
	rt   *goja.Runtime
	host common.Host
}

func Require(rt *goja.Runtime, module *goja.Object) {
	o := rt.Get("mokapi/internal").(*goja.Object)
	host := o.Get("host").Export().(common.Host)
	f := &Faker{
		rt:   rt,
		host: host,
	}
	obj := module.Get("exports").(*goja.Object)
	obj.Set("fake", f.Fake)
	obj.Set("findByName", f.FindByName)
}

func (m *Faker) Fake(v goja.Value) interface{} {
	if v == nil {
		return nil
	}

	t := v.ExportType()
	if t.Kind() != reflect.Map {
		panic(m.rt.ToValue(fmt.Errorf("expect object parameter but got: %v", util.JsType(v.Export()))))
	}

	r := &generator.Request{}
	if isOpenApiSchema(v.ToObject(m.rt)) {
		s, err := ToOpenAPISchema(v, m.rt)
		if err != nil {
			panic(m.rt.ToValue(err.Error()))
		}
		r.Path = generator.Path{
			&generator.PathElement{
				Schema: schema.ConvertToJsonSchema(s),
			},
		}
	} else {
		s, err := ToJsonSchema(v, m.rt)
		if err != nil {
			panic(m.rt.ToValue(err.Error()))
		}
		r.Path = generator.Path{
			&generator.PathElement{
				Schema: s,
			},
		}
	}

	i, err := generator.New(r)
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}
	return i
}

func (m *Faker) FindByName(name string) goja.Value {
	ft := m.host.FindFakerTree(name)
	return convertToNode(ft, m)
}
