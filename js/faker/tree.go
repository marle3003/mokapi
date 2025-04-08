package faker

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
	"mokapi/schema/json/generator"
)

type node struct {
	t    *common.FakerTree
	f    *Faker
	name string
	test func(r *generator.Request) bool
	fake func(r *generator.Request) (interface{}, error)
}

func (n *node) Name() string {
	if n.t != nil {
		return n.t.Name()
	}
	return n.name
}

func (n *node) Fake(r *generator.Request) (interface{}, error) {
	return n.fake(r)
}

func (n *node) Append(v goja.Value) {
	t := n.createTree(v)
	n.t.Append(t)
}

func (n *node) RemoveAt(index int) {
	if err := n.t.RemoveAt(index); err != nil {
		panic(n.f.rt.ToValue(err))
	}
}

func (n *node) Remove(name string) {
	if err := n.t.Remove(name); err != nil {
		panic(n.f.rt.ToValue(err))
	}
}

func (n *node) Restore() {
	if err := n.t.Restore(); err != nil {
		panic(n.f.rt.ToValue(err))
	}
}

func convertToNode(t *common.FakerTree, m *Faker) goja.Value {
	n := &node{t: t, f: m}
	obj := m.rt.NewObject()
	obj.Set("name", n.Name)
	obj.Set("append", n.Append)
	obj.Set("removeAt", n.RemoveAt)
	obj.Set("remove", n.Remove)
	obj.Set("restore", n.Restore)
	return obj
}

func (n *node) createTree(v goja.Value) *node {
	if v != nil && !goja.IsUndefined(v) && !goja.IsNull(v) {
		newNode := &node{}
		obj := v.ToObject(n.f.rt)
		for _, k := range obj.Keys() {
			switch k {
			case "name":
				name := obj.Get(k)
				newNode.name = name.String()
			case "fake":
				fake, _ := goja.AssertFunction(obj.Get(k))
				newNode.fake = func(r *generator.Request) (interface{}, error) {
					n.f.host.Lock()
					defer n.f.host.Unlock()

					param := n.f.rt.ToValue(r)
					v, err := fake(goja.Undefined(), param)
					return v.Export(), err
				}
			}
		}
		if newNode.name == "" {
			panic(n.f.rt.ToValue("node must have a name"))
		}
		return newNode
	}
	panic(n.f.rt.ToValue("unexpected function parameter"))
}
