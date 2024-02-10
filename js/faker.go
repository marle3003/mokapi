package js

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
	"mokapi/providers/openapi/schema"
)

type fakerModule struct {
	rt *goja.Runtime
}

type jsSchema struct {
	Type                 string               `json:"type"`
	Format               string               `json:"format"`
	Pattern              string               `json:"pattern"`
	Properties           map[string]*jsSchema `json:"properties"`
	AdditionalProperties *jsSchema            `json:"additionalProperties,omitempty"`
	Items                *jsSchema            `json:"items"`
	Required             []string             `json:"required"`
	Nullable             bool                 `json:"nullable"`
	Example              interface{}          `json:"example"`
	Enum                 []interface{}        `json:"enum"`
	Minimum              *float64             `json:"minimum,omitempty"`
	Maximum              *float64             `json:"maximum,omitempty"`
	ExclusiveMinimum     *bool                `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     *bool                `json:"exclusiveMaximum,omitempty"`
	AnyOf                []*jsSchema          `json:"anyOf"`
	AllOf                []*jsSchema          `json:"allOf"`
	OneOf                []*jsSchema          `json:"oneOf"`
	UniqueItems          bool                 `json:"uniqueItems"`
	MinItems             *int                 `json:"minItems"`
	MaxItems             *int                 `json:"maxItems"`
	ShuffleItems         bool                 `json:"x-shuffleItems"`
	MinProperties        *int                 `json:"minProperties"`
	MaxProperties        *int                 `json:"maxProperties"`
}

func newFaker(_ common.Host, rt *goja.Runtime) interface{} {
	return &fakerModule{rt: rt}
}

func (m *fakerModule) Fake(v goja.Value) interface{} {
	s := &jsSchema{}
	err := m.rt.ExportTo(v, &s)
	if err != nil {
		panic(m.rt.ToValue("expected parameter type of OpenAPI schema"))
	}
	i, err := schema.CreateValue(&schema.Ref{Value: m.toSchema(s)})
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}
	return i
}

func (m *fakerModule) toSchema(js *jsSchema) *schema.Schema {
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
			s.Properties.Set(name, &schema.Ref{Value: m.toSchema(prop)})
		}
	}

	if js.AdditionalProperties != nil {
		s.AdditionalProperties = &schema.AdditionalProperties{
			Ref: &schema.Ref{
				Value: m.toSchema(js.AdditionalProperties),
			},
		}
	}

	if js.Items != nil {
		s.Items = &schema.Ref{
			Value: m.toSchema(js.Items),
		}
	}

	if len(js.AnyOf) > 0 {
		for _, any := range js.AnyOf {
			s.AnyOf = append(s.AnyOf, &schema.Ref{Value: m.toSchema(any)})
		}
	}

	if len(js.AllOf) > 0 {
		for _, all := range js.AllOf {
			s.AllOf = append(s.AllOf, &schema.Ref{Value: m.toSchema(all)})
		}
	}

	if len(js.OneOf) > 0 {
		for _, one := range js.OneOf {
			s.OneOf = append(s.OneOf, &schema.Ref{Value: m.toSchema(one)})
		}
	}

	return s
}
