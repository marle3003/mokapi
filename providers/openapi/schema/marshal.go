package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/media"
	"mokapi/schema/encoding"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
	"reflect"
	"strings"
)

func (s *Schema) Marshal(i interface{}, contentType media.ContentType) ([]byte, error) {
	if contentType.IsXml() {
		p := parser.Parser{ConvertStringToNumber: true, ConvertToSortedMap: true, ValidateAdditionalProperties: false}
		i, err := p.ParseWith(i, ConvertToJsonSchema(s))
		if err == nil {
			var b []byte
			b, err = marshalXml(i, s)
			if err == nil {
				return b, nil
			}
		}

		return nil, err
	}

	e := encoding.NewEncoder(ConvertToJsonSchema(s))
	return e.Write(i, contentType)
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	e := Encoder{visited: map[*Schema]bool{}}
	return e.ToJSON(s)
}

func (s *Schema) MarshalYAML() (interface{}, error) {
	e := Encoder{visited: map[*Schema]bool{}}
	return e.ToYAML(s)
}

type Encoder struct {
	KeepRef bool

	visited map[*Schema]bool
}

func (e *Encoder) ToYAML(s *Schema) (any, error) {
	if s == nil {
		return nil, nil
	}
	if s.Boolean != nil {
		return s.Boolean, nil
	}
	// check circular reference
	if e.hasVisited(s) {
		m := map[string]string{
			"description": "circular reference",
		}
		if s.Ref != "" {
			m["$ref"] = s.Ref
		}
		return m, nil
	}
	e.visited[s] = true
	defer delete(e.visited, s)

	result := &sortedmap.LinkedHashMap[string, any]{}
	v := reflect.ValueOf(s).Elem()
	t := v.Type()
	var err error
	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)
		if !ft.IsExported() {
			continue
		}
		f := v.FieldByName(ft.Name)
		if isEmptyValue(f) {
			continue
		}
		fv := f.Interface()
		var fieldValue any
		switch val := fv.(type) {
		case schema.Types:
			switch len(val) {
			case 0:
				continue
			case 1:
				fieldValue = val[0]
			default:
				fieldValue = val
			}
		case *Schemas:
			m := map[string]any{}
			for it := val.Iter(); it.Next(); {
				m[it.Key()], err = e.ToYAML(it.Value())
				if err != nil {
					return nil, err
				}
			}
			fieldValue = m
		case *Schema:
			fieldValue, err = e.ToYAML(val)
		case *schema.UnionType[float64, bool]:
			if val.IsA() {
				fieldValue = val.A
			} else {
				fieldValue = val.B
			}
		case dynamic.Reference[*Schema]:
			if val.Ref != "" && s.Sub == nil || e.KeepRef {
				result.Set("$ref", val.Ref)
			}
			continue
		case *schema.Example:
			if val.Value != nil {
				fieldValue = val.Value
			}
		default:
			fieldValue = val
		}

		if err != nil {
			return nil, err
		}

		tag := t.Field(i).Tag.Get("yaml")
		args := strings.Split(tag, ",")
		name := args[0]
		if name == "-" {
			continue
		}

		result.Set(name, fieldValue)
	}
	return result, nil
}

func (e *Encoder) ToJSON(s *Schema) ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}

	var b bytes.Buffer
	if s.Boolean != nil {
		b.Write([]byte(fmt.Sprintf("%v", *s.Boolean)))
		return b.Bytes(), nil
	}

	// check circular reference
	if e.hasVisited(s) {
		var v string
		if s.Ref != "" {
			v = fmt.Sprintf(`{"$ref":"%s","description":"circular reference"}`, s.Ref)

		} else {
			v = `{"description":"circular reference"}`
		}
		return []byte(v), nil
	}
	e.visited[s] = true
	defer delete(e.visited, s)

	b.WriteRune('{')

	v := reflect.ValueOf(s).Elem()
	t := v.Type()
	var err error
	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)
		if !ft.IsExported() {
			continue
		}
		f := v.FieldByName(ft.Name)
		if isEmptyValue(f) {
			continue
		}

		fv := f.Interface()
		var bVal []byte
		switch val := fv.(type) {
		case schema.Types:
			if len(val) == 0 {
				continue
			}
			bVal, err = val.MarshalJSON()
		case *Schemas:
			var fields bytes.Buffer
			fields.WriteRune('{')
			for it := val.Iter(); it.Next(); {
				if fields.Len() > 1 {
					fields.WriteRune(',')
				}
				sField, err := e.ToJSON(it.Value())
				if err != nil {
					return nil, err
				}
				fields.WriteString(fmt.Sprintf(`"%v":`, it.Key()))
				fields.Write(sField)
			}
			fields.WriteRune('}')
			bVal = fields.Bytes()
		case *Schema:
			bVal, err = e.ToJSON(val)
		case *schema.UnionType[float64, bool]:
			if val.IsA() {
				bVal, err = json.Marshal(val.A)
			} else {
				bVal, err = json.Marshal(val.B)
			}
		case dynamic.Reference[*Schema]:
			if val.Ref != "" && s.Sub == nil || e.KeepRef {
				if b.Len() > 1 {
					b.Write([]byte{','})
				}
				bVal, err = json.Marshal(val)
				b.WriteString(strings.Trim(string(bVal), "{}"))
			}
			continue
		default:
			bVal, err = json.Marshal(val)
		}

		if err != nil {
			return nil, err
		}

		tag := t.Field(i).Tag.Get("json")
		args := strings.Split(tag, ",")
		name := args[0]
		if name == "-" {
			continue
		}

		if b.Len() > 1 {
			b.Write([]byte{','})
		}

		b.WriteString(fmt.Sprintf(`"%v":`, name))
		b.Write(bVal)
	}

	b.WriteRune('}')
	return b.Bytes(), nil
}

func (e *Encoder) hasVisited(s *Schema) bool {
	if e.visited == nil {
		e.visited = map[*Schema]bool{}
		return false
	}
	return e.visited[s]
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Interface, reflect.Pointer:
		return v.IsZero()
	default:
		return false
	}
}
