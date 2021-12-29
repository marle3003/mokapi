package encoding

import (
	"encoding/json"
	"fmt"
	"io"
	"mokapi/config/dynamic/openapi"
	"mokapi/models/media"
	"reflect"
	"strings"
)

func ParseFrom(r io.Reader, contentType *media.ContentType, schema *openapi.SchemaRef) (v interface{}, err error) {
	switch contentType.Subtype {
	case "json":
		err = json.NewDecoder(r).Decode(&v)
		if err != nil {
			return nil, fmt.Errorf("invalid json format: %v", err)
		}
	}

	return parse(v, schema)
}

func Parse(data []byte, contentType *media.ContentType, schema *openapi.SchemaRef) (v interface{}, err error) {
	switch contentType.Subtype {
	case "json":
		err = json.Unmarshal(data, &v)
		if err != nil {
			return nil, fmt.Errorf("invalid json format: %v", err)
		}
	}

	return parse(v, schema)
}

func parse(v interface{}, schema *openapi.SchemaRef) (interface{}, error) {
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
		return readObject(v, schema.Value)
	case "array":
		return readArray(v, schema.Value)
	case "boolean":
		return readBoolean(v, schema)
	case "integer":
		return readInteger(v, schema.Value)
	case "number":
		return readNumber(v, schema.Value)
	case "string":
		s, err := readString(v, schema)
		if err == nil {
			return s, validateString(s, schema.Value)
		}
		return nil, err
	}

	return nil, fmt.Errorf("unsupported type %q", schema.Value.Type)
}

func parseAnyOf(i interface{}, schemas openapi.SchemaRefs) (interface{}, error) {
	if m, ok := i.(map[string]interface{}); ok {
		return parseAnyObject(m, schemas)
	}
	return parseAnyValue(i, schemas)
}

func parseAnyObject(m map[string]interface{}, schemas openapi.SchemaRefs) (interface{}, error) {
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
			pRef := it.Value().(*openapi.SchemaRef)
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

func parseAnyValue(i interface{}, schemas openapi.SchemaRefs) (interface{}, error) {
	for _, s := range schemas {
		i, err := parse(i, s)
		if err == nil {
			return i, nil
		}
	}
	return nil, fmt.Errorf("value %v does not match any of expected schema", i)
}

func parseAllOf(i interface{}, schemas openapi.SchemaRefs) (interface{}, error) {
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
			pRef := it.Value().(*openapi.SchemaRef)
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

func parseOneOf(i interface{}, schemas openapi.SchemaRefs) (interface{}, error) {
	if m, ok := i.(map[string]interface{}); ok {
		return parseOneOfObject(m, schemas)
	}
	return parseOneOfValue(i, schemas)
}

func parseOneOfValue(i interface{}, schemas openapi.SchemaRefs) (interface{}, error) {
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

func parseOneOfObject(m map[string]interface{}, schemas openapi.SchemaRefs) (interface{}, error) {
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
			pRef := it.Value().(*openapi.SchemaRef)
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

func getType(s *openapi.Schema) reflect.Type {
	switch s.Type {
	case "integer":
		if s.Format == "int32" {
			return reflect.TypeOf(int32(0))
		}
		return reflect.TypeOf(int64(0))
	case "number":
		if s.Format == "float32" {
			return reflect.TypeOf(float32(0))
		}
		return reflect.TypeOf(float64(0))
	case "string":
		return reflect.TypeOf("")
	case "boolean":
		return reflect.TypeOf(false)
	case "array":
		return reflect.SliceOf(getType(s.Items.Value))
	}

	panic(fmt.Sprintf("type %v not implemented", s.Type))
}
