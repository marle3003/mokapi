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
	if len(s.Format) == 0 {
		s.Format = patch.Format
	}
	if len(s.Pattern) == 0 {
		s.Pattern = patch.Pattern
	}
	if len(s.Description) == 0 {
		s.Description = patch.Description
	}
	if s.Properties == nil {
		s.Properties = patch.Properties
	} else {
		s.Properties.patch(patch.Properties)
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
}

func (s *SchemasRef) patch(patch *SchemasRef) {
	if patch == nil || patch.Value == nil {
		return
	}
	if s.Value == nil {
		s.Value = patch.Value
		return
	}
}

func (x *Xml) patch(patch *Xml) {
	if patch == nil {
		return
	}

	if len(x.Name) == 0 {
		x.Name = patch.Name
	}

	if len(x.Prefix) == 0 {
		x.Prefix = patch.Prefix
	}

	if len(x.Namespace) == 0 {
		x.Namespace = patch.Namespace
	}
}

func (s *Schemas) Patch(patch *Schemas) {
	if patch == nil {
		return
	}
	for it := patch.Iter(); it.Next(); {
		r := it.Value().(*Ref)
		name := it.Key().(string)
		if v := s.Get(name); v != nil {
			v.Patch(r)
		} else {
			s.Set(it.Key(), it.Value())
		}
	}
}
