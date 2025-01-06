package schema

func (r *Ref) Patch(patch *Ref) {
	if patch == nil {
		return
	}
	if patch.Boolean != nil {
		r.Boolean = patch.Boolean
	}
	if r.Value == nil {
		r.Value = patch.Value
	} else {
		r.Value.Patch(patch.Value)
	}
}

func (s *Schema) Patch(patch *Schema) {
	if patch == nil {
		return
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
	if len(patch.ContentMediaType) > 0 {
		s.ContentMediaType = patch.ContentMediaType
	}
	if len(patch.ContentEncoding) > 0 {
		s.ContentEncoding = patch.ContentEncoding
	}
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
