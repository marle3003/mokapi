package faker

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/js/eventloop"
	"mokapi/js/util"
	"mokapi/providers/openapi/schema"
	"mokapi/schema/json/generator"
	jsonSchema "mokapi/schema/json/schema"
	"reflect"

	"github.com/dop251/goja"
)

type Module struct {
	vm   *goja.Runtime
	host common.Host
	loop *eventloop.EventLoop
}

func Require(rt *goja.Runtime, module *goja.Object) {
	o := rt.Get("mokapi/internal").(*goja.Object)
	host := o.Get("host").Export().(common.Host)
	loop := o.Get("loop").Export().(*eventloop.EventLoop)
	f := &Module{
		vm:   rt,
		host: host,
		loop: loop,
	}
	obj := module.Get("exports").(*goja.Object)
	_ = obj.Set("fake", f.Fake)
	_ = obj.Set("fakeAsync", f.FakeAsync)
	_ = obj.Set("findByName", f.FindByName)
	_ = obj.Set("ROOT_NAME", generator.RootName)
}

func (m *Module) Fake(v goja.Value) interface{} {
	req, err := m.newRequest(v)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}
	i, err := generate(req)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}
	return i
}

func (m *Module) FakeAsync(v goja.Value) *goja.Promise {
	req, err := m.newRequest(v)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}
	return util.RunAsync(m.vm, m.loop, func() (any, error) {
		return generate(req)
	}, nil)
}

// newRequest converts the JavaScript schema to a plain Go request. It has to run on the event
// loop because it reads from the goja runtime. Without a schema there is nothing to generate
// and the request is nil.
func (m *Module) newRequest(v goja.Value) (*generator.Request, error) {
	if v == nil {
		return nil, nil
	}

	t := v.ExportType()
	if t.Kind() != reflect.Map {
		return nil, fmt.Errorf("expect object parameter but got: %v", util.JsType(v.Export()))
	}

	r := &generator.Request{}
	if isOpenApiSchema(v.ToObject(m.vm)) {
		s, err := toStruct[*schema.Schema](v, m.vm)
		if err != nil {
			return nil, err
		}
		r.Schema = schema.ConvertToJsonSchema(s)
	} else {
		s, err := toStruct[*jsonSchema.Schema](v, m.vm)
		if err != nil {
			return nil, err
		}
		r.Schema = s
	}
	return r, nil
}

// generate must not touch the goja runtime except through the event loop (custom nodes).
func generate(req *generator.Request) (any, error) {
	if req == nil {
		return nil, nil
	}
	return generator.New(req)
}

func toStruct[T any](v goja.Value, vm *goja.Runtime) (T, error) {
	var out T

	// stringify ensures order of fields instead of exporting to a map
	jsonGlobal := vm.GlobalObject().Get("JSON").ToObject(vm)
	stringify, ok := goja.AssertFunction(jsonGlobal.Get("stringify"))
	if !ok {
		return out, fmt.Errorf("JSON.stringify not available")
	}

	res, err := stringify(goja.Undefined(), v)
	if err != nil {
		return out, err
	}

	err = dynamic.UnmarshalJSON([]byte(res.String()), &out)
	if err != nil {
		return out, dynamic.FormatError(nil, err)
	}
	return out, err
}
