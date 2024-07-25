package schema

import (
	"mokapi/schema/json/generator"
	jsonRef "mokapi/schema/json/ref"
	"mokapi/schema/json/schema"
)

func CreateValue(ref *Ref) (interface{}, error) {
	if ref == nil {
		return generator.New(&generator.Request{})
	}
	c := &JsonSchemaConverter{}
	r := c.ConvertToJsonRef(ref)
	return generator.New(&generator.Request{Path: generator.Path{&generator.PathElement{Schema: r}}})
}

func ConvertToJsonSchema(ref *Ref) *schema.Ref {
	c := &JsonSchemaConverter{}
	return c.ConvertToJsonRef(ref)
}

type JsonSchemaConverter struct {
	history map[*Schema]*schema.Schema
	useXml  bool
}

func (c *JsonSchemaConverter) Convert(s *Schema) *schema.Schema {
	if s == nil {
		return nil
	}
	if c.history == nil {
		c.history = map[*Schema]*schema.Schema{}
	}
	if r, ok := c.history[s]; ok {
		return r
	}

	js := &schema.Schema{
		Type:                 s.Type,
		Schema:               s.Schema,
		Enum:                 s.Enum,
		Const:                s.Const,
		Default:              s.Default,
		MinLength:            s.MinLength,
		MaxLength:            s.MaxLength,
		Pattern:              s.Pattern,
		Format:               s.Format,
		MultipleOf:           s.MultipleOf,
		Items:                c.ConvertToJsonRef(s.Items),
		MinItems:             s.MinItems,
		MaxItems:             s.MaxItems,
		UniqueItems:          s.UniqueItems,
		ShuffleItems:         s.ShuffleItems,
		MaxProperties:        s.MaxProperties,
		MinProperties:        s.MinProperties,
		Required:             s.Required,
		DependentRequired:    nil,
		AdditionalProperties: schema.AdditionalProperties{},
	}
	c.history[s] = js

	js.Minimum = s.Minimum
	js.ExclusiveMinimum = s.ExclusiveMinimum

	js.Maximum = s.Maximum
	js.ExclusiveMaximum = s.ExclusiveMaximum

	if s.Properties != nil {
		js.Properties = &schema.Schemas{}
		for it := s.Properties.Iter(); it.Next(); {
			propName := it.Key()
			xml := it.Value().getXml()
			if c.useXml && xml != nil {
				propName = xml.Name
			}
			js.Properties.Set(propName, c.ConvertToJsonRef(it.Value()))
		}
	}

	if s.AdditionalProperties != nil {
		js.AdditionalProperties.Forbidden = s.AdditionalProperties.Forbidden
		js.AdditionalProperties.Ref = c.ConvertToJsonRef(s.AdditionalProperties.Ref)
	}

	for _, anyOf := range s.AnyOf {
		js.AnyOf = append(js.AnyOf, c.ConvertToJsonRef(anyOf))
	}

	for _, oneOf := range s.OneOf {
		js.OneOf = append(js.OneOf, c.ConvertToJsonRef(oneOf))
	}

	for _, allOf := range s.AllOf {
		js.AllOf = append(js.AllOf, c.ConvertToJsonRef(allOf))
	}

	if s.Nullable {
		js.Type = append(js.Type, "null")
	}

	js.Examples = s.Examples
	if s.Example != nil && s.Examples == nil {
		js.Examples = append(js.Examples, s.Example)
	}

	js.ContentMediaType = s.ContentMediaType
	js.ContentEncoding = s.ContentEncoding

	return js
}

func (c *JsonSchemaConverter) ConvertToJsonRef(r *Ref) *schema.Ref {
	if r == nil {
		return nil
	}
	js := &schema.Ref{Reference: jsonRef.Reference{Ref: r.Ref}}
	if r.Value == nil {
		return js
	}
	js.Value = c.Convert(r.Value)
	return js
}
