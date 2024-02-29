package schema

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"mokapi/media"
	"mokapi/sortedmap"
	"reflect"
	"strconv"
	"unicode"
)

const marshalErrorFormat = "marshal data to '%v' failed: %w"

type marshalObject struct {
	*sortedmap.LinkedHashMap[string, interface{}]
}

func (r *Ref) Marshal(i interface{}, contentType media.ContentType) ([]byte, error) {
	s := selector{}
	if contentType.Subtype != "json" {
		s.convertStringToNumber = true
	}

	i, err := s.selectData(i, r)
	if err != nil {
		return nil, fmt.Errorf(marshalErrorFormat, contentType.String(), err)
	}
	var b []byte
	switch {
	case contentType.Subtype == "json" || contentType.Subtype == "problem+json":
		b, err = json.Marshal(i)
	case contentType.IsXml():
		b, err = marshalXml(i, r)
	default:
		var s string
		switch i.(type) {
		case string:
			s = i.(string)
		case float64:
			s = strconv.FormatFloat(i.(float64), 'f', -1, 64)
		case float32:
			s = strconv.FormatFloat(float64(i.(float32)), 'f', -1, 32)
		case int, int32, int64:
			s = fmt.Sprintf("%v", i)
		default:
			err = fmt.Errorf("unspupported encoding for content type %v", contentType)
		}
		b = []byte(s)
	}

	if err != nil {
		return nil, fmt.Errorf(marshalErrorFormat, contentType.String(), err)
	}
	return b, nil
}

func (s *Schema) Marshal(i interface{}, contentType media.ContentType) ([]byte, error) {
	r := &Ref{Value: s}
	return r.Marshal(i, contentType)
}

type selector struct {
	convertStringToNumber bool
}

// selectData selects data by schema
func (s *selector) selectData(data interface{}, ref *Ref) (interface{}, error) {
	if ref == nil || ref.Value == nil || data == nil {
		return data, nil
	}

	schema := ref.Value

	switch {
	case len(schema.AnyOf) > 0:
		return s.selectAny(schema, data)
	case len(schema.AllOf) > 0:
		return s.selectAll(schema, data)
	case len(schema.OneOf) > 0:
		return s.selectOne(schema, data)
	}

	if schema.Type == "" {
		return data, nil
	}

	switch data.(type) {
	case []interface{}:
		if schema.Type != "array" {
			return nil, fmt.Errorf("found array, expected %v: %v", schema, data)
		}
	case map[string]interface{}:
		if schema.Type != "object" {
			return nil, fmt.Errorf("found object, expected %v: %v", schema, data)
		}
	case struct{}:
		if schema.Type != "object" {
			return nil, fmt.Errorf("found object, expected %v: %v", schema, data)
		}
	}

	p := parser{convertStringToNumber: s.convertStringToNumber}
	switch schema.Type {
	case "string":
		return data, validateString(data, schema)
	case "number":
		return p.parseNumber(data, schema)
	case "integer":
		return p.parseInteger(data, schema)
	case "array":
		return s.selectArray(data, schema)
	case "object":
		return s.selectObject(data, schema)
	}

	return data, nil
}

// UnTitle returns a copy of the string s with first letter mapped to its Unicode lower case
func unTitle(s string) string {
	for i, v := range s {
		return string(unicode.ToLower(v)) + s[i+1:]
	}
	return s
}

func (s *selector) selectArray(data interface{}, schema *Schema) (interface{}, error) {
	var result []interface{}

	switch list := data.(type) {
	case []interface{}:
		result = make([]interface{}, len(list))
		for i, e := range list {
			var err error
			result[i], err = s.selectData(e, schema.Items)
			if err != nil {
				return nil, err
			}
		}
	case map[interface{}]interface{}:
		result = make([]interface{}, len(list))
		for _, v := range list {
			if i, err := s.selectData(v, schema.Items); err != nil {
				return nil, err
			} else {
				result = append(result, i)
			}
		}
	default:
		return nil, fmt.Errorf("expected array but got %T", data)
	}

	if err := validateArray(result, schema); err != nil {
		return nil, fmt.Errorf("does not match %s: %w", schema, err)
	}

	return result, nil
}

func (s *selector) selectObject(data interface{}, schema *Schema) (interface{}, error) {
	if m, ok := data.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
		return s.fromLinkedMap(m, schema)
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return s.fromStruct(v, schema)
	case reflect.Map:
		return s.reflectFromMap(v, schema)
	}

	return nil, fmt.Errorf("encode '%v' to %v failed", data, schema)
}

func (s *selector) selectAny(schema *Schema, data interface{}) (interface{}, error) {
	result := newMarshalObject()

	isFreeFormUsed := false
	for _, any := range schema.AnyOf {
		if any == nil || any.Value == nil {
			return s.selectData(data, nil)
		}

		if any.Value.IsFreeForm() {
			// first we read data without free-form to get best matching data types
			// after we processed all schemas we read with free-form
			copySchema := *any.Value
			any = &Ref{Value: &copySchema}
			isFreeFormUsed = true
			any.Value.AdditionalProperties = &AdditionalProperties{Forbidden: true}
		}

		r, err := s.selectData(data, any)
		if err != nil {
			continue
		}
		o, ok := r.(*marshalObject)
		if !ok {
			return r, nil
		}
		for it := o.LinkedHashMap.Iter(); it.Next(); {
			if _, found := result.Get(it.Key()); !found {
				result.Set(it.Key(), it.Value())
			}
		}
	}

	if isFreeFormUsed {
		// read data with free-form and add only missing values
		// free-form returns never an error
		i, _ := s.selectObject(data, &Schema{Type: "object"})
		o := i.(*marshalObject)
		for it := o.Iter(); it.Next(); {
			if _, found := result.Get(it.Key()); !found {
				result.Set(it.Key(), it.Value())
			}
		}
	}

	return result, nil
}

