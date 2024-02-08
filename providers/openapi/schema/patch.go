package schema

func (r *Ref) Patch(patch *Ref) {
	if patch == nil || patch.Value == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value
		return
	}

	r.Value.Patch(patch.Value)
}

func (s *Schema) Patch(patch *Schema) {
	if len(patch.Type) > 0 {
		s.Type = patch.Type
	}
	if len(patch.Format) > 0 {
		s.Format = patch.Format
	}
	if len(patch.Pattern) > 0 {
		s.Pattern = patch.Pattern
	}
	if len(patch.Description) > 0 {
		s.Description = patch.Description
	}
	if s.Properties == nil {
		s.Properties = patch.Properties
	} else {
		s.Properties.Patch(patch.Properties)
	}

	if s.Items == nil {
		s.Items = patch.Items
	} else {
		s.Items.Patch(patch.Items)
	}

	if s.Xml == nil {
		s.Xml = patch.Xml
	} else {
		s.Xml.patch(patch.Xml)
	}

	if patch.MinLength != nil {
		s.MinLength = patch.MinLength
	}

	if patch.MaxLength != nil {
		s.MaxLength = patch.MaxLength
	}

	if patch.Enum != nil {
		s.Enum = patch.Enum
	}

	if patch.Example != nil {
		s.Example = patch.Example
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

	if patch.MinItems != nil {
		s.MinItems = patch.MinItems
	}

	if patch.MaxItems != nil {
		s.MaxItems = patch.MaxItems
	}

	if patch.MinProperties != nil {
		s.MinProperties = patch.MinProperties
	}

	if patch.MaxProperties != nil {
		s.MaxProperties = patch.MaxProperties
	}

	if patch.Required != nil {
		s.Required = patch.Required
	}
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
