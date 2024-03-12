package js

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
	"mokapi/json/generator"
	"mokapi/providers/openapi/schema"
)

type fakerModule struct {
	rt *goja.Runtime
}

type jsonSchema struct {
	Type                 string                 `json:"type"`
	Format               string                 `json:"format"`
	Pattern              string                 `json:"pattern"`
	Properties           map[string]*jsonSchema `json:"properties"`
	AdditionalProperties *jsonSchema            `json:"additionalProperties,omitempty"`
	Items                *jsonSchema            `json:"items"`
	Required             []string               `json:"required"`
	Nullable             bool                   `json:"nullable"`
	Example              interface{}            `json:"example"`
	Enum                 []interface{}          `json:"enum"`
	Minimum              *float64               `json:"minimum,omitempty"`
	Maximum              *float64               `json:"maximum,omitempty"`
	ExclusiveMinimum     *bool                  `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     *bool                  `json:"exclusiveMaximum,omitempty"`
	AnyOf                []*jsonSchema          `json:"anyOf"`
	AllOf                []*jsonSchema          `json:"allOf"`
	OneOf                []*jsonSchema          `json:"oneOf"`
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

func newFaker(_ common.Host, rt *goja.Runtime) interface{} {
	return &fakerModule{rt: rt}
}

func (m *fakerModule) Fake(v goja.Value) interface{} {
	s := &jsonSchema{}
	err := m.rt.ExportTo(v, &s)
	if err != nil {
		panic(m.rt.ToValue("expected parameter type of OpenAPI schema"))
	}
	i, err := schema.CreateValue(&schema.Ref{Value: toSchema(s)})
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}
	return i
}

func toSchema(js *jsonSchema) *schema.Schema {
	s := &schema.Schema{
		Type:             js.Type,
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

	if len(js.Properties) > 0 {
		s.Properties = &schema.Schemas{}
		for name, prop := range js.Properties {
			s.Properties.Set(name, &schema.Ref{Value: toSchema(prop)})
		}
	}

	if js.AdditionalProperties != nil {
		s.AdditionalProperties = &schema.AdditionalProperties{
			Ref: &schema.Ref{
				Value: toSchema(js.AdditionalProperties),
			},
		}
	}

	if js.Items != nil {
		s.Items = &schema.Ref{
			Value: toSchema(js.Items),
		}
	}

	if len(js.AnyOf) > 0 {
		for _, any := range js.AnyOf {
			s.AnyOf = append(s.AnyOf, &schema.Ref{Value: toSchema(any)})
		}
	}

	if len(js.AllOf) > 0 {
		for _, all := range js.AllOf {
			s.AllOf = append(s.AllOf, &schema.Ref{Value: toSchema(all)})
		}
	}

	if len(js.OneOf) > 0 {
		for _, one := range js.OneOf {
			s.OneOf = append(s.OneOf, &schema.Ref{Value: toSchema(one)})
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
	t  *generator.Tree
	rt *goja.Runtime
}

func (m *fakerModule) FindByName(name string) *node {
	t := generator.FindByName(name)
	return &node{t: t, rt: m.rt}
}

func (n *node) Append(v goja.Value) {
	t := n.createTree(v)
	n.t.Append(t)
}

func (n *node) Insert(index int, v goja.Value) {
	t := n.createTree(v)
	n.t.Insert(index, t)
}

func (n *node) createTree(v goja.Value) *generator.Tree {
	if v != nil && !goja.IsUndefined(v) && !goja.IsNull(v) {
		newNode := &generator.Tree{}
		obj := v.ToObject(n.rt)
		for _, k := range obj.Keys() {
			switch k {
			case "name":
				name := obj.Get(k)
				newNode.Name = name.String()
			case "test":
				test, _ := goja.AssertFunction(obj.Get(k))
				newNode.Test = func(r *generator.Request) bool {
					param := n.rt.ToValue(r)
					v, _ := test(goja.Undefined(), param)
					return v.ToBoolean()
				}
			case "fake":
				fake, _ := goja.AssertFunction(obj.Get(k))
				newNode.Fake = func(r *generator.Request) (interface{}, error) {
					param := n.rt.ToValue(r)
					v, err := fake(goja.Undefined(), param)
					return v.Export(), err
				}
			}
		}
		if newNode.Name == "" {
			panic(n.rt.ToValue("node must have a name"))
		}
		return newNode
	}
	panic(n.rt.ToValue("unexpected function parameter"))
}
