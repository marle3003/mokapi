package schema

import "slices"

// Patch applies patch to s. Values s adopts from patch are copied, so s never
// shares schemas, maps or slices with patch.
func (s *Schema) Patch(patch *Schema) {
	s.patch(patch, schemaCopies{}, patchedPairs{})
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

func (s *Schema) patch(patch *Schema, copies schemaCopies, visited patchedPairs) {
	if patch == nil || !visited.add(s, patch) {
		return
	}

	if patch.Id != "" {
		s.Id = patch.Id
	}

	if patch.Boolean != nil {
		s.Boolean = patch.Boolean
	}

	if patch.Anchor != "" {
		s.Anchor = patch.Anchor
	}

	if len(patch.Type) > 0 {
		s.Type = mergeTypes(s.Type, patch.Type)
	}
	if patch.Enum != nil {
		s.Enum = patch.Enum
	}
	if patch.Const != nil {
		s.Const = patch.Const
	}
	if len(patch.Format) > 0 {
		s.Format = patch.Format
	}

	if len(patch.Pattern) > 0 {
		s.Pattern = patch.Pattern
	}
	if patch.MinLength != nil {
		s.MinLength = patch.MinLength
	}
	if patch.MaxLength != nil {
		s.MaxLength = patch.MaxLength
	}
	if patch.MultipleOf != nil {
		s.MultipleOf = patch.MultipleOf
	}
	if patch.Minimum != nil {
		s.Minimum = patch.Minimum
	}
	if patch.Maximum != nil {
		s.Maximum = patch.Maximum
	}
	if patch.ExclusiveMinimum != nil {
		s.ExclusiveMinimum = patch.ExclusiveMinimum
	}
	if patch.ExclusiveMaximum != nil {
		s.ExclusiveMaximum = patch.ExclusiveMaximum
	}
	if s.Items == nil {
		s.Items = patch.Items.clone(copies)
	} else {
		s.Items.patch(patch.Items, copies, visited)
	}

	s.UniqueItems = patch.UniqueItems

	if patch.MinItems != nil {
		s.MinItems = patch.MinItems
	}
	if patch.MaxItems != nil {
		s.MaxItems = patch.MaxItems
	}

	s.ShuffleItems = patch.ShuffleItems

	if s.Properties == nil {
		s.Properties = patch.Properties.clone(copies)
	} else {
		s.Properties.patch(patch.Properties, copies, visited)
	}

	if patch.Required != nil {
		s.Required = slices.Clone(patch.Required)
	}

	if s.AdditionalProperties == nil {
		s.AdditionalProperties = patch.AdditionalProperties.clone(copies)
	} else {
		s.AdditionalProperties.patch(patch.AdditionalProperties, copies, visited)
	}

	if patch.MinProperties != nil {
		s.MinProperties = patch.MinProperties
	}
	if patch.MaxProperties != nil {
		s.MaxProperties = patch.MaxProperties
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
		s.Examples = patch.Examples
	}
	if len(patch.ContentMediaType) > 0 {
		s.ContentMediaType = patch.ContentMediaType
	}
	if len(patch.ContentEncoding) > 0 {
		s.ContentEncoding = patch.ContentEncoding
	}

	s.Definitions = patchSchemaMap(s.Definitions, patch.Definitions, copies, visited)
	s.Defs = patchSchemaMap(s.Defs, patch.Defs, copies, visited)
}

func patchSchemaMap(m, patch map[string]*Schema, copies schemaCopies, visited patchedPairs) map[string]*Schema {
	if m == nil {
		return cloneSchemaMap(patch, copies)
	}
	for k, v := range patch {
		if s, ok := m[k]; ok {
			s.patch(v, copies, visited)
		} else {
			m[k] = v.clone(copies)
		}
	}
	return m
}

// Patch applies patch to s, see Schema.Patch.
func (s *Schemas) Patch(patch *Schemas) {
	s.patch(patch, schemaCopies{}, patchedPairs{})
}

func (s *Schemas) patch(patch *Schemas, copies schemaCopies, visited patchedPairs) {
	if patch == nil {
		return
	}
	for it := patch.Iter(); it.Next(); {
		r := it.Value()
		name := it.Key()
		if v := s.Get(name); v != nil {
			v.patch(r, copies, visited)
		} else {
			s.Set(name, r.clone(copies))
		}
	}
}

func mergeTypes(origin, patch Types) Types {
	m := map[string]struct{}{}
	for _, t := range origin {
		m[t] = struct{}{}
	}
	for _, t := range patch {
		if _, exists := m[t]; !exists {
			origin = append(origin, t)
		}
	}
	return origin
}
