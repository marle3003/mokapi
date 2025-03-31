package schema

import jsonSchema "mokapi/schema/json/schema"

func (s *Schema) Patch(patch *Schema) {
	if patch == nil {
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
	if len(patch.AnyOf) > 0 {
		if len(s.AnyOf) == 0 {
			s.AnyOf = patch.AnyOf
		} else {
			s.AnyOf = patchComposition(s.AnyOf, patch.AnyOf)
		}
	}
	if len(patch.AllOf) > 0 {
		if len(s.AllOf) == 0 {
			s.AllOf = patch.AllOf
		} else {
			s.AllOf = patchComposition(s.AllOf, patch.AllOf)
		}
	}
	if len(patch.OneOf) > 0 {
		if len(s.OneOf) == 0 {
			s.OneOf = patch.OneOf
		} else {
			s.OneOf = patchComposition(s.OneOf, patch.OneOf)
		}
	}

	if patch.Enum != nil {
		s.Enum = patch.Enum
	}
	if patch.Const != nil {
		s.Const = patch.Const
	}
	if s.Xml == nil {
		s.Xml = patch.Xml
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
		s.Items = patch.Items
	} else {
		s.Items.Patch(patch.Items)
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
		s.Properties = patch.Properties
	} else {
		s.Properties.Patch(patch.Properties)
	}

	if patch.Required != nil {
		s.Required = patch.Required
	}
	if s.AdditionalProperties == nil {
		s.AdditionalProperties = patch.AdditionalProperties
	} else {
		s.AdditionalProperties.Patch(patch.AdditionalProperties)
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
	if patch.Example != nil {
		s.Example = patch.Example
	}
	if len(patch.ContentMediaType) > 0 {
		s.ContentMediaType = patch.ContentMediaType
	}
	if len(patch.ContentEncoding) > 0 {
		s.ContentEncoding = patch.ContentEncoding
	}

	if s.Definitions == nil {
		s.Definitions = patch.Definitions
	} else {
		for k, v := range patch.Definitions {
			if def, ok := s.Definitions[k]; ok {
				def.Patch(v)
			} else {
				s.Definitions[k] = v
			}
		}
	}

	if s.Defs == nil {
		s.Defs = patch.Defs
	} else {
		for k, v := range patch.Defs {
			if def, ok := s.Defs[k]; ok {
				def.Patch(v)
			} else {
				s.Defs[k] = v
			}
		}
	}

	s.cm.Notify(s)
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
	if patch == nil {
		return
	}
	for it := patch.Iter(); it.Next(); {
		r := it.Value()
		name := it.Key()
		if v := s.Get(name); v != nil {
			v.Patch(r)
		} else {
			s.Set(it.Key(), it.Value())
		}
	}
}

func mergeTypes(origin, patch jsonSchema.Types) jsonSchema.Types {
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

func patchComposition(s []*Schema, patch []*Schema) []*Schema {
Patch:
	for _, p := range patch {
		if p == nil {
			continue
		}
		if p.Title == "" {
			s = append(s, p)
		} else {
			for _, r := range s {
				if r.Title == p.Title {
					r.Patch(p)
					continue Patch
				}
			}
			s = append(s, p)
		}
	}
	return s
}
