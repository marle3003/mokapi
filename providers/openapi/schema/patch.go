package schema

import (
	"slices"

	jsonSchema "mokapi/schema/json/schema"
)

func (s *Schema) Patch(patch *Schema) {
	s.patch(patch, patchedPairs{})
}

// patchedPairs records which (target, patch) pairs are already patched. Resolved
// references make schema graphs cyclic, so a pair can be reached again through its own
// properties; patching it a second time would never end.
type patchedPairs map[[2]*Schema]struct{}

// add returns false if the pair was already recorded.
func (p patchedPairs) add(s, patch *Schema) bool {
	pair := [2]*Schema{s, patch}
	if _, ok := p[pair]; ok {
		return false
	}
	p[pair] = struct{}{}
	return true
}

func (s *Schema) patch(patch *Schema, visited patchedPairs) {
	if patch == nil || s == patch || !visited.add(s, patch) {
		return
	}

	if patch.Id != "" {
		s.Id = patch.Id
	}

	if patch.Boolean != nil {
		s.Boolean = clonePointer(patch.Boolean)
	}

	if patch.Schema != "" {
		s.Schema = patch.Schema
	}

	if patch.Anchor != "" {
		s.Anchor = patch.Anchor
	}

	if patch.DynamicAnchor != "" {
		s.DynamicAnchor = patch.DynamicAnchor
	}

	if len(patch.Type) > 0 {
		s.Type = mergeTypes(s.Type, patch.Type)
	}
	if len(patch.AnyOf) > 0 {
		if len(s.AnyOf) == 0 {
			s.AnyOf = cloneSchemas(patch.AnyOf, schemaCopies{})
		} else {
			s.AnyOf = patchComposition(s.AnyOf, patch.AnyOf, visited)
		}
	}
	if len(patch.AllOf) > 0 {
		if len(s.AllOf) == 0 {
			s.AllOf = cloneSchemas(patch.AllOf, schemaCopies{})
		} else {
			s.AllOf = patchComposition(s.AllOf, patch.AllOf, visited)
		}
	}
	if len(patch.OneOf) > 0 {
		if len(s.OneOf) == 0 {
			s.OneOf = cloneSchemas(patch.OneOf, schemaCopies{})
		} else {
			s.OneOf = patchComposition(s.OneOf, patch.OneOf, visited)
		}
	}

	s.Not = patchSchema(s.Not, patch.Not, visited)
	s.If = patchSchema(s.If, patch.If, visited)
	s.Then = patchSchema(s.Then, patch.Then, visited)
	s.Else = patchSchema(s.Else, patch.Else, visited)

	if patch.Enum != nil {
		s.Enum = slices.Clone(patch.Enum)
	}
	if patch.Const != nil {
		s.Const = clonePointer(patch.Const)
	}
	if s.Xml == nil {
		s.Xml = patch.Xml.clone()
	} else {
		s.Xml.patch(patch.Xml)
	}
	if len(patch.Format) > 0 {
		s.Format = patch.Format
	}

	s.Nullable = patch.Nullable

	if len(patch.Pattern) > 0 {
		s.Pattern = patch.Pattern
	}
	if patch.MinLength != nil {
		s.MinLength = clonePointer(patch.MinLength)
	}
	if patch.MaxLength != nil {
		s.MaxLength = clonePointer(patch.MaxLength)
	}
	if patch.MultipleOf != nil {
		s.MultipleOf = clonePointer(patch.MultipleOf)
	}
	if patch.Minimum != nil {
		s.Minimum = clonePointer(patch.Minimum)
	}
	if patch.Maximum != nil {
		s.Maximum = clonePointer(patch.Maximum)
	}
	if patch.ExclusiveMinimum != nil {
		s.ExclusiveMinimum = clonePointer(patch.ExclusiveMinimum)
	}
	if patch.ExclusiveMaximum != nil {
		s.ExclusiveMaximum = clonePointer(patch.ExclusiveMaximum)
	}
	if s.Items == nil {
		s.Items = patch.Items.Clone()
	} else {
		s.Items.patch(patch.Items, visited)
	}
	// prefixItems are positional, so they are merged by index and not by title
	for i, p := range patch.PrefixItems {
		if i < len(s.PrefixItems) {
			s.PrefixItems[i] = patchSchema(s.PrefixItems[i], p, visited)
		} else {
			s.PrefixItems = append(s.PrefixItems, p.Clone())
		}
	}
	s.UnevaluatedItems = patchSchema(s.UnevaluatedItems, patch.UnevaluatedItems, visited)
	s.Contains = patchSchema(s.Contains, patch.Contains, visited)
	if patch.MaxContains != nil {
		s.MaxContains = clonePointer(patch.MaxContains)
	}
	if patch.MinContains != nil {
		s.MinContains = clonePointer(patch.MinContains)
	}

	if patch.UniqueItems != nil {
		s.UniqueItems = clonePointer(patch.UniqueItems)
	}

	if patch.MinItems != nil {
		s.MinItems = clonePointer(patch.MinItems)
	}
	if patch.MaxItems != nil {
		s.MaxItems = clonePointer(patch.MaxItems)
	}

	s.ShuffleItems = patch.ShuffleItems

	if s.Properties == nil {
		s.Properties = patch.Properties.Clone()
	} else {
		s.Properties.patch(patch.Properties, visited)
	}

	s.PatternProperties = patchSchemaMap(s.PatternProperties, patch.PatternProperties, visited)

	if patch.Required != nil {
		s.Required = slices.Clone(patch.Required)
	}
	if s.DependentRequired == nil {
		s.DependentRequired = cloneDependentRequired(patch.DependentRequired)
	} else {
		for k, v := range patch.DependentRequired {
			s.DependentRequired[k] = slices.Clone(v)
		}
	}
	s.DependentSchemas = patchSchemaMap(s.DependentSchemas, patch.DependentSchemas, visited)
	if s.AdditionalProperties == nil {
		s.AdditionalProperties = patch.AdditionalProperties.Clone()
	} else {
		s.AdditionalProperties.patch(patch.AdditionalProperties, visited)
	}
	s.UnevaluatedProperties = patchSchema(s.UnevaluatedProperties, patch.UnevaluatedProperties, visited)
	s.PropertyNames = patchSchema(s.PropertyNames, patch.PropertyNames, visited)
	if patch.MinProperties != nil {
		s.MinProperties = clonePointer(patch.MinProperties)
	}
	if patch.MaxProperties != nil {
		s.MaxProperties = clonePointer(patch.MaxProperties)
	}
	if len(patch.Title) > 0 {
		s.Title = patch.Title
	}
	if len(patch.Description) > 0 {
		s.Description = patch.Description
	}
	if patch.Default != nil {
		s.Default = patch.Default
	}

	s.Deprecated = patch.Deprecated

	if patch.Examples != nil {
		s.Examples = slices.Clone(patch.Examples)
	}
	if patch.Example != nil {
		s.Example = clonePointer(patch.Example)
	}
	if len(patch.ContentMediaType) > 0 {
		s.ContentMediaType = patch.ContentMediaType
	}
	if len(patch.ContentEncoding) > 0 {
		s.ContentEncoding = patch.ContentEncoding
	}

	s.Definitions = patchSchemaMap(s.Definitions, patch.Definitions, visited)
	s.Defs = patchSchemaMap(s.Defs, patch.Defs, visited)
}

