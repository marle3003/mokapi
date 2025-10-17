package faker

import (
	"fmt"
	"mokapi/js/util"
	"mokapi/schema/json/generator"
	"reflect"
	"strconv"

	"github.com/dop251/goja"
)

type converter[T any] func(v goja.Value) T

func (m *Module) FindByName(name string) goja.Value {
	n := m.host.FindFakerNode(name)
	if n == nil {
		return nil
	}
	return NewNode(m, n)
}

func convertToNode(v goja.Value, m *Module) *generator.Node {
	if v != nil && !goja.IsUndefined(v) && !goja.IsNull(v) {
		n := &generator.Node{Custom: true}
		obj := v.ToObject(m.vm)
		for _, k := range obj.Keys() {
			switch k {
			case "name":
				name := obj.Get(k)
				n.Name = name.String()
			case "fake":
				fake, _ := goja.AssertFunction(obj.Get(k))
				n.Fake = func(r *generator.Request) (interface{}, error) {
					m.host.Lock()
					defer m.host.Unlock()

					param := m.vm.ToValue(r)
					v, err := fake(goja.Undefined(), param)
					if err != nil {
						return nil, err
					}
					return v.Export(), err
				}
			case "attributes":
				i := obj.Get(k).Export()
				n.Attributes = toStringArray(i)
			case "dependsOn":
				i := obj.Get(k).Export()
				n.Attributes = toStringArray(i)
			case "children":
				val := obj.Get(k)
				if val.ExportType().Kind() != reflect.Slice {
					s := fmt.Sprintf("unexpected type for 'children': got %s, expected Array", util.JsType(val))
					panic(m.vm.ToValue(s))
				}
				arr := val.ToObject(m.vm)
				length := int(arr.Get("length").ToInteger())
				for i := 0; i < length; i++ {
					item := arr.Get(strconv.Itoa(i))
					n.Children = append(n.Children, convertToNode(item, m))
				}
			}
		}
		if n.Name == "" {
			panic(m.vm.ToValue("node must have a name"))
		}
		return n
	}
	panic(m.vm.ToValue("unexpected function parameter"))
}

func toStringArray(i interface{}) []string {
	values, ok := i.([]any)
	var result []string
	if ok {
		for _, val := range values {
			s, ok := val.(string)
			if !ok {
				err := fmt.Errorf("unexpected type: %v", util.JsType(val))
				panic(err)
			}
			result = append(result, s)
		}
	}
	return result
}
