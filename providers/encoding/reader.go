package encoding

import (
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic/openapi"
	"mokapi/models/media"
)

type Reader interface {
	ReadObject(interface{}, *openapi.SchemaRef) (map[string]interface{}, error)
	ReadArray(interface{}, *openapi.SchemaRef) ([]interface{}, error)
	ReadInteger(interface{}, *openapi.SchemaRef) (int64, error)
	ReadNumber(interface{}, *openapi.SchemaRef) (float64, error)
	ReadString(interface{}, *openapi.SchemaRef) (string, error)
	ReadBoolean(interface{}, *openapi.SchemaRef) (bool, error)
}

func Parse(data []byte, contentType *media.ContentType, schema *openapi.SchemaRef) (v interface{}, err error) {
	var r Reader
	switch contentType.Subtype {
	case "json":
		err = json.Unmarshal(data, &v)
		r = &JsonReader{}
	}

	if err != nil {
		return nil, err
	}

	return parse(v, schema, r)
}

func parse(v interface{}, schema *openapi.SchemaRef, r Reader) (interface{}, error) {
	if schema == nil || schema.Value == nil {
		return v, nil
	}

	if len(schema.Value.AnyOf) > 0 {
		return parseAnyOf(v, schema.Value.AnyOf, r)
	} else if len(schema.Value.AllOf) > 0 {
		return parseAllOf(v, schema.Value.AllOf, r)
	} else if len(schema.Value.OneOf) > 0 {
		return parseOneOf(v, schema.Value.OneOf, r)
	} else if len(schema.Value.Type) == 0 {
		// A schema without a type matches any data type
		return v, nil
	}

	switch schema.Value.Type {
	case "object":
		return r.ReadObject(v, schema)
	case "array":
		return r.ReadArray(v, schema)
	case "boolean":
		return r.ReadBoolean(v, schema)
	case "integer":
		i, err := r.ReadInteger(v, schema)
		if err == nil {
			return i, validateInt64(i, schema)
		}
		return nil, err
	case "number":
		f, err := r.ReadNumber(v, schema)
		if err == nil {
			return f, validateFloat64(f, schema)
		}
		return nil, err
	case "string":
		s, err := r.ReadString(v, schema)
		if err == nil {
			return s, validateString(s, schema)
		}
		return nil, err
	}

	return nil, fmt.Errorf("unsupported type %q", schema.Value.Type)
}

func parseAnyOf(v interface{}, schemas openapi.SchemaRefs, r Reader) (interface{}, error) {
	result := make(map[string]interface{})

	for _, s := range schemas {
		o, err := parse(v, s, r)

		if s.Value.Type == "object" && err == nil {
			m := o.(map[string]interface{})
			for k, v := range m {
				result[k] = v
			}
		} else if err == nil {
			return o, nil
		}
	}
	return result, nil
}

func parseAllOf(v interface{}, schemas openapi.SchemaRefs, r Reader) (interface{}, error) {
	result := make(map[string]interface{})

	for _, s := range schemas {
		o, err := parse(v, s, r)
		if err != nil {
			// To be valid against allOf, the data provided by the client must be valid against all
			// of the given subschemas
			return nil, err
		}

		m, ok := o.(map[string]interface{})

		if !ok {
			return nil, fmt.Errorf("expected schema type of object inside allOf")
		}

		for k, v := range m {
			result[k] = v
		}
	}
	return result, nil
}

func parseOneOf(v interface{}, schemas openapi.SchemaRefs, r Reader) (interface{}, error) {
	var result interface{}

	for _, s := range schemas {
		o, err := parse(v, s, r)
		if err != nil {
			continue
		}

		if m, ok := o.(map[string]interface{}); ok && len(m) == 0 {
			continue
		}

		if result != nil {
			return nil, fmt.Errorf("oneOf: given data is valid against more as one schema")
		}

		result = o
	}
	return result, nil
}
