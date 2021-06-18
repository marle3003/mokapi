package encoding

import (
	"fmt"
	"mokapi/config/dynamic/openapi"
)

type JsonReader struct {
	data interface{}
}

func (r *JsonReader) ReadObject(v interface{}, schema *openapi.SchemaRef) (map[string]interface{}, error) {
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected object but got %t", v)
	}
	result := make(map[string]interface{})
	for k, v := range m {
		if schema.Value == nil || schema.Value.Properties == nil || schema.Value.Properties.Value == nil {
			continue
		}
		if p, ok := schema.Value.Properties.Value[k]; ok {
			var err error
			result[k], err = parse(v, p, r)
			if err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}
func (r *JsonReader) ReadArray(v interface{}, schema *openapi.SchemaRef) ([]interface{}, error) {
	a, ok := v.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected array but got %t", v)
	}
	result := make([]interface{}, 0)
	for _, v := range a {
		v, err := parse(v, schema.Value.Items, r)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}
func (r *JsonReader) ReadInteger(v interface{}, _ *openapi.SchemaRef) (int64, error) {
	f, ok := v.(float64)
	if !ok {
		return 0, fmt.Errorf("expected integer got %t", v)
	}
	return int64(f), nil
}
func (r *JsonReader) ReadNumber(v interface{}, schema *openapi.SchemaRef) (float64, error) {
	f, ok := v.(float64)
	if !ok {
		return 0, fmt.Errorf("expected float got %t", v)
	}
	return f, nil
}
func (r *JsonReader) ReadString(v interface{}, schema *openapi.SchemaRef) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("expected string got %t", v)
	}
	return s, nil
}
func (r *JsonReader) ReadBoolean(v interface{}, _ *openapi.SchemaRef) (bool, error) {
	switch t := v.(type) {
	case bool:
		return t, nil
	case string:
		return t == "true" || t == "1", nil
	}
	return false, nil
}