func (s *selector) selectOne(schema *Schema, data interface{}) (interface{}, error) {
	var result interface{}

	for _, one := range schema.OneOf {
		if one == nil || one.Value == nil {
			next, err := s.selectData(data, nil)
			if err != nil {
				continue
			}
			if result != nil {
				return nil, fmt.Errorf("oneOf can only match exactly one schema")
			}
			result = next
			continue
		}

		next, err := s.selectData(data, one)
		if err != nil {
			continue
		}
		if obj, ok := next.(*marshalObject); ok && obj.Len() == 0 {
			// empty object does not match
			continue
		}
		if result != nil {
			return nil, fmt.Errorf("oneOf can only match exactly one schema")
		}
		result = next
	}

	return result, nil
}

func (s *selector) selectAll(schema *Schema, data interface{}) (interface{}, error) {
	r := newMarshalObject()

	isFreeFormUsed := false
	for _, all := range schema.AllOf {
		if all == nil || all.Value == nil {
			return nil, fmt.Errorf("schema is not defined: allOf only supports type of object")
		}
		if all.Value.Type != "object" {
			return nil, fmt.Errorf("type of '%v' is not allowed: allOf only supports type of object", all.Value.Type)
		}

		origin := all
		if all.Value.IsFreeForm() {
			// first we read data without free-form to get best matching data types
			// after we processed all schemas we read with free-form
			copySchema := *all.Value
			all = &Ref{Value: &copySchema}
			isFreeFormUsed = true
			all.Value.AdditionalProperties = &AdditionalProperties{Forbidden: true}
		}

		o, err := s.selectObject(data, all.Value)
		if err != nil {
			err := errors.Unwrap(err)
			return nil, fmt.Errorf("does not match %v: %w", origin, err)
		}
		so := o.(*marshalObject)
		for it := so.LinkedHashMap.Iter(); it.Next(); {
			if _, found := r.Get(it.Key()); !found {
				r.Set(it.Key(), it.Value())
			}
		}
	}

	if isFreeFormUsed {
		// read data with free-form and add only missing values
		// free-form returns never an error
		i, _ := s.selectObject(data, &Schema{Type: "object"})
		o := i.(*marshalObject)
		for it := o.Iter(); it.Next(); {
			if _, found := r.Get(it.Key()); !found {
				r.Set(it.Key(), it.Value())
			}
		}
	}

	return r, nil
}

func (s *selector) fromLinkedMap(m *sortedmap.LinkedHashMap[string, interface{}], schema *Schema) (*marshalObject, error) {
	if schema.IsDictionary() {
		return &marshalObject{m}, validateObject(m, schema)
	}

	obj := newMarshalObject()
	for it := m.Iter(); it.Next(); {
		name := it.Key()

		var field *Ref
		if schema.Properties != nil {
			field = schema.Properties.Get(name)
		}

		if field != nil || schema.IsFreeForm() {
			d, err := s.selectData(it.Value(), field)
			if err != nil {
				return nil, err
			}
			obj.Set(name, d)
		}
	}

	if err := validateObject(obj.LinkedHashMap, schema); err != nil {
		return nil, fmt.Errorf("does not match %s: %w", schema, err)
	}

	return obj, nil
}

func (s *selector) fromStruct(v reflect.Value, schema *Schema) (*marshalObject, error) {
	t := v.Type()
	obj := newMarshalObject()
	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)
		name := unTitle(ft.Name)
		val := v.Field(i)

		if p := schema.Properties.Get(name); p != nil || schema.IsFreeForm() {
			d, err := s.selectData(val.Interface(), p)
			if err != nil {
				return nil, fmt.Errorf("encode property '%v' failed: %w", name, err)
			}
			obj.Set(name, d)
		}
	}

	if err := validateObject(obj.LinkedHashMap, schema); err != nil {
		return nil, fmt.Errorf("does not match %s: %w", schema, err)
	}

	return obj, nil
}

func (s *selector) reflectFromMap(v reflect.Value, schema *Schema) (*marshalObject, error) {
	obj := newMarshalObject()

	if schema.HasProperties() {
		for it := schema.Properties.Iter(); it.Next(); {
			name := it.Key()
			o := v.MapIndex(reflect.ValueOf(name))
			if !o.IsValid() {
				continue
			}
			d, err := s.selectData(o.Interface(), it.Value())
			if err != nil {
				return nil, fmt.Errorf("parsing property '%v' failed: %w", name, err)
			}
			obj.Set(name, d)
		}
	}

	if schema.IsDictionary() {
		for _, k := range v.MapKeys() {
			name := fmt.Sprintf("%v", k.Interface())
			if _, found := obj.Get(name); !found {
				o := v.MapIndex(k)
				d, err := s.selectData(o.Interface(), schema.AdditionalProperties.Ref)
				if err != nil {
					_, err := s.selectData(o.Interface(), schema.AdditionalProperties.Ref)
					return nil, err
				}
				obj.Set(name, d)
			}
		}
	}

	if schema.IsFreeForm() || schema.IsDictionary() {
		for _, k := range v.MapKeys() {
			name := fmt.Sprintf("%v", k.Interface())
			if _, found := obj.Get(name); !found {
				o := v.MapIndex(k)
				obj.Set(name, o.Interface())
			}
		}
	}

	if err := validateObject(obj.LinkedHashMap, schema); err != nil {
		return nil, fmt.Errorf("does not match %s: %w", schema, err)
	}

	return obj, nil
}

func newMarshalObject() *marshalObject {
	return &marshalObject{sortedmap.NewLinkedHashMap()}
}
