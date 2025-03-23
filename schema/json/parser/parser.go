package parser

import (
	"fmt"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
	"unicode"
)

type Parser struct {
	Schema *schema.Schema

	ConvertStringToNumber  bool
	ConvertStringToBoolean bool
	ConvertToSortedMap     bool
	// To avoid having to filter out additional properties in the JavaScript scripts by the developer,
	// this option is offered in the parser. For example response.data = { ... }
	ValidateAdditionalProperties bool
	// JSON schema: By default, format is just an annotation and does not affect validation.
	SkipValidationFormatKeyword bool
}

func (p *Parser) ParseWith(data interface{}, schema *schema.Schema) (interface{}, error) {
	v, err := p.parse(data, schema)
	if err != nil {
		return v, &Error{err: err}
	}

	return v, nil
}

func (p *Parser) Parse(data interface{}) (interface{}, error) {
	v, err := p.parse(data, p.Schema)
	if err != nil {
		return v, &Error{err: err}
	}

	return v, nil
}

func (p *Parser) parse(data interface{}, s *schema.Schema) (interface{}, error) {
	if s == nil {
		return data, nil
	}
	if s.Boolean != nil {
		if *s.Boolean {
			return data, nil
		}
		return data, &ErrorDetail{
			Message: "schema always fails validation",
			Field:   "valid",
		}
	}

	if data == nil {
		if s.IsNullable() {
			return nil, nil
		} else if s.Default != nil {
			data = s.Default
		}
	}

	evaluatedProperties := map[string]bool{}
	evaluatedItems := map[int]bool{}

	var v interface{}
	var err error
	if len(s.Type) == 0 {
		t := toType(data)
		v, err = p.parseType(data, s, t, evaluatedProperties, evaluatedItems)
		if err != nil {
			return nil, err
		}
	}

	for _, typeName := range s.Type {
		v, err = p.parseType(data, s, typeName, evaluatedProperties, evaluatedItems)
		if err != nil {
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}

	switch {
	case len(s.AnyOf) > 0:
		v, err = p.ParseAny(s, v, evaluatedProperties)
		if err != nil {
			return nil, err
		}
	case len(s.AllOf) > 0:
		v, err = p.ParseAll(s, v, evaluatedProperties)
		if err != nil {
			return nil, err
		}
	case len(s.OneOf) > 0:
		v, err = p.ParseOne(s, v, evaluatedProperties)
		if err != nil {
			return nil, err
		}
	}

	v, err = p.evaluateUnevaluatedProperties(v, s, evaluatedProperties)
	if err != nil {
		return nil, err
	}
	v, err = p.evaluateUnevaluatedItems(v, s, evaluatedItems)
	if err != nil {
		return nil, err
	}

	if m, ok := v.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
		if p.ConvertToSortedMap {
			v = m
		} else {
			v = m.ToMap()
		}
	}

	return v, nil
}

func (p *Parser) parseType(data interface{}, s *schema.Schema, typeName string, evaluatedProperties map[string]bool, evaluatedItems map[int]bool) (interface{}, error) {
	switch data.(type) {
	case []interface{}:
		if typeName != "array" {
			return nil, &ErrorDetail{
				Message: fmt.Sprintf("invalid type, expected %v but got %v", typeName, toType(data)),
				Field:   "type",
			}
		}
	case map[string]interface{}:
		if typeName != "object" {
			return nil, &ErrorDetail{
				Message: fmt.Sprintf("invalid type, expected %v but got %v", typeName, toType(data)),
				Field:   "type",
			}
		}
	case struct{}:
		if typeName != "object" {
			return nil, &ErrorDetail{
				Message: fmt.Sprintf("invalid type, expected %v but got %v", typeName, toType(data)),
				Field:   "type",
			}
		}
	}

	var err error
	switch typeName {
	case "string":
		data, err = p.ParseString(data, s)
	case "number":
		data, err = p.ParseNumber(data, s)
	case "integer":
		data, err = p.ParseInteger(data, s)
	case "boolean":
		data, err = p.parseBoolean(data, s)
	case "array":
		data, err = p.ParseArray(data, s, evaluatedItems)
	case "object":
		data, err = p.parseObject(data, s, evaluatedProperties)
	}

	if s.Const != nil {
		s2 := *s
		s2.Const = nil
		p2 := Parser{ConvertToSortedMap: true}
		c, constErr := p2.parse(*s.Const, &s2)
		if constErr != nil {
			return data, &ErrorDetail{
				Message: fmt.Sprintf("const value does not match schema: %v", constErr),
				Field:   "const",
			}
		}
		if !compare(data, c) {
			return data, &ErrorDetail{
				Message: fmt.Sprintf("value '%v' does not match const '%v'", ToString(data), ToString(c)),
				Field:   "const",
			}
		}
	}

	if s.Not != nil {
		if _, notErr := p.parse(data, s.Not); notErr == nil {
			return nil, &ErrorDetail{
				Message: "is valid against schema",
				Field:   "not",
			}
		}
	}

	return data, err
}

// UnTitle returns a copy of the string s with first letter mapped to its Unicode lower case
func unTitle(s string) string {
	for i, v := range s {
		return string(unicode.ToLower(v)) + s[i+1:]
	}
	return s
}

func (p *Parser) evaluateUnevaluatedProperties(data interface{}, schema *schema.Schema, evaluatedProperties map[string]bool) (interface{}, error) {
	if schema.UnevaluatedProperties == nil {
		return data, nil
	}
	var err ErrorList

	if object, ok := data.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
		for it := object.Iter(); it.Next(); {
			name := it.Key()
			val := it.Value()
			if _, evaluated := evaluatedProperties[name]; !evaluated {
				if schema.UnevaluatedProperties.Boolean != nil && !*schema.UnevaluatedProperties.Boolean {
					err = append(err, &ErrorDetail{
						Message: fmt.Sprintf("property %s not successfully evaluated and schema does not allow unevaluated properties", name),
						Field:   "unevaluatedProperties",
					})
				} else {
					v, evalErr := p.parse(val, schema.UnevaluatedProperties)
					if evalErr != nil {
						err = append(err, wrapErrorDetail(evalErr, &ErrorDetail{Field: "unevaluatedProperties"}))
					} else {
						object.Set(name, v)
					}
				}
			}
		}
	}

	if len(err) > 0 {
		return nil, &err
	}

	return data, nil
}

func (p *Parser) evaluateUnevaluatedItems(data interface{}, schema *schema.Schema, evaluatedItems map[int]bool) (interface{}, error) {
	if schema.UnevaluatedItems == nil {
		return data, nil
	}
	var err ErrorList

	if arr, ok := data.([]interface{}); ok {
		for i, val := range arr {
			if _, evaluated := evaluatedItems[i]; !evaluated {
				if schema.UnevaluatedItems.Boolean != nil && !*schema.UnevaluatedItems.Boolean {
					err = append(err, &ErrorDetail{
						Message: fmt.Sprintf("item at index %v has not been successfully evaluated and the schema does not allow unevaluated items", i),
						Field:   "unevaluatedItems",
					})
				} else {
					v, evalErr := p.parse(val, schema.UnevaluatedItems)
					if evalErr != nil {
						err = append(err, wrapErrorDetail(
							evalErr,
							&ErrorDetail{
								Message: fmt.Sprintf("item at index %v has not been successfully evaluated and the schema does not allow unevaluated items", i),
								Field:   "unevaluatedItems",
							},
						))
					} else {
						arr[i] = v
					}
				}
			}
		}
	}

	if len(err) > 0 {
		return nil, &err
	}

	return data, nil
}
