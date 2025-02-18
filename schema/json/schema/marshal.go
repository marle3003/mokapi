package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func (s *Schema) MarshalJSON() ([]byte, error) {
	e := encoder{refs: map[string]bool{}}
	return e.encode(s)
}

type encoder struct {
	refs map[string]bool
}

func (e *encoder) encode(r *Schema) ([]byte, error) {
	var b bytes.Buffer
	if r.Boolean != nil {
		b.Write([]byte(fmt.Sprintf("%v", *r.Boolean)))
		return b.Bytes(), nil
	}

	b.WriteRune('{')

	if r.Ref != "" {
		// loop protection, only return reference
		if _, ok := e.refs[r.Ref]; ok {
			b.Write([]byte(fmt.Sprintf(`"$ref":"%v"`, r.Ref)))

			b.WriteRune('}')
			return b.Bytes(), nil
		}
		e.refs[r.Ref] = true
		defer func() {
			delete(e.refs, r.Ref)
		}()
	}

	if r != nil {
		v := reflect.ValueOf(r).Elem()
		t := v.Type()
		var err error
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if isEmptyValue(f) {
				continue
			}

			fv := f.Interface()
			var bVal []byte
			switch val := fv.(type) {
			case Types:
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

	b.WriteRune('}')
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
