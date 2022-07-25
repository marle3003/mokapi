package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mokapi/media"
	"mokapi/sortedmap"
	"reflect"
	"unicode"
)

func (r *Ref) Marshal(i interface{}, contentType media.ContentType) ([]byte, error) {
	i, err := selectData(i, r)
	if err != nil {
		return nil, err
	}
	switch {
	case contentType.Subtype == "json":
		b, err := json.Marshal(i)
		if err, ok := err.(*json.SyntaxError); ok {
			return nil, fmt.Errorf("json error (%v): %v", err.Offset, err.Error())
		}
		return b, err
	case contentType.IsXml():
		return writeXml(i, r)
	default:
		if s, ok := i.(string); ok {
			return []byte(s), nil
		}
		return nil, fmt.Errorf("unspupported encoding for content type %v", contentType)
	}
}

type schemaObject struct {
	*sortedmap.LinkedHashMap
}

func newSchemaObject() *schemaObject {
	return &schemaObject{sortedmap.NewLinkedHashMap()}
}

// selectData selects data by schema
func selectData(data interface{}, schema *Ref) (interface{}, error) {
	if schema == nil || schema.Value == nil || data == nil {
		return data, nil
	}

	if len(schema.Value.AnyOf) > 0 {
		return selectAny(schema.Value, data)
	}
	if len(schema.Value.AllOf) > 0 {
		return selectAll(schema.Value, data)
	}

	if schema.Value == nil || schema.Value.Type == "" {
		return data, nil
	}

	switch data.(type) {
	case []interface{}:
		if schema.Value.Type != "array" {
			return nil, fmt.Errorf("found array, expected %v", schema)
		}
	case map[string]interface{}:
		if schema.Value.Type != "object" {
			return nil, fmt.Errorf("found object, expected %v", schema)
		}
	case struct{}:
		if schema.Value.Type != "object" {
			return nil, fmt.Errorf("found object, expected %v", schema)
		}
	}

	switch schema.Value.Type {
	case "number":
		return parseNumber(data, schema.Value)
	case "integer":
		return parseInteger(data, schema.Value)
	case "array":
		return selectArray(data, schema.Value)
	case "object":
		return selectObject(data, schema.Value)
	}

	return data, nil
}

func (m *schemaObject) MarshalJSON() ([]byte, error) {
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

func selectArray(data interface{}, schema *Schema) (interface{}, error) {
	if list, ok := data.([]interface{}); ok {
		result := make([]interface{}, len(list))
		for i, e := range list {
			var err error
			result[i], err = selectData(e, schema.Items)
			if err != nil {
				return nil, err
			}
		}
		return result, validateArray(result, schema)
	}
	if m, ok := data.(map[interface{}]interface{}); ok {
		result := make([]interface{}, 0)
		for _, v := range m {
			if i, err := selectData(v, schema.Items); err != nil {
				return nil, err
			} else {
				result = append(result, i)
			}
		}
		return result, validateArray(result, schema)
	}

	return nil, fmt.Errorf("unexpected type for schema type array")
}

func selectObject(data interface{}, schema *Schema) (interface{}, error) {
	if m, ok := data.(*sortedmap.LinkedHashMap); ok {
		return fromLinkedMap(m, schema)
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return fromStruct(v, schema)
	case reflect.Map:
		return fromMap(v, schema)
	}

	return nil, fmt.Errorf("could not encode '%v' to %v", data, schema)
}

func selectAny(schema *Schema, data interface{}) (interface{}, error) {
	result := newSchemaObject()
	for _, any := range schema.AnyOf {
		r, err := selectData(data, any)
		if err != nil {
			continue
		}
		o, ok := r.(*schemaObject)
		if !ok {
			if result.Len() > 0 {
				continue
			}
			return r, nil
		}
		result.Merge(o.LinkedHashMap)
	}
	return result, nil
}

func selectAll(schema *Schema, data interface{}) (interface{}, error) {
	r := newSchemaObject()
	schemas := make([]*Ref, 0, len(schema.AllOf)+1)
	schemas = append(schemas, schema.AllOf...)
	schemas = append(schemas, &Ref{Value: &Schema{Type: "object"}})

	for i, all := range schemas {
		if all == nil {
			continue
		}
		if all.Value.Type != "object" {
			return nil, fmt.Errorf("allOf only supports type of object")
		}
		s := all.Value
		if i < len(schemas)-1 {
			c := *all.Value
			c.AdditionalProperties = &AdditionalProperties{Forbidden: true}
			s = &c
		}

		o, err := selectObject(data, s)
		if err != nil {
			return nil, err
		}
		so := o.(*schemaObject)
		for it := so.LinkedHashMap.Iter(); it.Next(); {
			if v := r.Get(it.Key()); v == nil {
				r.Set(it.Key(), it.Value())
			}
		}
	}
	return r, nil
}

func fromLinkedMap(m *sortedmap.LinkedHashMap, schema *Schema) (*schemaObject, error) {
	if schema.IsDictionary() {
		return &schemaObject{m}, validateObject(m, schema)
	}

	obj := newSchemaObject()
	for it := m.Iter(); it.Next(); {
		name := it.Key().(string)

		var field *Ref
		if schema.Properties != nil {
			field = schema.Properties.Get(name)
		}

		if field != nil || schema.IsFreeForm() {
			d, err := selectData(it.Value(), field)
			if err != nil {
				return nil, err
			}
			obj.Set(name, d)
		}
	}
	return obj, validateObject(obj.LinkedHashMap, schema)
}

func fromStruct(v reflect.Value, schema *Schema) (*schemaObject, error) {
	t := v.Type()
	obj := newSchemaObject()
	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)
		name := unTitle(ft.Name)
		val := v.Field(i)

		if p := schema.Properties.Get(name); p != nil || schema.IsFreeForm() {
			d, err := selectData(val.Interface(), p)
			if err != nil {
				return nil, err
			}
			obj.Set(name, d)
		}
	}
	return obj, validateObject(obj.LinkedHashMap, schema)
}

func fromMap(v reflect.Value, schema *Schema) (*schemaObject, error) {
	obj := newSchemaObject()

	if schema.HasProperties() {
		for it := schema.Properties.Value.Iter(); it.Next(); {
			name := fmt.Sprintf("%v", it.Key())
			o := v.MapIndex(reflect.ValueOf(name))
			if !o.IsValid() {
				continue
			}
			d, err := selectData(o.Interface(), it.Value().(*Ref))
			if err != nil {
				return nil, err
			}
			obj.Set(name, d)
		}
	}

	if schema.IsFreeForm() || schema.IsDictionary() {
		for _, k := range v.MapKeys() {
			name := fmt.Sprintf("%v", k.Interface())
			if obj.Get(name) == nil {
				o := v.MapIndex(k)
				obj.Set(name, o.Interface())
			}
		}
	}

	return obj, validateObject(obj.LinkedHashMap, schema)
}
