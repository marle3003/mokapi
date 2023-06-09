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
