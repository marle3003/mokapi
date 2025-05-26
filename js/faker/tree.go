package faker

import (
	"github.com/dop251/goja"
	"mokapi/schema/json/generator"
)

type converter[T any] func(v goja.Value) T

func (m *Module) FindByName(name string) goja.Value {
	n := m.host.FindFakerNode(name)
	if n == nil {
		return nil
	}
	return NewNode(m, n)
}

func (n *Node) Fake(r *generator.Request) (interface{}, error) {
	return n.fake(r)
}

func convertToNode(v goja.Value, m *Module) *generator.Node {
	if v != nil && !goja.IsUndefined(v) && !goja.IsNull(v) {
		n := &generator.Node{}
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
					return v.Export(), err
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
