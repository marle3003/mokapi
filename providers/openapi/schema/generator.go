package schema

import (
	"mokapi/json/generator"
	"mokapi/json/schema"
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
		Enum:                 s.Enum,
		MinLength:            s.MinLength,
		MaxLength:            s.MaxLength,
		Pattern:              s.Pattern,
		Format:               s.Format,
		Items:                c.ConvertToJsonRef(s.Items),
		MinItems:             s.MinItems,
		MaxItems:             s.MaxItems,
		UniqueItems:          s.UniqueItems,
		ShuffleItems:         s.ShuffleItems,
		MaxProperties:        s.MaxProperties,
		MinProperties:        s.MinProperties,
		Required:             nil,
		DependentRequired:    nil,
		AdditionalProperties: schema.AdditionalProperties{},
	}
	c.history[s] = js

	if s.Minimum != nil {
		if s.ExclusiveMinimum != nil && *s.ExclusiveMinimum {
			if s.Type == "integer" {
				min := *s.Minimum + 1
				js.ExclusiveMinimum = &min
			} else {
				js.ExclusiveMinimum = s.Minimum
			}
		} else {
			js.Minimum = s.Minimum
		}
	}
	if s.Maximum != nil {
		if s.ExclusiveMaximum != nil && *s.ExclusiveMaximum {
			if s.Type == "integer" {
				max := *s.Maximum - 1
				js.ExclusiveMaximum = &max
			} else {
				js.ExclusiveMaximum = s.Maximum
			}
		} else {
			js.Maximum = s.Maximum
		}
	}

	if len(s.Type) > 0 {
		js.Type = append(js.Type, s.Type)
	}

	if s.Properties != nil {
		js.Properties = &schema.Schemas{}
		for it := s.Properties.Iter(); it.Next(); {
			js.Properties.Set(it.Key(), c.ConvertToJsonRef(it.Value()))
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

	if s.Example != nil {
		js.Examples = append(js.Examples, s.Example)
	}

	return js
}

func (c *JsonSchemaConverter) ConvertToJsonRef(r *Ref) *schema.Ref {
	if r == nil {
		return nil
	}
	js := &schema.Ref{Reference: r.Reference}
	if r.Value == nil {
		return js
	}
	js.Value = c.Convert(r.Value)
	return js
}
