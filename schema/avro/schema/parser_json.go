package schema

import "math"

func (p *Parser) parseFromInterface(v interface{}) (interface{}, error) {
	return parseFromInterface(v, p.Schema)
}

func parseFromInterface(v interface{}, s *Schema) (interface{}, error) {
	var err error
	var result interface{}
	for _, t := range s.Type {
		switch t := t.(type) {
		case string:
			result, err = parseType(v, s, t)
			if err == nil {
				return result, nil
			}
		case *Schema:
			result, err = parseFromInterface(v, t)
			if err == nil {
				return result, nil
			}
		}
	}

	return nil, err
}

func parseType(v interface{}, s *Schema, typeName string) (interface{}, error) {
	switch typeName {
	case "null":
		if v == nil {
			return nil, nil
		}
		return nil, Errorf("type", "invalid type, expected null but got %v", ToType(v))
	case "string":
		if str, ok := v.(string); ok {
			return str, nil
		}
	case "boolean":
		if b, ok := v.(bool); ok {
			return b, nil
		}
	case "int":
		switch val := v.(type) {
		case int, int64:
			return val, nil
		case float64:
			if math.Trunc(val) != val {
				return 0, Errorf("type", "invalid type, expected %v but got %v", typeName, ToType(v))
			}
			return int(val), nil
		}
	case "long":
		switch val := v.(type) {
		case int:
			return val, nil
		case float64:
			if math.Trunc(val) != val {
				return 0, Errorf("type", "invalid type, expected %v but got %v", typeName, ToType(v))
			}
			return int64(val), nil
		}
	case "float":
		if f, ok := v.(float64); ok {
			return float32(f), nil
		}
		if f, ok := v.(float32); ok {
			return f, nil
		}
	case "double":
		if f, ok := v.(float64); ok {
			return f, nil
		}
	case "enum":
		if str, ok := v.(string); ok {
			for _, val := range s.Symbols {
				if val == str {
					return val, nil
				}
			}
			return nil, Errorf("enum", "value '%v' does not match one in the symbols %v", str, ToString(s.Symbols))
		}
	case "record":
		if m, ok := v.(map[string]interface{}); ok {
			return parseRecord(m, s)
		}
	case "array":
		var result []interface{}
		if a, ok := v.([]interface{}); ok {
			for _, val := range a {
				item, err := parseFromInterface(val, s.Items)
				if err != nil {
					return nil, wrapError("items", err)
				}
				result = append(result, item)
			}
		}
		return result, nil
	case "map":
		if m, ok := v.(map[string]interface{}); ok {
			for key, val := range m {
				_, err := parseFromInterface(val, s.Values)
				if err != nil {
					return nil, Errorf("map", "value of '%v' does not match schema: %v", key, err)
				}
			}
			return nil, nil
		}
	case "fixed":
		switch val := v.(type) {
		case []byte:
			if len(val) != s.Size {
				return nil, Errorf("fixed", "invalid fixed size, expected %v but got %v", s.Size, len(val))
			}
		case string:
			if len(val) != s.Size {
				return nil, Errorf("fixed", "invalid fixed size, expected %v but got %v", s.Size, len(val))
			}
		}
		return nil, nil
	}

	return nil, Errorf("type", "invalid type, expected %v but got %v", typeName, ToType(v))
}

func parseRecord(m map[string]interface{}, s *Schema) (interface{}, error) {
	result := make(map[string]interface{})
	for _, field := range s.Fields {
		if v, ok := m[field.Name]; ok {
			vf, err := parseFromInterface(v, &field)
			if err != nil {
				return nil, wrapError(field.Name, err)
			}
			result[field.Name] = vf
		}
	}

	return result, nil
}
