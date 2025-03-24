package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mokapi/media"
	"mokapi/schema/encoding"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
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

		if uw, ok := err.(interface{ Unwrap() []error }); ok {
			errs := uw.Unwrap()
			if len(errs) > 1 {
				return nil, fmt.Errorf("encoding data to '%v' failed:\n %w", contentType.String(), err)
			}
		}

		return nil, fmt.Errorf("encoding data to '%v' failed: %w", contentType, err)
	}

	e := encoding.NewEncoder(ConvertToJsonSchema(s))
	return e.Write(i, contentType)
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	e := encoder{refs: map[string]bool{}}
	return e.encode(s)
}

type encoder struct {
	refs map[string]bool
}

func (e *encoder) encode(s *Schema) ([]byte, error) {
	var b bytes.Buffer
	if s.SubSchema != nil && s.SubSchema.Boolean != nil {
		b.Write([]byte(fmt.Sprintf("%v", *s.Boolean)))
		return b.Bytes(), nil
	}

	b.WriteRune('{')

	if s.Ref != "" {
		// loop protection, only return reference
		if _, ok := e.refs[s.Ref]; ok {
			b.Write([]byte(fmt.Sprintf(`"$ref":"%v"`, s.Ref)))

			b.WriteRune('}')
			return b.Bytes(), nil
		}
		e.refs[s.Ref] = true
		defer func() {
			delete(e.refs, s.Ref)
		}()
	}

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
		case *SubSchema:
			bVal, err = e.encodeSub(val)
			if err != nil {
				return nil, err
			}
			if len(bVal) > 0 {
				if b.Len() > 1 {
					b.Write([]byte{','})
				}
				b.Write(bVal)
			}
		default:
			bVal, err = json.Marshal(val)
			if err != nil {
				return nil, err
			}
			tag := t.Field(i).Tag.Get("json")
			args := strings.Split(tag, ",")
			name := args[0]

			if b.Len() > 1 {
				b.Write([]byte{','})
			}

			b.WriteString(fmt.Sprintf(`"%v":`, name))
			b.Write(bVal)
		}
	}

	b.WriteRune('}')
	return b.Bytes(), nil
}

func (e *encoder) encodeSub(s *SubSchema) ([]byte, error) {
	var b bytes.Buffer

	if s != nil {
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
					sField, err := e.encode(it.Value())
					if err != nil {
						return nil, err
					}
					fields.WriteString(fmt.Sprintf(`"%v":`, it.Key()))
					fields.Write(sField)
				}
				fields.WriteRune('}')
				bVal = fields.Bytes()
			case *Schema:
				bVal, err = e.encode(val)
			case *SubSchema:
				bVal, err = e.encodeSub(val)
			default:
				bVal, err = json.Marshal(val)
			}

			if err != nil {
				return nil, err
			}

			if b.Len() > 1 {
				b.Write([]byte{','})
			}

			tag := t.Field(i).Tag.Get("json")
			args := strings.Split(tag, ",")
			name := args[0]

			b.WriteString(fmt.Sprintf(`"%v":`, name))
			b.Write(bVal)
		}
	}

	return b.Bytes(), nil
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
