package parser

import (
	"fmt"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
	"unicode"
)

type Parser struct {
	ConvertStringToNumber  bool
	ConvertStringToBoolean bool
	ConvertToSortedMap     bool
	// To avoid having to filter out additional properties in the JavaScript scripts by the developer,
	// this option is offered in the parser. For example response.data = { ... }
	ValidateAdditionalProperties bool
	// JSON schema: By default, format is just an annotation and does not affect validation.
	SkipValidationFormatKeyword bool
}

func (p *Parser) Parse(data interface{}, ref *schema.Ref) (interface{}, error) {
	v, err := p.parse(data, ref)
	if err != nil {
		return v, &Error{
			NumErrors: NumErrors(err),
			Err:       err,
		}
	}
	return v, nil
}

func (p *Parser) parse(data interface{}, ref *schema.Ref) (interface{}, error) {
	if ref == nil {
		return data, nil
	}
	if ref.Boolean != nil {
		if *ref.Boolean {
			return data, nil
		}
		return data, Errorf("valid", "schema always fails validation")
	}
	if ref.Value == nil {
		return data, nil
	}

	if data == nil {
		if ref.Value.IsNullable() {
			return nil, nil
		} else if ref.Value.Default != nil {
			data = ref.Value.Default
		} else {
			return nil, fmt.Errorf("parse NULL failed, expected %v", ref)
		}
	}

	evaluatedProperties := map[string]bool{}
	evaluatedItems := map[int]bool{}

	switch {
	case len(ref.Value.AnyOf) > 0:
		v, err := p.ParseAny(ref.Value, data, evaluatedProperties)
		if err != nil {
			return nil, err
		}
		data = v
	case len(ref.Value.AllOf) > 0:
		v, err := p.ParseAll(ref.Value, data, evaluatedProperties)
		if err != nil {
			return nil, err
		}
		data = v
	case len(ref.Value.OneOf) > 0:
		v, err := p.ParseOne(ref.Value, data)
		if err != nil {
			return nil, err
		}
		data = v
	}

	if len(ref.Value.Type) == 0 {
		t := toType(data)
		return p.parseType(data, ref.Value, t, evaluatedProperties, evaluatedItems)
	}

	var v interface{}
	var err error
	for _, typeName := range ref.Value.Type {
		v, err = p.parseType(data, ref.Value, typeName, evaluatedProperties, evaluatedItems)
		if err != nil {
			continue
		}
		v, err = p.evaluateUnevaluatedProperties(v, ref.Value, evaluatedProperties)
		if err != nil {
			continue
		}
		v, err = p.evaluateUnevaluatedItems(v, ref.Value, evaluatedItems)
		if err != nil {
			continue
		}
		return v, nil
	}

	return nil, err
}

func (p *Parser) parseType(data interface{}, s *schema.Schema, typeName string, evaluatedProperties map[string]bool, evaluatedItems map[int]bool) (interface{}, error) {
	switch data.(type) {
	case []interface{}:
		if typeName != "array" {
			return nil, Errorf("type", "invalid type, expected %v but got %v", typeName, toType(data))
		}
	case map[string]interface{}:
		if typeName != "object" {
			return nil, Errorf("type", "invalid type, expected %v but got %v", typeName, toType(data))
		}
	case struct{}:
		if typeName != "object" {
			return nil, Errorf("type", "invalid type, expected %v but got %v", typeName, toType(data))
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
		var m *sortedmap.LinkedHashMap[string, interface{}]
		m, err = p.parseObject(data, s, evaluatedProperties)
		if err != nil {
			return nil, err
		}
		if p.ConvertToSortedMap {
			data = m
		} else {
			data = m.ToMap()
		}
	}

	if s.Const != nil {
		s2 := *s
		s2.Const = nil
		c, constErr := p.parse(*s.Const, &schema.Ref{Value: &s2})
		if constErr != nil {
			return data, Errorf("const", "const value does not match schema: %v", constErr)
		}
		if !compare(data, c) {
			return data, Errorf("const", "value '%v' does not match const '%v'", ToString(data), ToString(c))
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
	var err PathErrors

	if object, ok := data.(map[string]interface{}); ok {
		for name, val := range object {
			if _, evaluated := evaluatedProperties[name]; !evaluated {
				if schema.UnevaluatedProperties.Boolean != nil && !*schema.UnevaluatedProperties.Boolean {
					err = append(err, Errorf("unevaluatedProperties", "property %s not successfully evaluated and schema does not allow unevaluated properties", name))
				} else {
					v, evalErr := p.parse(val, schema.UnevaluatedProperties)
					if evalErr != nil {
						err = append(err, wrapError("unevaluatedProperties", evalErr))
					} else {
						object[name] = v
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
	var err PathErrors

	if arr, ok := data.([]interface{}); ok {
		for i, val := range arr {
			if _, evaluated := evaluatedItems[i]; !evaluated {
				if schema.UnevaluatedItems.Boolean != nil && !*schema.UnevaluatedItems.Boolean {
					err = append(err, Errorf("unevaluatedItems", "item at index %v has not been successfully evaluated and the schema does not allow unevaluated items.", i))
				} else {
					v, evalErr := p.parse(val, schema.UnevaluatedItems)
					if evalErr != nil {
						err = append(err, wrapError("unevaluatedItems", evalErr))
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
