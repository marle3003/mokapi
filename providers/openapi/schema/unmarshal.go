package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mokapi/media"
	"mokapi/sortedmap"
	"reflect"
	"strconv"
	"strings"
)

const unmarshalErrorFormat = "unmarshal data failed: %w"

func ParseString(s string, schema *Ref) (interface{}, error) {
	return parse(s, schema)
}

func (r *Ref) Unmarshal(b []byte, contentType media.ContentType) (i interface{}, err error) {
	switch {
	case contentType.Subtype == "json":
		err = json.Unmarshal(b, &i)
		if err != nil {
			err = fmt.Errorf("invalid json format: %v", err)
		}
	case contentType.IsXml():
		i, err = unmarshalXml(b, r)
	default:
		i = string(b)
	}

	if err != nil {
		return nil, fmt.Errorf(unmarshalErrorFormat, err)
	}

	i, err = parse(i, r)
	if err != nil {
		err = fmt.Errorf(unmarshalErrorFormat, err)
	}
	return
}

func UnmarshalFrom(r io.Reader, contentType media.ContentType, schema *Ref) (i interface{}, err error) {
	switch contentType.Subtype {
	case "json":
		err = json.NewDecoder(r).Decode(&i)
		if err != nil {
			return nil, fmt.Errorf("invalid json format: %v", err)
		}
	default:
		i, err = io.ReadAll(r)
	}

	if err != nil {
		return
	}

	return parse(i, schema)
}

func parse(v interface{}, schema *Ref) (interface{}, error) {
	if schema == nil || schema.Value == nil {
		return v, nil
	}

	if len(schema.Value.AnyOf) > 0 {
		return parseAny(v, schema.Value)
	}
	if len(schema.Value.AllOf) > 0 {
		return parseAllOf(v, schema.Value)
	}
	if len(schema.Value.OneOf) > 0 {
		return parseOneOf(v, schema.Value)
	}
	if len(schema.Value.Type) == 0 {
		// A schema without a type matches any data type
		if _, ok := v.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
			return parseObject(v, &Schema{Type: "object"})
		}
		return v, nil
	}

	switch schema.Value.Type {
	case "object":
		return parseObject(v, schema.Value)
	case "array":
		return parseArray(v, schema.Value)
	case "boolean":
		return readBoolean(v, schema.Value)
	case "integer":
		return parseInteger(v, schema.Value)
	case "number":
		return parseNumber(v, schema.Value)
	case "string":
		return parseString(v, schema.Value)
	}

	return nil, fmt.Errorf("unsupported type %q", schema.Value.Type)
}

func parseAny(i interface{}, schema *Schema) (interface{}, error) {
	var result interface{}

	for _, ref := range schema.AnyOf {
		// free-form object
		if ref.Value.Type == "object" && ref.Value.Properties == nil {
			return i, nil
		}

		part, err := parse(i, ref)
		if err != nil {
			continue
		}

		if values, ok := part.(map[string]interface{}); ok {
			m, ok := result.(map[string]interface{})
			if !ok {
				if result != nil {
					return result, nil
				} else {
					m = map[string]interface{}{}
				}
			}

			for k, v := range values {
				// only overwrite value if prop is in schema
				prop := ref.Value.Properties.Get(k)
				if _, found := m[k]; found && prop != nil {
					m[k] = v
				} else if !found {
					m[k] = v
				}
			}
			result = m
		} else {
			result = part
		}
	}

	if result == nil {
		return nil, fmt.Errorf("parse %v failed, expected %v", i, schema)
	}

	return result, nil
}

func parseAllOf(i interface{}, schema *Schema) (interface{}, error) {
	m := i.(map[string]interface{})

	result := map[string]interface{}{}
	for _, sRef := range schema.AllOf {
		s := sRef.Value

		// free-form object
		if s.Properties == nil {
			return m, nil
		}

		part, err := parse(m, sRef)
		if err != nil {
			return nil, fmt.Errorf("parse %v failed: value does not match part of allOf: %w", toString(i), err)
		}

		values, ok := part.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("ERROR")
		}
		for k, v := range values {
			// only overwrite value if prop is in schema
			prop := s.Properties.Get(k)
			if _, found := result[k]; found && prop != nil {
				result[k] = v
			} else if !found {
				result[k] = v
			}
		}
	}

	return result, nil
}

func parseOneOf(i interface{}, schema *Schema) (interface{}, error) {
	return parseOneOfObject(i, schema)
}

func parseOneOfObject(i interface{}, schema *Schema) (interface{}, error) {
	var result interface{}

	for _, ref := range schema.OneOf {
		// free-form object
		if ref.Value.Type == "object" && ref.Value.Properties == nil {
			result = i
		}

		part, err := parse(i, ref)
		if err != nil {
			continue
		}

		if result != nil {
			return nil, fmt.Errorf("parse %v failed: it is valid for more than one schema, expected %v", toString(i), schema)
		}

		result = part
	}

	if result == nil {
		return nil, fmt.Errorf("parse %v failed: expected to match one of schema but it matches none", toString(i))
	}

	return result, nil
}

