package faker

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
	"mokapi/providers/openapi/schema"
	"mokapi/schema/json/generator"
	jsonSchema "mokapi/schema/json/schema"
)

type Faker struct {
	rt   *goja.Runtime
	host common.Host
}

type jsonXml struct {
	Wrapped   bool   `json:"wrapped"`
	Name      string `json:"name"`
	Attribute bool   `json:"attribute"`
	Prefix    string `json:"prefix"`
	Namespace string `json:"namespace"`
}

type requestExample struct {
	Name   string             `json:"name"`
	Schema *jsonSchema.Schema `json:"schema"`
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
	r := &generator.Request{}
	err := m.rt.ExportTo(v, &r)
	if err == nil && r.Path != nil {
		v, err := generator.New(r)
		if err != nil {
			panic(m.rt.ToValue(err.Error()))
		}
		return v
	}

	s, err := ToOpenAPISchema(v, m.rt)
	//s, err := ToJsonSchema(v, m.rt)
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}

	r = &generator.Request{
		Path: generator.Path{
			&generator.PathElement{
				Schema: schema.ConvertToJsonSchema(s),
			},
		},
	}

	i, err := generator.New(r)
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}
	return i
}

type node struct {
	t    common.FakerTree
	f    *Faker
	name string
	test func(r *generator.Request) bool
	fake func(r *generator.Request) (interface{}, error)
}

func (m *Faker) FindByName(name string) *node {
	ft := m.host.FindFakerTree(name)
	return &node{t: ft, f: m}
}

func (n *node) Name() string {
	return n.name
}

func (n *node) Test(r *generator.Request) bool {
	return n.test(r)
}

func (n *node) Fake(r *generator.Request) (interface{}, error) {
	return n.fake(r)
}

func (n *node) Append(v goja.Value) {
	t := n.createTree(v)
	n.t.Append(t)
}

func (n *node) Insert(index int, v goja.Value) {
	t := n.createTree(v)
	err := n.t.Insert(index, t)
	if err != nil {
		panic(n.f.rt.ToValue(err))
	}
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

func (n *node) createTree(v goja.Value) *node {
	if v != nil && !goja.IsUndefined(v) && !goja.IsNull(v) {
		newNode := &node{}
		obj := v.ToObject(n.f.rt)
		for _, k := range obj.Keys() {
			switch k {
			case "name":
				name := obj.Get(k)
				newNode.name = name.String()
			case "test":
				test, _ := goja.AssertFunction(obj.Get(k))
				newNode.test = func(r *generator.Request) bool {
					n.f.host.Lock()
					defer n.f.host.Unlock()

					param := n.f.rt.ToValue(r)
					v, _ := test(goja.Undefined(), param)
					return v.ToBoolean()
				}
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

func toRequest(v goja.Value, rt *goja.Runtime) *generator.Request {
	r := &generator.Request{
		Path: generator.Path{
			&generator.PathElement{},
		},
	}
	obj := v.ToObject(rt)
	var err error
	for _, k := range obj.Keys() {
		switch k {
		case "name":
			name := obj.Get(k)
			r.Path[0].Name = name.String()
		case "schema":
			s := obj.Get(k)
			r.Path[0].Schema, err = ToJsonSchema(s, rt)
			if err != nil {
				return nil
			}
		default:
			return nil
		}
	}
	return r
}
