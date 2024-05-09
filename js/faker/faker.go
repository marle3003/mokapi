package faker

import (
	"fmt"
	"github.com/dop251/goja"
	"mokapi/engine/common"
	"mokapi/json/generator"
	jsonSchema "mokapi/json/schema"
	"mokapi/providers/openapi/schema"
)

type Faker struct {
	rt   *goja.Runtime
	host common.Host
}

type JsonSchema struct {
	Type                 interface{}            `json:"type"`
	Format               string                 `json:"format"`
	Pattern              string                 `json:"pattern"`
	Properties           map[string]*JsonSchema `json:"properties"`
	AdditionalProperties *JsonSchema            `json:"additionalProperties,omitempty"`
	Items                *JsonSchema            `json:"items"`
	Required             []string               `json:"required"`
	Nullable             bool                   `json:"nullable"`
	Example              interface{}            `json:"example"`
	Enum                 []interface{}          `json:"enum"`
	Minimum              *float64               `json:"minimum,omitempty"`
	Maximum              *float64               `json:"maximum,omitempty"`
	ExclusiveMinimum     *bool                  `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     *bool                  `json:"exclusiveMaximum,omitempty"`
	AnyOf                []*JsonSchema          `json:"anyOf"`
	AllOf                []*JsonSchema          `json:"allOf"`
	OneOf                []*JsonSchema          `json:"oneOf"`
	UniqueItems          bool                   `json:"uniqueItems"`
	MinItems             *int                   `json:"minItems"`
	MaxItems             *int                   `json:"maxItems"`
	ShuffleItems         bool                   `json:"x-shuffleItems"`
	MinProperties        *int                   `json:"minProperties"`
	MaxProperties        *int                   `json:"maxProperties"`
	Xml                  *jsonXml               `json:"xml"`
}

type jsonXml struct {
	Wrapped   bool   `json:"wrapped"`
	Name      string `json:"name"`
	Attribute bool   `json:"attribute"`
	Prefix    string `json:"prefix"`
	Namespace string `json:"namespace"`
}

type requestExample struct {
	Name   string      `json:"name"`
	Schema *JsonSchema `json:"schema"`
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
	r := toRequest(v, m.rt)
	err := m.rt.ExportTo(v, &r)
	if err == nil && r.Path != nil {
		v, err := generator.New(r)
		if err != nil {
			panic(m.rt.ToValue(err.Error()))
		}
		return v
	}

	s := &JsonSchema{}
	err = m.rt.ExportTo(v, &s)
	if err != nil {
		panic(m.rt.ToValue("expected parameter type of OpenAPI schema"))
	}
	i, err := schema.CreateValue(&schema.Ref{Value: ConvertToSchema(s)})
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}
	return i
}

func ConvertToSchema(js *JsonSchema) *schema.Schema {
	s := &schema.Schema{
		Format:           js.Format,
		Pattern:          js.Pattern,
		Required:         js.Required,
		Nullable:         js.Nullable,
		Example:          js.Example,
		Enum:             js.Enum,
		Minimum:          js.Minimum,
		Maximum:          js.Maximum,
		ExclusiveMinimum: js.ExclusiveMinimum,
		ExclusiveMaximum: js.ExclusiveMaximum,
		UniqueItems:      js.UniqueItems,
		MinItems:         js.MinItems,
		MaxItems:         js.MaxItems,
		ShuffleItems:     js.ShuffleItems,
		MinProperties:    js.MinProperties,
		MaxProperties:    js.MaxProperties,
	}

	if js.Type != nil {
		switch v := js.Type.(type) {
		case string:
			s.Type = append(s.Type, v)
		case []interface{}:
			for _, typeName := range v {
				s.Type = append(s.Type, fmt.Sprintf("%v", typeName))
			}
		}
	}

	if len(js.Properties) > 0 {
		s.Properties = &schema.Schemas{}
		for name, prop := range js.Properties {
			s.Properties.Set(name, &schema.Ref{Value: ConvertToSchema(prop)})
		}
	}

	if js.AdditionalProperties != nil {
		s.AdditionalProperties = &schema.AdditionalProperties{
			Ref: &schema.Ref{
				Value: ConvertToSchema(js.AdditionalProperties),
			},
		}
	}

	if js.Items != nil {
		s.Items = &schema.Ref{
			Value: ConvertToSchema(js.Items),
		}
	}

	if len(js.AnyOf) > 0 {
		for _, any := range js.AnyOf {
			s.AnyOf = append(s.AnyOf, &schema.Ref{Value: ConvertToSchema(any)})
		}
	}

	if len(js.AllOf) > 0 {
		for _, all := range js.AllOf {
			s.AllOf = append(s.AllOf, &schema.Ref{Value: ConvertToSchema(all)})
		}
	}

	if len(js.OneOf) > 0 {
		for _, one := range js.OneOf {
			s.OneOf = append(s.OneOf, &schema.Ref{Value: ConvertToSchema(one)})
		}
	}

	if js.Xml != nil {
		s.Xml = &schema.Xml{
			Wrapped:   js.Xml.Wrapped,
			Name:      js.Xml.Name,
			Attribute: js.Xml.Attribute,
			Prefix:    js.Xml.Prefix,
			Namespace: js.Xml.Namespace,
		}
	}

	return s
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
			r.Path[0].Schema, err = toJsonSchema(s, rt)
			if err != nil {
				return nil
			}
		default:
			return nil
		}
	}
	return r
}

func toJsonSchema(v goja.Value, rt *goja.Runtime) (*jsonSchema.Ref, error) {
	s := &jsonSchema.Schema{}
	obj := v.ToObject(rt)
	for _, k := range obj.Keys() {
		switch k {
		case "type":
			i := obj.Get(k).Export()
			if arr, ok := i.([]interface{}); ok {
				for _, t := range arr {
					tn, ok := t.(string)
					if !ok {
						return nil, fmt.Errorf("unexpected type: %v", t)
					}
					s.Type = append(s.Type, tn)
				}
			} else if t, ok := i.(string); ok {
				s.Type = []string{t}
			} else {
				return nil, fmt.Errorf("unexpected type for attribute 'type'")
			}
		case "enum":
			i := obj.Get(k).Export()
			if enums, ok := i.([]interface{}); ok {
				s.Enum = enums
			} else {
				return nil, fmt.Errorf("unexpected type for attribute 'enum'")
			}
		case "const":
			s.Const = obj.Get(k).Export()
		case "examples":
			i := obj.Get(k).Export()
			if examples, ok := i.([]interface{}); ok {
				s.Examples = examples
			} else {
				return nil, fmt.Errorf("unexpected type for attribute 'examples'")
			}
		case "multipleOf":
			f := obj.Get(k).ToFloat()
			s.MultipleOf = &f
		case "maximum":
			f := obj.Get(k).ToFloat()
			s.MultipleOf = &f
		case "exclusiveMaximum":
			f := obj.Get(k).ToFloat()
			s.ExclusiveMaximum = &f
		case "minimum":
			f := obj.Get(k).ToFloat()
			s.Minimum = &f
		case "exclusiveMinimum":
			f := obj.Get(k).ToFloat()
			s.ExclusiveMinimum = &f
		case "maxLength":
			i := int(obj.Get(k).ToInteger())
			s.MaxLength = &i
		case "minLength":
			i := int(obj.Get(k).ToInteger())
			s.MinLength = &i
		case "pattern":
			s.Pattern = obj.Get(k).String()
		case "format":
			s.Format = obj.Get(k).String()
		case "items":
			items, err := toJsonSchema(obj.Get(k), rt)
			if err != nil {
				return nil, err
			}
			s.Items = items
		case "maxItems":
			i := int(obj.Get(k).ToInteger())
			s.MaxItems = &i
		case "minItems":
			i := int(obj.Get(k).ToInteger())
			s.MinItems = &i
		case "uniqueItems":
			s.UniqueItems = obj.Get(k).ToBoolean()
		case "maxContains":
			i := int(obj.Get(k).ToInteger())
			s.MaxContains = &i
		case "minContains":
			i := int(obj.Get(k).ToInteger())
			s.MinContains = &i
		case "x-shuffleItems":
			s.ShuffleItems = obj.Get(k).ToBoolean()
		case "properties":
			s.Properties = &jsonSchema.Schemas{}
			propsObj := obj.Get(k).ToObject(rt)
			for _, name := range propsObj.Keys() {
				prop, err := toJsonSchema(propsObj.Get(name), rt)
				if err != nil {
					return nil, err
				}
				s.Properties.Set(name, prop)
			}
		case "maxProperties":
			i := int(obj.Get(k).ToInteger())
			s.MaxProperties = &i
		case "minProperties":
			i := int(obj.Get(k).ToInteger())
			s.MinProperties = &i
		case "required":
			i := obj.Get(k).Export()
			if arr, ok := i.([]interface{}); ok {
				for _, t := range arr {
					req, ok := t.(string)
					if !ok {
						return nil, fmt.Errorf("unexpected type: %v", t)
					}
					s.Required = append(s.Type, req)
				}
			}
		}
	}
	return &jsonSchema.Ref{Value: s}, nil
}
