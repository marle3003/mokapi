package parser

import (
	"fmt"
	"mokapi/schema/json/schema"
)

func (p *Parser) ParseOne(s *schema.Schema, data interface{}) (interface{}, error) {
	var result interface{}

	for _, one := range s.OneOf {
		next, err := p.Parse(data, one)
		if err != nil {
			continue
		}
		if result != nil {
			return nil, fmt.Errorf("parse %v failed: valid against more than one schema from 'oneOf': %s", ToString(data), s)
		}
		result = next
	}

	if result == nil {
		return nil, fmt.Errorf("parse %v failed: valid against no schemas from 'oneOf': %s", ToString(data), s)
	}

	return result, nil
}
