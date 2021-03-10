package encoding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic/openapi"
)

type custom map[string]interface{}

func MarshalJSON(obj interface{}, schema *openapi.SchemaRef) ([]byte, error) {
	data := selectData(obj, schema)
	return json.Marshal(data)
}

func (m custom) MarshalJSON() ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	buf.WriteRune('{')
	l := len(m)
	i := 0
	for k, v := range m {
		key, error := json.Marshal(k)
		if error != nil {
			return nil, error
		}
		buf.Write(key)
		buf.WriteRune(':')
		value, error := json.Marshal(v)
		if error != nil {
			return nil, error
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

func selectData(data interface{}, schema *openapi.SchemaRef) interface{} {
	if schema.Value == nil || schema.Value.Type == "" {
		return data
	}

	if schema.Value.Type == "array" {
		if list, ok := data.([]interface{}); ok {
			for i, e := range list {
				list[i] = selectData(e, schema.Value.Items)
			}
			return list
		}
		// todo error handling
		return nil
	} else if schema.Value.Type == "object" {
		var obj map[string]interface{}
		if o, isObject := data.(map[string]interface{}); isObject {
			obj = o
		} else {
			s := fmt.Sprintf("%v", data)
			err := json.Unmarshal([]byte(s), &obj)
			if err != nil {
				// todo error handling
				return nil
			}
		}

		if len(schema.Value.Properties.Value) == 0 {
			return obj
		}

		selectedData := make(custom)

		for k, v := range obj {
			if p, ok := schema.Value.Properties.Value[k]; ok {
				selectedData[k] = selectData(v, p)
			} else if schema.Value.AdditionalProperties != nil {
				selectedData[k] = selectData(v, schema.Value.AdditionalProperties)
			}
		}
		return selectedData
	}
	return data
}
