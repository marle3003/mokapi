package parameter

func (p *Parameters) Patch(patch Parameters) {
Loop:
	for _, pParam := range patch {
		for _, param := range *p {
			if param.Value != nil && pParam.Value != nil && param.Value.Name == pParam.Value.Name {
				param.Patch(pParam)
				continue Loop
			}
		}
		*p = append(*p, pParam)
	}
}

func (r *Ref) Patch(patch *Ref) {
	if r.Value == nil {
		r.Value = patch.Value
		return
	}

	if len(patch.Value.Description) > 0 {
		r.Value.Description = patch.Value.Description
	}

	r.Value.Deprecated = patch.Value.Deprecated

	if r.Value.Schema == nil {
		r.Value.Schema = patch.Value.Schema
	} else {
		r.Value.Schema.Patch(patch.Value.Schema)
	}
}
