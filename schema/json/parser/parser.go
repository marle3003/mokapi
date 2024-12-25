package parser

import (
	"fmt"
	"mokapi/schema/json/schema"
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
	if ref == nil || ref.Value == nil {
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

	evaluated := map[string]bool{}

	switch {
	case len(ref.Value.AnyOf) > 0:
		v, err := p.ParseAny(ref.Value, data, evaluated)
		if err != nil {
			return nil, err
		}
		data = v
	case len(ref.Value.AllOf) > 0:
		v, err := p.ParseAll(ref.Value, data, evaluated)
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
		return p.parseType(data, ref.Value, t, evaluated)
	}

	var v interface{}
	var err error
	for _, typeName := range ref.Value.Type {
		v, err = p.parseType(data, ref.Value, typeName, evaluated)
		if err != nil {
			continue
		}
		v, err = p.evaluateUnevaluatedProperties(v, ref.Value, evaluated)
		if err == nil {
			return v, nil
		}
	}

	return nil, err
}

func (p *Parser) parseType(data interface{}, schema *schema.Schema, typeName string, evaluated map[string]bool) (interface{}, error) {
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

	switch typeName {
	case "string":
		return p.ParseString(data, schema)
	case "number":
		return p.ParseNumber(data, schema)
	case "integer":
		return p.ParseInteger(data, schema)
	case "boolean":
		return p.ParseBoolean(data, schema)
	case "array":
		return p.ParseArray(data, schema)
	case "object":
		obj, err := p.parseObject(data, schema, evaluated)
		if err != nil {
			return nil, err
		}
		if p.ConvertToSortedMap {
			return obj, nil
		}
		return obj.ToMap(), nil
	}

	return data, nil
}

// UnTitle returns a copy of the string s with first letter mapped to its Unicode lower case
func unTitle(s string) string {
	for i, v := range s {
		return string(unicode.ToLower(v)) + s[i+1:]
	}
	return s
}

func (p *Parser) evaluateUnevaluatedProperties(data interface{}, schema *schema.Schema, evaluated map[string]bool) (interface{}, error) {
	if schema.UnevaluatedProperties == nil {
		return data, nil
	}
	var err PathErrors

	if object, ok := data.(map[string]interface{}); ok {
		for name, val := range object {
			if _, evaluated := evaluated[name]; !evaluated {
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
