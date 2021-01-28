package encoding

import (
	"bytes"
	"encoding/json"
	"mokapi/models"
)

type custom map[string]interface{}

func MarshalJSON(obj interface{}, schema *models.Schema) ([]byte, error) {
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

func selectData(data interface{}, schema *models.Schema) interface{} {
	if schema == nil || schema.Type == "" {
		if list, ok := data.([]interface{}); ok {
			for i, e := range list {
				list[i] = selectData(e, nil)
			}
			return list
		}
		if o, ok := data.(map[string]interface{}); ok {
			selectedData := make(custom)
			for k, v := range o {
				selectedData[k] = selectData(v, nil)
			}
			return selectedData
		}
		return data
	}

	if schema.Type == "array" {
		if list, ok := data.([]interface{}); ok {
			for i, e := range list {
				list[i] = selectData(e, schema.Items)
			}
			return list
		}
		// todo error handling
		return nil
	} else if schema.Type == "object" {
		if o, isObject := data.(map[string]interface{}); !isObject {
			// todo error handling
			return nil
		} else {
			selectedData := make(custom)

			for k, v := range o {
				if p, ok := schema.Properties[k]; ok {
					selectedData[k] = selectData(v, p)
				} else if schema.AdditionalProperties != nil {
					selectedData[k] = selectData(v, schema.AdditionalProperties)
				}
			}
			return selectedData
		}
	}
	return data
}
