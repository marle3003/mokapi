package parser

import (
	"mokapi/schema/json/schema"
)

func (p *Parser) ParseOne(s *schema.Schema, data interface{}) (interface{}, error) {
	var result interface{}
	validIndex := -1

	// TODO evaluatedProperties

	for index, one := range s.OneOf {
		next, err := p.parse(data, one)
		if err != nil {
			continue
		}
		if result != nil {
			return nil, Errorf("oneOf", "valid against more than one schema from 'oneOf': valid schema indexes: %v, %v", validIndex, index)
		}
		result = next
		validIndex = index
	}

	if result == nil {
		return nil, Errorf("oneOf", "valid against no schemas from 'oneOf'")
	}

	return result, nil
}
