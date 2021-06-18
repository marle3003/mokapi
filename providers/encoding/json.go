package encoding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic/openapi"
)

type custom map[string]interface{}

func MarshalJSON(obj interface{}, schema *openapi.SchemaRef) ([]byte, error) {
	data, err := selectData(obj, schema)
	if err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

func (m custom) MarshalJSON() ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	buf.WriteRune('{')
	l := len(m)
	i := 0
	for k, v := range m {
		key, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteRune(':')
		value, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		buf.Write(value)
		if i != l-1 {
			buf.WriteRune(',')
		}
		i++
	}
	buf.WriteRune('}')
	return buf.Bytes(), nil
}

func selectData(data interface{}, schema *openapi.SchemaRef) (interface{}, error) {
	if schema.Value == nil || schema.Value.Type == "" {
		return data, nil
	}

	if data == nil {
		return nil, nil
	}

	switch data.(type) {
	case []interface{}:
		if schema.Value.Type != "array" {
			return nil, fmt.Errorf("expected %q but found array", schema.Value.Type)
		}
	case map[string]interface{}:
		if schema.Value.Type != "object" {
			return nil, fmt.Errorf("expected %q but found object", schema.Value.Type)
		}
	}

	if schema.Value.Type == "array" {
		if list, ok := data.([]interface{}); ok {
			for i, e := range list {
				var err error
				list[i], err = selectData(e, schema.Value.Items)
				if err != nil {
					return nil, err
				}
			}
			return list, nil
		}

		return nil, fmt.Errorf("unexpected type for schema type array")
	} else if schema.Value.Type == "object" {
		var obj map[string]interface{}
		if o, isObject := data.(map[string]interface{}); isObject {
			obj = o
		} else {
			s := fmt.Sprintf("%v", data)
			err := json.Unmarshal([]byte(s), &obj)
			if err != nil {
				return nil, err
			}
		}

		if schema.Value.Properties == nil || len(schema.Value.Properties.Value) == 0 {
			return obj, nil
		}

		selectedData := make(custom)

		for k, v := range obj {
			if p, ok := schema.Value.Properties.Value[k]; ok {
				var err error
				selectedData[k], err = selectData(v, p)
				if err != nil {
					return nil, err
				}
			} else if schema.Value.AdditionalProperties != nil {
				var err error
				selectedData[k], err = selectData(v, schema.Value.AdditionalProperties)
				if err != nil {
					return nil, err
				}
			}
		}
		return selectedData, nil
	}
	return data, nil
}
