package openapi

func (p *Parameters) Patch(patch Parameters) {
Loop:
	for _, pParam := range patch {
		for _, param := range *p {
			if param.Value != nil && pParam.Value != nil && param.Value.Name == pParam.Value.Name {
				param.Patch(pParam)
				continue Loop
			}
		}
		*p = append(*p, pParam.clone())
	}
}

func (r *ParameterRef) Patch(patch *ParameterRef) {
	if patch == nil || patch.Value == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value.clone()
		return
	}

	r.Value.patch(patch.Value)
}

func (p *Parameter) patch(patch *Parameter) {
	if len(patch.Name) > 0 {
		p.Name = patch.Name
	}

	if patch.Type != "" {
		p.Type = patch.Type
	}

	if patch.Required != nil {
		p.Required = clonePointer(patch.Required)
	}

	if patch.Explode != nil {
		p.Explode = clonePointer(patch.Explode)
	}

	if patch.AllowReserved != nil {
		p.AllowReserved = clonePointer(patch.AllowReserved)
	}

	if len(patch.Description) > 0 {
		p.Description = patch.Description
	}

	if patch.Deprecated != nil {
		p.Deprecated = clonePointer(patch.Deprecated)
	}

	if patch.Style != "" {
		p.Style = patch.Style
	}

	if p.Schema == nil {
		p.Schema = patch.Schema.Clone()
	} else {
		p.Schema.Patch(patch.Schema)
	}

	if p.Content == nil {
		p.Content = patch.Content.clone()
	} else {
		p.Content.patch(patch.Content)
	}
}

func (p Parameters) clone() Parameters {
	return cloneSlice(p, (*ParameterRef).clone)
}

func (r *ParameterRef) clone() *ParameterRef {
	if r == nil {
		return nil
	}
	return &ParameterRef{Reference: r.Reference, Value: r.Value.clone()}
}

func (p *Parameter) clone() *Parameter {
	if p == nil {
		return nil
	}
	c := *p
	c.Schema = p.Schema.Clone()
	c.Required = clonePointer(p.Required)
	c.Deprecated = clonePointer(p.Deprecated)
	c.Explode = clonePointer(p.Explode)
	c.AllowReserved = clonePointer(p.AllowReserved)
	c.Content = p.Content.clone()
	return &c
}
