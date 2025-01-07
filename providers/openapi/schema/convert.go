package schema

import (
	jsonRef "mokapi/schema/json/ref"
	"mokapi/schema/json/schema"
)

type JsonSchemaConverter struct {
	history map[*Schema]*schema.Schema
	useXml  bool
}

func ConvertToJsonSchema(ref *Ref) *schema.Ref {
	c := &JsonSchemaConverter{}
	return c.ConvertToJsonRef(ref)
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
		Type:                  s.Type,
		Schema:                s.Schema,
		Enum:                  s.Enum,
		Const:                 s.Const,
		Default:               s.Default,
		MinLength:             s.MinLength,
		MaxLength:             s.MaxLength,
		Pattern:               s.Pattern,
		Format:                s.Format,
		MultipleOf:            s.MultipleOf,
		Items:                 c.ConvertToJsonRef(s.Items),
		UnevaluatedItems:      c.ConvertToJsonRef(s.UnevaluatedItems),
		Contains:              c.ConvertToJsonRef(s.Contains),
		MaxContains:           s.MaxContains,
		MinContains:           s.MinContains,
		MinItems:              s.MinItems,
		MaxItems:              s.MaxItems,
		UniqueItems:           s.UniqueItems,
		ShuffleItems:          s.ShuffleItems,
		MaxProperties:         s.MaxProperties,
		MinProperties:         s.MinProperties,
		Required:              s.Required,
		DependentRequired:     s.DependentRequired,
		UnevaluatedProperties: c.ConvertToJsonRef(s.UnevaluatedProperties),
		PropertyNames:         c.ConvertToJsonRef(s.PropertyNames),
		Not:                   c.ConvertToJsonRef(s.Not),
		If:                    c.ConvertToJsonRef(s.If),
		Then:                  c.ConvertToJsonRef(s.Then),
		Else:                  c.ConvertToJsonRef(s.Else),
		Title:                 s.Title,
		Description:           s.Description,
		Deprecated:            s.Deprecated,
	}
	c.history[s] = js

	js.Minimum = s.Minimum
	js.ExclusiveMinimum = s.ExclusiveMinimum

	js.Maximum = s.Maximum
	js.ExclusiveMaximum = s.ExclusiveMaximum

	for _, ref := range s.PrefixItems {
		js.PrefixItems = append(js.PrefixItems, c.ConvertToJsonRef(ref))
	}

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

	for k, ref := range s.PatternProperties {
		if js.PatternProperties == nil {
			js.PatternProperties = map[string]*schema.Ref{}
		}
		js.PatternProperties[k] = c.ConvertToJsonRef(ref)
	}

	for k, ref := range s.DependentSchemas {
		if js.DependentSchemas == nil {
			js.DependentSchemas = map[string]*schema.Ref{}
		}
		js.DependentSchemas[k] = c.ConvertToJsonRef(ref)
	}

	if s.AdditionalProperties != nil {
		js.AdditionalProperties = c.ConvertToJsonRef(s.AdditionalProperties)
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

	if s.Definitions != nil {
		js.Definitions = map[string]*schema.Ref{}
		for k, v := range s.Definitions {
			js.Definitions[k] = ConvertToJsonSchema(v)
		}
	}

	if s.Defs != nil {
		js.Defs = map[string]*schema.Ref{}
		for k, v := range s.Defs {
			js.Defs[k] = ConvertToJsonSchema(v)
		}
	}

	return js
}

func (c *JsonSchemaConverter) ConvertToJsonRef(r *Ref) *schema.Ref {
	if r == nil {
		return nil
	}
	js := &schema.Ref{Reference: jsonRef.Reference{Ref: r.Ref}, Boolean: r.Boolean}
	if r.Value == nil {
		return js
	}
	js.Value = c.Convert(r.Value)
	return js
}
