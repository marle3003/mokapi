package generator

import (
	"mokapi/schema/json/schema"
)

func applyObject(sub, base *schema.Schema) *schema.Schema {
	result := sub.Clone()
	result.Type = schema.Types{"object"}

	if result.Properties != nil && base.Properties != nil {
		for it := result.Properties.Iter(); it.Next(); {
			ps := base.Properties.Get(it.Key())
			if ps != nil {
				ps = mergeSchema(ps, it.Value())
				result.Properties.Set(it.Key(), ps)
			}
		}
	}

	for _, req := range sub.Required {
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

	return result
}
