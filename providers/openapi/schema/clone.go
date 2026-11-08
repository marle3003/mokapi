package schema

import (
	"maps"
	"slices"
)

// Clone returns a deep copy of s. Resolved references make the schema graph
// cyclic (a property that refers to its parent), so every schema is copied once
// and the copy has the same cycles as the original.
func (s *Schema) Clone() *Schema {
	return s.clone(map[*Schema]*Schema{})
}

// Clone returns a deep copy of s, see Schema.Clone.
func (s *Schemas) Clone() *Schemas {
	if s == nil {
		return nil
	}
	return s.clone(map[*Schema]*Schema{})
}

type schemaCopies map[*Schema]*Schema

func (s *Schemas) clone(copies schemaCopies) *Schemas {
	if s == nil {
		return nil
	}
	return &Schemas{LinkedHashMap: s.CloneWith(func(v *Schema) *Schema { return v.clone(copies) })}
}

func (s *Schema) clone(copies schemaCopies) *Schema {
	if s == nil {
		return nil
	}
	if c, ok := copies[s]; ok {
		return c
	}
	c := &Schema{}
	copies[s] = c
	*c = *s

	c.Sub = s.Sub.clone(copies)
	c.m = maps.Clone(s.m)

	c.Boolean = clonePointer(s.Boolean)
	c.Type = slices.Clone(s.Type)
	// values of type any are taken over, they are never patched
	c.Enum = slices.Clone(s.Enum)
	c.Const = clonePointer(s.Const)

	c.MultipleOf = clonePointer(s.MultipleOf)
	c.Minimum = clonePointer(s.Minimum)
	c.Maximum = clonePointer(s.Maximum)
	c.ExclusiveMinimum = clonePointer(s.ExclusiveMinimum)
	c.ExclusiveMaximum = clonePointer(s.ExclusiveMaximum)

	c.MinLength = clonePointer(s.MinLength)
	c.MaxLength = clonePointer(s.MaxLength)

	c.Items = s.Items.clone(copies)
	c.PrefixItems = cloneSchemas(s.PrefixItems, copies)
	c.UnevaluatedItems = s.UnevaluatedItems.clone(copies)
	c.Contains = s.Contains.clone(copies)
	c.MaxContains = clonePointer(s.MaxContains)
	c.MinContains = clonePointer(s.MinContains)
	c.MinItems = clonePointer(s.MinItems)
	c.MaxItems = clonePointer(s.MaxItems)
	c.UniqueItems = clonePointer(s.UniqueItems)

	c.Properties = s.Properties.clone(copies)
	c.PatternProperties = cloneSchemaMap(s.PatternProperties, copies)
	c.MinProperties = clonePointer(s.MinProperties)
	c.MaxProperties = clonePointer(s.MaxProperties)
	c.Required = slices.Clone(s.Required)
	c.DependentRequired = cloneDependentRequired(s.DependentRequired)
	c.DependentSchemas = cloneSchemaMap(s.DependentSchemas, copies)
	c.AdditionalProperties = s.AdditionalProperties.clone(copies)
	c.UnevaluatedProperties = s.UnevaluatedProperties.clone(copies)
	c.PropertyNames = s.PropertyNames.clone(copies)

	c.AnyOf = cloneSchemas(s.AnyOf, copies)
	c.AllOf = cloneSchemas(s.AllOf, copies)
	c.OneOf = cloneSchemas(s.OneOf, copies)
	c.Not = s.Not.clone(copies)
	c.If = s.If.clone(copies)
	c.Then = s.Then.clone(copies)
	c.Else = s.Else.clone(copies)

	c.Examples = slices.Clone(s.Examples)
	c.Example = clonePointer(s.Example)

	c.Xml = s.Xml.clone()

	c.Definitions = cloneSchemaMap(s.Definitions, copies)
	c.Defs = cloneSchemaMap(s.Defs, copies)
	return c
}

func (x *Xml) clone() *Xml {
	return clonePointer(x)
}

func cloneSchemas(s []*Schema, copies schemaCopies) []*Schema {
	if s == nil {
		return nil
	}
	c := make([]*Schema, len(s))
	for i, v := range s {
		c[i] = v.clone(copies)
	}
	return c
}

func cloneSchemaMap(m map[string]*Schema, copies schemaCopies) map[string]*Schema {
	if m == nil {
		return nil
	}
	c := make(map[string]*Schema, len(m))
	for k, v := range m {
		c[k] = v.clone(copies)
	}
	return c
}

func cloneDependentRequired(m map[string][]string) map[string][]string {
	if m == nil {
		return nil
	}
	c := make(map[string][]string, len(m))
	for k, v := range m {
		c[k] = slices.Clone(v)
	}
	return c
}

func clonePointer[T any](p *T) *T {
	if p == nil {
		return nil
	}
	c := *p
	return &c
}
