package parser

import (
	"fmt"
	"maps"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
)

func (p *Parser) ParseOne(s *schema.Schema, data any, evaluated map[string]bool) (any, error) {
	if m, ok := data.(*sortedmap.LinkedHashMap[string, any]); ok {
		p2 := *p
		p2.ValidateAdditionalProperties = true
		return p2.parseOneObject(m, s, evaluated)
	}

	var result any
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

func (p *Parser) parseOneObject(m *sortedmap.LinkedHashMap[string, any], s *schema.Schema, evaluated map[string]bool) (any, error) {
	var result any
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
		var next *sortedmap.LinkedHashMap[string, any]
		next, err = p.parseObjectWithComposition(m, one, eval)
		if err != nil {
			continue
		}

		var v any
		v, err = p.evaluateUnevaluatedProperties(next, one, eval)
		if err != nil {
			continue
		}
		next = v.(*sortedmap.LinkedHashMap[string, any])

		if result != nil {
			return nil, &ErrorDetail{
				Message: fmt.Sprintf("valid against more than one schema: valid schema indexes: %v, %v", validIndex, index),
				Field:   "oneOf",
			}
		}

		result = next
		validIndex = index

		maps.Copy(evaluated, eval)
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
