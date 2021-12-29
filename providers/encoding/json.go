package encoding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic/openapi"
	"mokapi/sortedmap"
	"reflect"
	"unicode"
)

type custom struct {
	sortedmap.LinkedHashMap
}

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
	l := m.Len()
	i := 0
	for it := m.Iter(); it.Next(); {
		k := it.Key()
		v := it.Value()

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
	case struct{}:
		if schema.Value.Type != "object" {
			return nil, fmt.Errorf("expected %q but found object", schema.Value.Type)
		}
	}

	if schema.Value.Type == "array" {
		if list, ok := data.([]interface{}); ok {
			result := make([]interface{}, len(list))
			for i, e := range list {
				var err error
				result[i], err = selectData(e, schema.Value.Items)
				if err != nil {
					return nil, err
				}
			}
			return result, nil
		}
		if m, ok := data.(map[interface{}]interface{}); ok {
			result := make([]interface{}, 0)
			for _, v := range m {
				if i, err := selectData(v, schema.Value.Items); err != nil {
					return nil, err
				} else {
					result = append(result, i)
				}
			}
			return result, nil
		}

		return nil, fmt.Errorf("unexpected type for schema type array")
	} else if schema.Value.Type == "object" {
		v := reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			t := v.Type()
			obj := custom{}
			for i := 0; i < v.NumField(); i++ {
				ft := t.Field(i)
				name := lowerTitle(ft.Name)
				val := v.Field(i)

				if schema.Value != nil || schema.Value.Properties.Value.Len() > 0 {
					if p := schema.Value.Properties.Get(name); p != nil {
						d, err := selectData(val.Interface(), p)
						if err != nil {
							return nil, err
						}
						obj.Set(name, d)
					} else if schema.Value.AdditionalProperties != nil {
						d, err := selectData(val.Interface(), schema.Value.AdditionalProperties)
						if err != nil {
							return nil, err
						}
						obj.Set(name, d)
					}
				} else {
					obj.Set(name, val.Interface())
				}
			}
		} else {
			panic("to do")
			//s := fmt.Sprintf("%v", data)
			//err := json.Unmarshal([]byte(s), &obj)
			//if err != nil {
			//	return nil, fmt.Errorf("unable to map %T with value \"%v\" to object", data, data)
			//}
		}
	}
	return data, nil
}

func lowerTitle(s string) string {
	for i, v := range s {
		return string(unicode.ToLower(v)) + s[i+1:]
	}
	return s
}
