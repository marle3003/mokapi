package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mokapi/models/media"
	"mokapi/sortedmap"
	"reflect"
	"unicode"
)

func (s *Ref) Marshal(i interface{}, contentType *media.ContentType) ([]byte, error) {
	switch contentType.Subtype {
	case "json":
		o, err := selectData(i, s)
		if err != nil {
			return nil, err
		}
		b, err := json.Marshal(o)
		if err, ok := err.(*json.SyntaxError); ok {
			return nil, fmt.Errorf("json error (%v): %v", err.Offset, err.Error())
		}
		return b, err
	case "xml", "rss+xml":
		//var buffer bytes.Buffer
		//w := newXmlWriter(&buffer)
		//err := w.write(i, schema)
		//if err != nil {
		//	return nil, err
		//}
		//return buffer.Bytes(), nil
	default:
		if s, ok := i.(string); ok {
			return []byte(s), nil
		}
		return nil, fmt.Errorf("unspupported encoding for content type %v", contentType)
	}

	return nil, fmt.Errorf("unsupported content type %v", contentType)
}

type schemaObject struct {
	sortedmap.LinkedHashMap
}

// selectData selects data by schema
func selectData(data interface{}, schema *Ref) (interface{}, error) {
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
			obj := schemaObject{}
			for i := 0; i < v.NumField(); i++ {
				ft := t.Field(i)
				name := unTitle(ft.Name)
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
			return obj, nil
		} else if v.Kind() == reflect.Map {
			// a map ensures not the order of entries
			// => loop over schema.fields to ensure an order
			// we get a map for example from lua scripts
			obj := schemaObject{}
			m := data.(map[interface{}]interface{})
			if schema.Value != nil || schema.Value.Properties.Value.Len() > 0 {
				for it := schema.Value.Properties.Value.Iter(); it.Next(); {
					name := it.Key().(string)
					p := it.Value().(*Ref)
					if v, ok := m[name]; ok {
						d, err := selectData(v, p)
						if err != nil {
							return nil, err
						}
						obj.Set(name, d)
					}
				}
			} else {
				for k, v := range m {
					obj.Set(k, v)
				}
			}

			return obj, nil
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

func (m schemaObject) MarshalJSON() ([]byte, error) {
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

// UnTitle returns a copy of the string s with first letter mapped to its Unicode lower case
func unTitle(s string) string {
	for i, v := range s {
		return string(unicode.ToLower(v)) + s[i+1:]
	}
	return s
}
