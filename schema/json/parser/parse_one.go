package parser

import (
	"fmt"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
)

func (p *Parser) ParseOne(s *schema.Schema, data interface{}) (interface{}, error) {
	var result interface{}

	for _, one := range s.OneOf {
		if one == nil || one.Value == nil {
			next, err := p.Parse(data, nil)
			if err != nil {
				continue
			}
			if result != nil {
				return nil, fmt.Errorf("oneOf can only match exactly one schema")
			}
			result = next
			continue
		}

		next, err := p.Parse(data, one)
		if err != nil {
			continue
		}
		if obj, ok := next.(*sortedmap.LinkedHashMap[string, interface{}]); ok && obj.Len() == 0 {
			// empty object does not match
			continue
		}
		if result != nil {
			return nil, fmt.Errorf("parse %v failed: it is valid for more than one schema, expected %v", toString(data), s)
		}
		result = next
	}

	if result == nil {
		return nil, fmt.Errorf("parse %v failed: expected to match one of schema but it matches none", toString(data))
	}

	return result, nil
}
