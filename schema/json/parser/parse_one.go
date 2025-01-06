package parser

import (
	"fmt"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
)

func (p *Parser) ParseOne(s *schema.Schema, data interface{}, evaluated map[string]bool) (interface{}, error) {
	if m, ok := data.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
		p2 := *p
		p2.ValidateAdditionalProperties = true
		return p2.parseOneObject(m, s, evaluated)
	}

	var result interface{}
	validIndex := -1
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

func (p *Parser) parseOneObject(m *sortedmap.LinkedHashMap[string, interface{}], s *schema.Schema, evaluated map[string]bool) (interface{}, error) {
	var result interface{}
	var err error
	validIndex := -1
	index := 0
	var one *schema.Ref

	for index, one = range s.OneOf {
		if one == nil || one.Value == nil {
			if result != nil {
				return nil, Errorf("oneOf", "valid against more than one schema from 'oneOf': valid schema indexes: %v, %v", validIndex, index)
			}
			result = m
			validIndex = index
			continue
		}

		eval := map[string]bool{}
		var next *sortedmap.LinkedHashMap[string, interface{}]
		next, err = p.parseObject(m, one.Value, eval)
		if err != nil {
			continue
		}

		var v interface{}
		v, err = p.evaluateUnevaluatedProperties(next, one.Value, eval)
		if err != nil {
			continue
		}
		next = v.(*sortedmap.LinkedHashMap[string, interface{}])

		if result != nil {
			return nil, Errorf("oneOf", "valid against more than one schema from 'oneOf': valid schema indexes: %v, %v", validIndex, index)
		}

		result = next
		validIndex = index

		for key, val := range eval {
			evaluated[key] = val
		}
	}

	if result == nil {
		pe := &PathCompositionError{
			Path:    fmt.Sprintf("%v", index),
			Message: "valid against no schemas from 'oneOf'",
		}
		pe.append(err)
		return nil, wrapError("oneOf", pe)
	}

	return result, nil
}
