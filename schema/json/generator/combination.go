package generator

import (
	"mokapi/schema/json/schema"
)

func extendBranchWithBase(branch *schema.Schema, base *schema.Schema) (*schema.Schema, error) {
	result := branch.Clone()
	result.Type = schema.Types{"object"}
	var err error

	if result.Properties != nil && base.Properties != nil {
		for it := result.Properties.Iter(); it.Next(); {
			ps := base.Properties.Get(it.Key())
			if ps != nil {
				ps, err = intersectSchema(ps, it.Value())
				if err != nil {
					return nil, err
				}
				result.Properties.Set(it.Key(), ps)
			}
		}
	}

	for _, req := range branch.Required {
		ps := result.Properties.Get(req)
		if ps == nil {
			ps = base.Properties.Get(req)
			if ps != nil {
				if result.Properties == nil {
					result.Properties = new(schema.Schemas)
				}
				result.Properties.Set(req, ps)
			}
		}
	}

	return result, nil
}
