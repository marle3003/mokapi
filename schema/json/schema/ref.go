package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/schema/json/ref"
	"reflect"
	"strings"
)

type Ref struct {
	ref.Reference
	Boolean *bool
	Value   *Schema
}

func (r *Ref) IsAny() bool {
	return r == nil || r.Value == nil || len(r.Value.Type) == 0
}

func (r *Ref) IsString() bool {
	return r != nil && r.Value != nil && r.Value.IsString()
}

func (r *Ref) IsInteger() bool {
	return r != nil && r.Value != nil && r.Value.IsInteger()
}

func (r *Ref) IsNumber() bool {
	return r != nil && r.Value != nil && r.Value.IsNumber()
}

func (r *Ref) IsArray() bool {
	return r != nil && r.Value != nil && r.Value.IsArray()
}

func (r *Ref) IsObject() bool {
	return r != nil && r.Value != nil && r.Value.IsObject()
}

func (r *Ref) IsNullable() bool {
	return r != nil && r.Value != nil && r.Value.IsNullable()
}

func (r *Ref) HasProperties() bool {
	return r != nil && r.Value != nil && r.Value.HasProperties()
}

func (r *Ref) IsAnyString() bool {
	return r != nil && r.Value != nil && r.Value.IsAnyString()
}

func (r *Ref) IsOneOf(typeNames ...string) bool {
	return r != nil && r.Value != nil && r.Value.Type.IsOneOf(typeNames...)
}

func (r *Ref) Type() string {
	if r == nil || r.Value == nil {
		return ""
	}
	return fmt.Sprintf("%s", r.Value.Type)
}

func (r *Ref) String() string {
	if r == nil || r.Value == nil {
		return "empty schema"
	}
	return r.Value.String()
}

func (r *Ref) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}
	if len(r.Ref) > 0 {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
	}

	if r.Value == nil {
		return nil
	}

	return r.Value.Parse(config, reader)
}

func (r *Ref) IsFreeForm() bool {
	if r == nil {
		return true
	}
	if r.Boolean != nil {
		return *r.Boolean
	}
	return r.Value.IsFreeForm()
}

func (r *Ref) UnmarshalJSON(b []byte) error {
	var boolVal bool
	if err := json.Unmarshal(b, &boolVal); err == nil {
		r.Boolean = &boolVal
		return nil
	}

	return r.UnmarshalJson(b, &r.Value)
}

func (r *Ref) UnmarshalYAML(node *yaml.Node) error {
	var boolVal bool
	if err := node.Decode(&boolVal); err == nil {
		r.Boolean = &boolVal
		return nil
	}

	return r.UnmarshalYaml(node, &r.Value)
}

func NewRef(b bool) *Ref {
	return &Ref{Boolean: &b}
}

func (r *Ref) IsFalse() bool {
	if r == nil {
		return false
	}
	if r.Boolean != nil {
		return !*r.Boolean
	}
	return r.Value.IsFalse()
}

func (r *Ref) MarshalJSON() ([]byte, error) {
	e := encoder{refs: map[string]bool{}}
	return e.encode(r)
}

type encoder struct {
	refs map[string]bool
}

func (e *encoder) encode(r *Ref) ([]byte, error) {
	var b bytes.Buffer
	if r.Boolean != nil {
		b.Write([]byte(fmt.Sprintf("%v", *r.Boolean)))
		return b.Bytes(), nil
	}

	b.WriteRune('{')

	if r.Ref != "" {
		b.Write([]byte(fmt.Sprintf(`"ref":"%v"`, r.Ref)))

		// loop protection, only return reference
		if _, ok := e.refs[r.Ref]; ok {
			b.WriteRune('}')
			return b.Bytes(), nil
		}
		e.refs[r.Ref] = true
		defer func() {
			delete(e.refs, r.Ref)
		}()
	}

	if r.Value != nil {
		v := reflect.ValueOf(r.Value).Elem()
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
			case *Ref:
				if val == nil {
					continue
				}
				bVal, err = e.encode(val)
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
			name := strings.Split(tag, ",")[0]

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