// patchSchema merges patch into s and returns the result; a missing s is replaced by a copy of patch.
func patchSchema(s, patch *Schema, visited patchedPairs) *Schema {
	if s == nil {
		return patch.Clone()
	}
	s.patch(patch, visited)
	return s
}

func patchSchemaMap(s, patch map[string]*Schema, visited patchedPairs) map[string]*Schema {
	if s == nil {
		return cloneSchemaMap(patch, schemaCopies{})
	}
	for k, v := range patch {
		if def, ok := s[k]; ok {
			def.patch(v, visited)
		} else {
			s[k] = v.Clone()
		}
	}
	return s
}

func (x *Xml) patch(patch *Xml) {
	if patch == nil {
		return
	}

	if len(patch.Name) > 0 {
		x.Name = patch.Name
	}

	if len(patch.Prefix) > 0 {
		x.Prefix = patch.Prefix
	}

	if len(patch.Namespace) > 0 {
		x.Namespace = patch.Namespace
	}

	x.Wrapped = patch.Wrapped
	x.Attribute = patch.Attribute
}

func (s *Schemas) Patch(patch *Schemas) {
	s.patch(patch, patchedPairs{})
}

func (s *Schemas) patch(patch *Schemas, visited patchedPairs) {
	if patch == nil {
		return
	}
	for it := patch.Iter(); it.Next(); {
		r := it.Value()
		name := it.Key()
		if v := s.Get(name); v != nil {
			v.patch(r, visited)
		} else {
			s.Set(it.Key(), r.Clone())
		}
	}
}

func mergeTypes(origin, patch jsonSchema.Types) jsonSchema.Types {
	// a copy, so appending never writes into the backing array of origin
	merged := slices.Clone(origin)
	for _, t := range patch {
		if !slices.Contains(merged, t) {
			merged = append(merged, t)
		}
	}
	return merged
}

func patchComposition(s []*Schema, patch []*Schema, visited patchedPairs) []*Schema {
Patch:
	for _, p := range patch {
		if p == nil {
			continue
		}
		if p.Title == "" {
			s = append(s, p.Clone())
		} else {
			for _, r := range s {
				if r.Title == p.Title {
					r.patch(p, visited)
					continue Patch
				}
			}
			s = append(s, p.Clone())
		}
	}
	return s
}
