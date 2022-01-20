package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mokapi/models/media"
	"reflect"
	"strconv"
	"strings"
)

func ParseString(s string, schema *Ref) (interface{}, error) {
	return parse(s, schema)
}

func Parse(b []byte, contentType *media.ContentType, schema *Ref) (i interface{}, err error) {
	switch {
	case contentType.Subtype == "json":
		err = json.Unmarshal(b, &i)
		if err != nil {
			return nil, fmt.Errorf("invalid json format: %v", err)
		}
	default:
		i = string(b)
	}

	if err != nil {
		return
	}

	return parse(i, schema)
}

func ParseFrom(r io.Reader, contentType *media.ContentType, schema *Ref) (i interface{}, err error) {
	switch contentType.Subtype {
	case "json":
		err = json.NewDecoder(r).Decode(&i)
		if err != nil {
			return nil, fmt.Errorf("invalid json format: %v", err)
		}
	default:
		i, err = ioutil.ReadAll(r)
	}

	if err != nil {
		return
	}

	return parse(i, schema)
}

func parse(v interface{}, schema *Ref) (interface{}, error) {
	if schema == nil || schema.Value == nil {
		if m, ok := v.(map[string]interface{}); ok {
			return toObject(m), nil
		}
		return v, nil
	}

	if len(schema.Value.AnyOf) > 0 {
		return parseAnyOf(v, schema.Value.AnyOf)
	} else if len(schema.Value.AllOf) > 0 {
		return parseAllOf(v, schema.Value.AllOf)
	} else if len(schema.Value.OneOf) > 0 {
		return parseOneOf(v, schema.Value.OneOf)
	} else if len(schema.Value.Type) == 0 {
		// A schema without a type matches any data type
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

func parseAnyOf(i interface{}, schemas []*Ref) (interface{}, error) {
	if m, ok := i.(map[string]interface{}); ok {
		return parseAnyObject(m, schemas)
	}
	return parseAnyValue(i, schemas)
}

func parseAnyObject(m map[string]interface{}, schemas []*Ref) (interface{}, error) {
	fields := make([]reflect.StructField, 0, len(m))
	values := make([]reflect.Value, 0, len(m))

	for _, ref := range schemas {
		s := ref.Value

		required := make(map[string]struct{})
		for _, r := range s.Required {
			required[r] = struct{}{}
		}

		// free-form object
		if s.Properties == nil {
			return toObject(m), nil
		}

		for it := s.Properties.Value.Iter(); it.Next(); {
			name := it.Key().(string)
			pRef := it.Value().(*Ref)
			p := pRef.Value

			if _, ok := m[name]; !ok {
				if _, ok := required[name]; ok && len(required) > 0 {
					return nil, fmt.Errorf("expected required property %v", name)
				}
				continue
			}

			v, err := parse(m[name], pRef)
			if err != nil {
				continue
			}
			values = append(values, reflect.ValueOf(v))
			fields = append(fields, reflect.StructField{
				Name: strings.Title(name),
				Type: getType(p),
				Tag:  reflect.StructTag(fmt.Sprintf(`json:"%v"`, name)),
			})
		}
	}

	if len(m) > len(fields) {
		return nil, fmt.Errorf("too many properties for object")
	}

	t := reflect.StructOf(fields)
	v := reflect.New(t).Elem()
	for i, val := range values {
		v.Field(i).Set(val)
	}
	return v.Addr().Interface(), nil
}

func parseAnyValue(i interface{}, schemas []*Ref) (interface{}, error) {
	for _, s := range schemas {
		i, err := parse(i, s)
		if err == nil {
			return i, nil
		}
	}
	return nil, fmt.Errorf("value %v does not match any of expected schema", i)
}

func parseAllOf(i interface{}, schemas []*Ref) (interface{}, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected an object for allOf")
	}
	fields := make([]reflect.StructField, 0, len(m))
	values := make([]reflect.Value, 0, len(m))

	for _, sRef := range schemas {
		s := sRef.Value

		required := make(map[string]struct{})
		for _, r := range s.Required {
			required[r] = struct{}{}
		}

		// free-form object
		if s.Properties == nil {
			return toObject(m), nil
		}

		for it := s.Properties.Value.Iter(); it.Next(); {
			name := it.Key().(string)
			pRef := it.Value().(*Ref)
			p := pRef.Value

			if _, ok := m[name]; !ok {
				if _, ok := required[name]; ok && len(required) > 0 {
					return nil, fmt.Errorf("expected required property %v", name)
				}
				continue
			}

			v, err := parse(m[name], pRef)
			if err != nil {
				return nil, fmt.Errorf("value does not match all schema")
			}
			values = append(values, reflect.ValueOf(v))
			fields = append(fields, reflect.StructField{
				Name: strings.Title(name),
				Type: getType(p),
				Tag:  reflect.StructTag(fmt.Sprintf(`json:"%v"`, name)),
			})
		}
	}

	t := reflect.StructOf(fields)
	v := reflect.New(t).Elem()
	for i, val := range values {
		v.Field(i).Set(val)
	}
	return v.Addr().Interface(), nil
}

func parseOneOf(i interface{}, schemas []*Ref) (interface{}, error) {
	if m, ok := i.(map[string]interface{}); ok {
		return parseOneOfObject(m, schemas)
	}
	return parseOneOfValue(i, schemas)
}

func parseOneOfObject(m map[string]interface{}, schemas []*Ref) (interface{}, error) {
	var result interface{}
	for _, sRef := range schemas {
		s := sRef.Value
		fields := make([]reflect.StructField, 0, len(m))
		values := make([]reflect.Value, 0, len(m))

		required := make(map[string]struct{})
		for _, r := range s.Required {
			required[r] = struct{}{}
		}

		for it := s.Properties.Value.Iter(); it.Next(); {
			name := it.Key().(string)
			pRef := it.Value().(*Ref)
			p := pRef.Value

			if _, ok := m[name]; !ok {
				if _, ok := required[name]; ok && len(required) > 0 {
					return nil, fmt.Errorf("expected required property %v", name)
				}
				continue
			}

			v, err := parse(m[name], pRef)
			if err != nil {
				continue
			}
			values = append(values, reflect.ValueOf(v))
			fields = append(fields, reflect.StructField{
				Name: strings.Title(name),
				Type: getType(p),
				Tag:  reflect.StructTag(fmt.Sprintf(`json:"%v"`, name)),
			})
		}

		if len(m) > len(fields) {
			continue
		}

		if result != nil {
			return nil, fmt.Errorf("oneOf: given data is valid against more as one schema")
		}

		t := reflect.StructOf(fields)
		v := reflect.New(t).Elem()
		for i, val := range values {
			v.Field(i).Set(val)
		}
		result = v.Addr().Interface()
	}

	if result == nil {
		return nil, fmt.Errorf("value does not match any of oneof schema")
	}

	return result, nil
}

func parseOneOfValue(i interface{}, schemas []*Ref) (interface{}, error) {
	var result interface{}
	for _, s := range schemas {
		v, err := parse(i, s)
		if err != nil {
			continue
		}
		if result != nil {
			return nil, fmt.Errorf("oneOf: given data is valid against more as one schema")
		}
		result = v
	}

	return result, nil
}

func parseObject(i interface{}, s *Schema) (interface{}, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected object but got %T", i)
	}

	required := make(map[string]struct{})
	for _, r := range s.Required {
		required[r] = struct{}{}
	}

	// free-form object
	if s.Properties == nil {
		return toObject(m), nil
	}

	if len(m) > s.Properties.Value.Len() {
		return nil, fmt.Errorf("too many properties for object")
	}

	fields := make([]reflect.StructField, 0, len(m))
	values := make([]reflect.Value, 0, len(m))

	for it := s.Properties.Value.Iter(); it.Next(); {
		name := it.Key().(string)
		pRef := it.Value().(*Ref)
		if _, ok := m[name]; !ok {
			if _, ok := required[name]; ok && len(required) > 0 {
				return nil, fmt.Errorf("expected required property %v", name)
			}
			continue
		}

		v, err := parse(m[name], pRef)
		if err != nil {
			return nil, err
		}
		values = append(values, reflect.ValueOf(v))
		fields = append(fields, reflect.StructField{
			Name: strings.Title(name),
			Type: reflect.TypeOf(v),
			Tag:  reflect.StructTag(fmt.Sprintf(`json:"%v"`, name)),
		})
	}

	t := reflect.StructOf(fields)
	v := reflect.New(t).Elem()
	for i, val := range values {
		v.Field(i).Set(val)
	}
	return v.Addr().Interface(), nil
}

