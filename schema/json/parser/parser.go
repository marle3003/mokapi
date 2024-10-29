package parser

import (
	"fmt"
	"mokapi/schema/json/schema"
	"unicode"
)

type Parser struct {
	ConvertStringToNumber        bool
	ConvertStringToBoolean       bool
	ConvertToSortedMap           bool
	ValidateAdditionalProperties bool
}

func (p *Parser) Parse(data interface{}, ref *schema.Ref) (interface{}, error) {
	if ref == nil || ref.Value == nil {
		return data, nil
	}

	if data == nil {
		if ref.Value.IsNullable() {
			return nil, nil
		}
		return nil, fmt.Errorf("parse NULL failed, expected %v", ref)
	}

	switch {
	case len(ref.Value.AnyOf) > 0:
		return p.ParseAny(ref.Value, data)
	case len(ref.Value.AllOf) > 0:
		return p.ParseAll(ref.Value, data)
	case len(ref.Value.OneOf) > 0:
		return p.ParseOne(ref.Value, data)
	}

	if len(ref.Value.Type) == 0 {
		t := toType(data)
		return p.parseType(data, ref.Value, t)
	}

	var v interface{}
	var err error
	for _, typeName := range ref.Value.Type {
		v, err = p.parseType(data, ref.Value, typeName)
		if err == nil {
			return v, nil
		}
	}

	return nil, err
}

func (p *Parser) parseType(data interface{}, schema *schema.Schema, typeName string) (interface{}, error) {
	switch data.(type) {
	case []interface{}:
		if typeName != "array" {
			return nil, fmt.Errorf("found array, expected %v: %v", schema, data)
		}
	case map[string]interface{}:
		if typeName != "object" {
			return nil, fmt.Errorf("found object, expected %v: %v", schema, data)
		}
	case struct{}:
		if typeName != "object" {
			return nil, fmt.Errorf("found object, expected %v: %v", schema, data)
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
		obj, err := p.ParseObject(data, schema)
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
