package parameter

func (p *Parameters) Patch(patch Parameters) {
Loop:
	for _, pParam := range patch {
		for _, param := range *p {
			if param.Value != nil && pParam.Value != nil && param.Value.Name == pParam.Value.Name {
				param.Value.Patch(pParam.Value)
				continue Loop
			}
		}
		*p = append(*p, pParam)
	}
}

func (p *Parameter) Patch(patch *Parameter) {
	if len(p.Description) == 0 {
		p.Description = patch.Description
	}
	if !p.Deprecated {
		p.Deprecated = patch.Deprecated
	}
	if p.Schema == nil {
		p.Schema = patch.Schema
	} else {
		p.Schema.Patch(patch.Schema)
	}
}

func (r *NamedParameters) Patch(patch *NamedParameters) {
	if patch == nil || patch.Value == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value
	}
	for k, p := range patch.Value {
		if p.Value == nil {
			continue
		}
		if v, ok := r.Value[k]; ok {
			if v.Value == nil {
				v.Value = p.Value
			} else {
				v.Value.Patch(p.Value)
			}
		} else {
			r.Value[k] = p
		}
	}
}
