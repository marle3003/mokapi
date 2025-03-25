package schema

import (
	"mokapi/schema/json/schema"
)

type JsonSchemaConverter struct {
	history map[*Schema]*schema.Schema
	useXml  bool
}

func ConvertToJsonSchema(s *Schema) *schema.Schema {
	c := &JsonSchemaConverter{}
	return c.Convert(s)
}

func (c *JsonSchemaConverter) Convert(s *Schema) *schema.Schema {
	if s == nil || s.SubSchema == nil {
		return nil
	}
	if c.history == nil {
		c.history = map[*Schema]*schema.Schema{}
	}
	if r, ok := c.history[s]; ok {
		return r
	}

	js := &schema.Schema{
		Id:                    s.Id,
		Anchor:                s.Anchor,
		Ref:                   s.Ref,
		DynamicRef:            s.DynamicRef,
		Boolean:               s.Boolean,
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
		Items:                 c.Convert(s.Items),
		UnevaluatedItems:      c.Convert(s.UnevaluatedItems),
		Contains:              c.Convert(s.Contains),
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
		UnevaluatedProperties: c.Convert(s.UnevaluatedProperties),
		PropertyNames:         c.Convert(s.PropertyNames),
		Not:                   c.Convert(s.Not),
		If:                    c.Convert(s.If),
		Then:                  c.Convert(s.Then),
		Else:                  c.Convert(s.Else),
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
		js.PrefixItems = append(js.PrefixItems, c.Convert(ref))
	}

	if s.Properties != nil {
		js.Properties = &schema.Schemas{}
		for it := s.Properties.Iter(); it.Next(); {
			propName := it.Key()
			xml := it.Value().Xml
			if c.useXml && xml != nil {
				propName = xml.Name
			}
			js.Properties.Set(propName, c.Convert(it.Value()))
		}
	}

	for k, ref := range s.PatternProperties {
		if js.PatternProperties == nil {
			js.PatternProperties = map[string]*schema.Schema{}
		}
		js.PatternProperties[k] = c.Convert(ref)
	}

	for k, ref := range s.DependentSchemas {
		if js.DependentSchemas == nil {
			js.DependentSchemas = map[string]*schema.Schema{}
		}
		js.DependentSchemas[k] = c.Convert(ref)
	}

	if s.AdditionalProperties != nil {
		js.AdditionalProperties = c.Convert(s.AdditionalProperties)
	}

	for _, anyOf := range s.AnyOf {
		js.AnyOf = append(js.AnyOf, c.Convert(anyOf))
	}

	for _, oneOf := range s.OneOf {
		js.OneOf = append(js.OneOf, c.Convert(oneOf))
	}

	for _, allOf := range s.AllOf {
		js.AllOf = append(js.AllOf, c.Convert(allOf))
	}

	if s.Nullable {
		js.Type = append(js.Type, "null")
	}

	js.Examples = s.Examples
	if s.Example != nil && s.Examples == nil {
		js.Examples = append(js.Examples, *s.Example)
	}

	js.ContentMediaType = s.ContentMediaType
	js.ContentEncoding = s.ContentEncoding

	if s.Definitions != nil {
		js.Definitions = map[string]*schema.Schema{}
		for k, v := range s.Definitions {
			js.Definitions[k] = ConvertToJsonSchema(v)
		}
	}

	if s.Defs != nil {
		js.Defs = map[string]*schema.Schema{}
		for k, v := range s.Defs {
			js.Defs[k] = ConvertToJsonSchema(v)
		}
	}

	return js
}
