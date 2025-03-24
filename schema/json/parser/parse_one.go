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
			return nil, &ErrorDetail{
				Message: fmt.Sprintf("valid against more than one schema: valid schema indexes: %v, %v", validIndex, index),
				Field:   "oneOf",
			}
		}
		result = next
		validIndex = index
	}

	if result == nil {
		return nil, &ErrorDetail{
			Message: "valid against no schemas",
			Field:   "oneOf",
		}
	}
	return result, nil
}

func (p *Parser) parseOneObject(m *sortedmap.LinkedHashMap[string, interface{}], s *schema.Schema, evaluated map[string]bool) (interface{}, error) {
	var result interface{}
	var err error
	validIndex := -1
	index := 0
	var one *schema.Schema

	for index, one = range s.OneOf {
		if one == nil {
			if result != nil {
				return nil, &ErrorDetail{
					Message: fmt.Sprintf("valid against more than one schema: valid schema indexes: %v, %v", validIndex, index),
					Field:   "oneOf",
				}
			}
			result = m
			validIndex = index
			continue
		}

		eval := map[string]bool{}
		var next *sortedmap.LinkedHashMap[string, interface{}]
		next, err = p.parseObject(m, one, eval)
		if err != nil {
			continue
		}

		var v interface{}
		v, err = p.evaluateUnevaluatedProperties(next, one, eval)
		if err != nil {
			continue
		}
		next = v.(*sortedmap.LinkedHashMap[string, interface{}])

		if result != nil {
			return nil, &ErrorDetail{
				Message: fmt.Sprintf("valid against more than one schema: valid schema indexes: %v, %v", validIndex, index),
				Field:   "oneOf",
			}
		}

		result = next
		validIndex = index

		for key, val := range eval {
			evaluated[key] = val
		}
	}

	if result == nil {
		return nil, &ErrorDetail{
			Message: "valid against no schemas",
			Field:   "oneOf",
			Errors: ErrorList{
				wrapErrorDetail(err, &ErrorDetail{Field: fmt.Sprintf("%d", index)}),
			},
		}
	}

	return result, nil
}