func parseObject(i interface{}, s *Schema) (interface{}, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("parse %v as object", toString(i))
	}

	var err error

	if err = validateObject(m, s); err != nil {
		return nil, err
	}

	if s.IsDictionary() {
		result := make(map[string]interface{})
		for k, v := range m {
			v, err = parse(v, s.AdditionalProperties.Ref)
			if err != nil {
				return nil, err
			}
			result[k] = v
		}
		return result, nil
	}

	if !s.HasProperties() {
		return m, nil
	}

	result := map[string]interface{}{}
	for k, v := range m {
		prop := s.Properties.Get(k)

		v, err := parse(v, prop)
		if err != nil {
			return nil, fmt.Errorf("parse '%v' failed: %w", k, err)
		}
		result[k] = v
	}

	return result, nil
}

func parseArray(i interface{}, s *Schema) (interface{}, error) {
	arr, ok := i.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected array but got %v", toString(i))
	}

	result := make([]interface{}, 0)
	for _, o := range arr {
		v, err := parse(o, s.Items)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}

	return result, validateArray(result, s)
}

func parseInteger(i interface{}, s *Schema) (n int64, err error) {
	switch v := i.(type) {
	case int:
		n = int64(v)
	case int64:
		n = v
	case float64:
		if math.Trunc(v) != v {
			return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
		}
		n = int64(v)
	case int32:
		n = int64(v)
	case string:
		switch s.Format {
		case "int64":
			n, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
			}
			return n, nil
		default:
			temp, err := strconv.Atoi(v)
			if err != nil {
				return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
			}
			n = int64(temp)
		}
	default:
		return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
	}

	switch s.Format {
	case "int32":
		if n > math.MaxInt32 || n < math.MinInt32 {
			return 0, fmt.Errorf("parse '%v' failed: represents a number either less than int32 min value or greater max value, expected %v", i, s)
		}
	}

	return n, validateInt64(n, s)
}

func parseNumber(i interface{}, s *Schema) (f float64, err error) {
	switch v := i.(type) {
	case float64:
		f = v
	case string:
		f, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
		}
	case int:
		f = float64(v)
	case int64:
		f = float64(v)
	default:
		return 0, fmt.Errorf("parse '%v' failed, expected %v", v, s)
	}

	switch s.Format {
	case "float":
		if f > math.MaxFloat32 {
			return 0, fmt.Errorf("parse %v failed, expected %v", i, s)
		}
	}

	return f, validateFloat64(f, s)
}

func parseString(v interface{}, schema *Schema) (interface{}, error) {
	s, ok := v.(string)
	if !ok {
		if schema.Nullable {
			return nil, nil
		}
		return nil, fmt.Errorf("parse %v failed, expected %v", v, schema)
	}

	return s, validateString(s, schema)
}

func readBoolean(i interface{}, s *Schema) (bool, error) {
	if b, ok := i.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("parse %v failed, expected %v", i, s)
}

func toString(i interface{}) string {
	var sb strings.Builder
	switch o := i.(type) {
	case []interface{}:
		sb.WriteRune('[')
		for i, v := range o {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(toString(v))
		}
		sb.WriteRune(']')
	case map[string]interface{}:
		sb.WriteRune('{')
		for key, val := range o {
			if sb.Len() > 1 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%v: %v", key, toString(val)))
		}
		sb.WriteRune('}')
	case string, int, int32, int64, float32, float64:
		sb.WriteString(fmt.Sprintf("%v", o))
	case *sortedmap.LinkedHashMap[string, interface{}]:
		return o.String()
	default:
		v := reflect.ValueOf(i)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		t := reflect.TypeOf(i)
		switch v.Kind() {
		case reflect.Slice:
			sb.WriteRune('[')
			for i := 0; i < v.Len(); i++ {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(toString(v.Index(i).Interface()))
			}
			sb.WriteRune(']')
		case reflect.Struct:
			sb.WriteRune('{')
			for i := 0; i < v.NumField(); i++ {
				if i > 0 {
					sb.WriteString(", ")
				}
				name := t.Field(i).Name
				fv := v.Field(i).Interface()
				sb.WriteString(fmt.Sprintf("%v: %v", name, fv))
			}
			sb.WriteRune('}')
		}
	}
	return sb.String()
}

func getType(s *Schema) (reflect.Type, error) {
	switch s.Type {
	case "integer":
		if s.Format == "int32" {
			return reflect.TypeOf(int32(0)), nil
		}
		return reflect.TypeOf(int64(0)), nil
	case "number":
		if s.Format == "float32" {
			return reflect.TypeOf(float32(0)), nil
		}
		return reflect.TypeOf(float64(0)), nil
	case "string":
		return reflect.TypeOf(""), nil
	case "boolean":
		return reflect.TypeOf(false), nil
	case "array":
		t, err := getType(s.Items.Value)
		if err != nil {
			return nil, err
		}
		return reflect.SliceOf(t), nil
	}

	return nil, fmt.Errorf("type %v not implemented", s.Type)
}