func parseArray(i interface{}, s *Schema) (interface{}, error) {
	a, ok := i.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected array but got %T", i)
	}

	if len(a) == 0 {
		var sliceOf reflect.Type
		switch s.Items.Value.Type {
		case "object":
			sliceOf = reflect.TypeOf([]interface{}{})
		default:
			sliceOf = reflect.SliceOf(getType(s.Items.Value))
		}
		return reflect.MakeSlice(sliceOf, 0, 0).Interface(), nil
	}

	v, err := parse(a[0], s.Items)
	if err != nil {
		return nil, err
	}
	result := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(v)), 0, len(a))

	for _, v := range a {
		v, err := parse(v, s.Items)
		if err != nil {
			return nil, err
		}
		result = reflect.Append(result, reflect.ValueOf(v))
	}
	return result.Interface(), nil
}

func parseInteger(i interface{}, s *Schema) (n int64, err error) {
	switch v := i.(type) {
	case int64:
		n = v
	case float64:
		if math.Trunc(v) != v {
			return 0, fmt.Errorf("expected integer but got floating number")
		}
		n = int64(v)
	case string:
		switch s.Format {
		case "int64":
			n, err = strconv.ParseInt(v, 10, 64)
		default:
			temp, err := strconv.Atoi(v)
			if err != nil {
				return 0, err
			}
			n = int64(temp)
		}
	default:
		return 0, fmt.Errorf("expected integer got %T", i)
	}

	switch s.Format {
	case "int32":
		if n > math.MaxInt32 || n < math.MinInt32 {
			return 0, fmt.Errorf("integer is not int32")
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
	default:
		err = fmt.Errorf("expected float got %T", v)
	}

	if err != nil {
		return
	}

	switch s.Format {
	case "float":
		if f > math.MaxFloat32 || f < -math.MaxFloat32 {
			return 0, fmt.Errorf("number is not float")
		}
	}

	return f, validateFloat64(f, s)
}

func parseString(v interface{}, schema *Schema) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("expected string got %T", v)
	}

	return s, validateString(s, schema)
}

func readBoolean(i interface{}, _ *Schema) (bool, error) {
	if b, ok := i.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("expected bool but got %T", i)
}

func toObject(m map[string]interface{}) interface{} {
	fields := make([]reflect.StructField, 0, len(m))
	values := make([]reflect.Value, 0, len(m))

	for name, v := range m {
		if child, ok := v.(map[string]interface{}); ok {
			v = toObject(child)
		}
		fields = append(fields, reflect.StructField{
			Name: strings.Title(name),
			Type: reflect.TypeOf(v),
		})
		values = append(values, reflect.ValueOf(v))
	}

	t := reflect.StructOf(fields)
	v := reflect.New(t).Elem()
	for i, val := range values {
		v.Field(i).Set(val)
	}
	return v.Addr().Interface()
}
