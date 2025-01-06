package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/schema/json/schema"
	"reflect"
	"strings"
)

type Schema struct {
	Schema  string `yaml:"$schema,omitempty" json:"$schema,omitempty"`
	Boolean *bool  `yaml:"-" json:"-"`

	Type  schema.Types  `yaml:"type,omitempty" json:"type,omitempty"`
	Enum  []interface{} `yaml:"enum,omitempty" json:"enum,omitempty"`
	Const *interface{}  `yaml:"const,omitempty" json:"const,omitempty"`

	// Numbers
	MultipleOf       *float64                         `yaml:"multipleOf,omitempty" json:"multipleOf,omitempty"`
	Minimum          *float64                         `yaml:"minimum,omitempty" json:"minimum,omitempty"`
	Maximum          *float64                         `yaml:"maximum,omitempty" json:"maximum,omitempty"`
	ExclusiveMinimum *schema.UnionType[float64, bool] `yaml:"exclusiveMinimum,omitempty" json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *schema.UnionType[float64, bool] `yaml:"exclusiveMaximum,omitempty" json:"exclusiveMaximum,omitempty"`

	// String
	Pattern   string `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	MinLength *int   `yaml:"minLength,omitempty" json:"minLength,omitempty"`
	MaxLength *int   `yaml:"maxLength,omitempty" json:"maxLength,omitempty"`
	Format    string `yaml:"format,omitempty" json:"format,omitempty"`

	// Array
	Items            *Ref   `yaml:"items,omitempty" json:"items,omitempty"`
	PrefixItems      []*Ref `yaml:"prefixItems,omitempty" json:"prefixItems,omitempty"`
	UnevaluatedItems *Ref   `yaml:"unevaluatedItems,omitempty" json:"unevaluatedItems,omitempty"`
	Contains         *Ref   `yaml:"contains,omitempty" json:"contains,omitempty"`
	MaxContains      *int   `yaml:"maxContains,omitempty" json:"maxContains,omitempty"`
	MinContains      *int   `yaml:"minContains,omitempty" json:"minContains,omitempty"`
	MinItems         *int   `yaml:"minItems,omitempty" json:"minItems,omitempty"`
	MaxItems         *int   `yaml:"maxItems,omitempty" json:"maxItems,omitempty"`
	UniqueItems      bool   `yaml:"uniqueItems,omitempty" json:"uniqueItems,omitempty"`
	ShuffleItems     bool   `yaml:"x-shuffleItems,omitempty" json:"x-shuffleItems,omitempty"`

	// Object
	Properties            *Schemas            `yaml:"properties,omitempty" json:"properties,omitempty"`
	PatternProperties     map[string]*Ref     `yaml:"patternProperties,omitempty" json:"patternProperties,omitempty"`
	MinProperties         *int                `yaml:"minProperties,omitempty" json:"minProperties,omitempty"`
	MaxProperties         *int                `yaml:"maxProperties,omitempty" json:"maxProperties,omitempty"`
	Required              []string            `yaml:"required,omitempty" json:"required,omitempty"`
	DependentRequired     map[string][]string `yaml:"dependentRequired,omitempty" json:"dependentRequired,omitempty"`
	DependentSchemas      map[string]*Ref     `yaml:"dependentSchemas,omitempty" json:"dependentSchemas,omitempty"`
	AdditionalProperties  *Ref                `yaml:"additionalProperties,omitempty" json:"additionalProperties,omitempty"`
	UnevaluatedProperties *Ref                `yaml:"unevaluatedProperties,omitempty" json:"unevaluatedProperties,omitempty"`
	PropertyNames         *Ref                `yaml:"propertyNames,omitempty" json:"propertyNames,omitempty"`

	AnyOf []*Ref `yaml:"anyOf,omitempty" json:"anyOf,omitempty"`
	AllOf []*Ref `yaml:"allOf,omitempty" json:"allOf,omitempty"`
	OneOf []*Ref `yaml:"oneOf,omitempty" json:"oneOf,omitempty"`
	Not   *Ref   `yaml:"not,omitempty" json:"not,omitempty"`

	If   *Ref `yaml:"if,omitempty" json:"if,omitempty"`
	Then *Ref `yaml:"then,omitempty" json:"then,omitempty"`
	Else *Ref `yaml:"else,omitempty" json:"else,omitempty"`

	// Annotations
	Title       string        `yaml:"title,omitempty" json:"title,omitempty"`
	Description string        `yaml:"description,omitempty" json:"description,omitempty"`
	Default     interface{}   `yaml:"default,omitempty" json:"default,omitempty"`
	Deprecated  bool          `yaml:"deprecated,omitempty" json:"deprecated,omitempty"`
	Examples    []interface{} `yaml:"examples,omitempty" json:"examples,omitempty"`
	Example     interface{}   `yaml:"example,omitempty" json:"example,omitempty"`

	// Media
	ContentMediaType string `yaml:"contentMediaType,omitempty" json:"contentMediaType,omitempty"`
	ContentEncoding  string `yaml:"contentEncoding,omitempty" json:"contentEncoding,omitempty"`

	// OpenAPI
	Xml      *Xml `yaml:"xml,omitempty" json:"xml,omitempty"`
	Nullable bool `yaml:"nullable,omitempty" json:"nullable,omitempty"`
}

func (s *Schema) HasProperties() bool {
	return s.Properties != nil && s.Properties.Len() > 0
}

func (s *Schema) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if s == nil {
		return nil
	}

	if err := s.Items.Parse(config, reader); err != nil {
		return err
	}

	if err := s.Properties.Parse(config, reader); err != nil {
		return err
	}

	if err := s.AdditionalProperties.Parse(config, reader); err != nil {
		return err
	}

	for _, r := range s.AnyOf {
		if err := r.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, r := range s.AllOf {
		if err := r.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, r := range s.OneOf {
		if err := r.Parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (s *Schema) String() string {
	var sb strings.Builder

	if len(s.AnyOf) > 0 {
		sb.WriteString("any of ")
		for _, i := range s.AnyOf {
			if sb.Len() > 7 {
				sb.WriteString(", ")
			}
			sb.WriteString(i.String())
		}
		return sb.String()
	}
	if len(s.AllOf) > 0 {
		sb.WriteString("all of ")
		for _, i := range s.AllOf {
			if sb.Len() > 7 {
				sb.WriteString(", ")
			}
			sb.WriteString(i.String())
		}
		return sb.String()
	}
	if len(s.OneOf) > 0 {
		sb.WriteString("one of ")
		for _, i := range s.OneOf {
			if sb.Len() > 7 {
				sb.WriteString(", ")
			}
			sb.WriteString(i.String())
		}
		return sb.String()
	}

	if len(s.Type) > 0 {
		sb.WriteString(fmt.Sprintf("schema type=%v", s.Type.String()))
	}

	if len(s.Format) > 0 {
		sb.WriteString(fmt.Sprintf(" format=%v", s.Format))
	}
	if len(s.Pattern) > 0 {
		sb.WriteString(fmt.Sprintf(" pattern=%v", s.Pattern))
	}
	if s.MinLength != nil {
		sb.WriteString(fmt.Sprintf(" minLength=%v", *s.MinLength))
	}
	if s.MaxLength != nil {
		sb.WriteString(fmt.Sprintf(" maxLength=%v", *s.MaxLength))
	}

	if s.ExclusiveMinimum != nil {
		if s.ExclusiveMinimum.IsA() {
			sb.WriteString(fmt.Sprintf(" exclusiveMinimum=%v", s.ExclusiveMinimum.Value()))
		} else if s.ExclusiveMinimum.B {
			sb.WriteString(fmt.Sprintf(" exclusiveMinimum=%v", *s.Minimum))
		}
	} else if s.Minimum != nil {
		sb.WriteString(fmt.Sprintf(" minimum=%v", *s.Minimum))
	}

	if s.ExclusiveMaximum != nil {
		if s.ExclusiveMaximum.IsA() {
			sb.WriteString(fmt.Sprintf(" exclusiveMaximum=%v", s.ExclusiveMaximum.Value()))
		} else if s.ExclusiveMaximum.B {
			sb.WriteString(fmt.Sprintf(" exclusiveMaximum=%v", *s.Maximum))
		}
	} else if s.Maximum != nil {
		sb.WriteString(fmt.Sprintf(" maximum=%v", *s.Maximum))
	}

	if s.MinItems != nil {
		sb.WriteString(fmt.Sprintf(" minItems=%v", *s.MinItems))
	}
	if s.MaxItems != nil {
		sb.WriteString(fmt.Sprintf(" maxItems=%v", *s.MaxItems))
	}
	if s.MinProperties != nil {
		sb.WriteString(fmt.Sprintf(" minProperties=%v", *s.MinProperties))
	}
	if s.MaxProperties != nil {
		sb.WriteString(fmt.Sprintf(" maxProperties=%v", *s.MaxProperties))
	}
	if s.UniqueItems {
		sb.WriteString(" unique-items")
	}

	if s.Type.Includes("object") && s.Properties != nil {
		var sbProp strings.Builder
		for _, p := range s.Properties.Keys() {
			if sbProp.Len() > 0 {
				sbProp.WriteString(", ")
			}
			sbProp.WriteString(fmt.Sprintf("%v", p))
		}
		sb.WriteString(fmt.Sprintf(" properties=[%v]", sbProp.String()))
	}
	if len(s.Required) > 0 {
		sb.WriteString(fmt.Sprintf(" required=%v", s.Required))
	}
	if s.Type.Includes("object") && !s.IsFreeForm() {
		sb.WriteString(" free-form=false")
	}

	if s.Type.Includes("array") && s.Items != nil {
		sb.WriteString(" items=")
		sb.WriteString(s.Items.String())
	}

	if len(s.Title) > 0 {
		sb.WriteString(fmt.Sprintf(" title=%v", s.Title))
	} else if len(s.Description) > 0 {
		sb.WriteString(fmt.Sprintf(" description=%v", s.Description))
	}

	return sb.String()
}

func (s *Schema) IsFreeForm() bool {
	if !s.Type.Includes("object") {
		return false
	}
	free := s.Type.Includes("object") && (s.Properties == nil || s.Properties.Len() == 0)
	if s.AdditionalProperties == nil || free {
		return true
	}
	return s.AdditionalProperties.IsFreeForm()
}

func (s *Schema) IsDictionary() bool {
	return s.AdditionalProperties != nil && s.AdditionalProperties.Value != nil && len(s.AdditionalProperties.Value.Type) > 0
}

func (s *Schema) IsNullable() bool {
	return s.Nullable || s.Type.IsNullable()
}

func (s *Schema) ConvertTo(i interface{}) (interface{}, error) {
	if _, ok := i.(*schema.Schema); ok {
		return ConvertToJsonSchema(&Ref{Value: s}).Value, nil
	}
	return nil, fmt.Errorf("cannot convert %v to json schema", i)
}

func (s *Schema) UnmarshalJSON(b []byte) error {
	var boolVal bool
	if err := json.Unmarshal(b, &boolVal); err == nil {
		s.Boolean = &boolVal
		return nil
	}

	type alias Schema
	a := alias{}
	err := dynamic.UnmarshalJSON(b, &a)
	if err != nil {
		return err
	}
	*s = Schema(a)
	return nil
}

func (s *Schema) UnmarshalYAML(node *yaml.Node) error {
	var boolVal bool
	if err := node.Decode(&boolVal); err == nil {
		s.Boolean = &boolVal
		return nil
	}

	type alias Schema
	a := alias{}
	err := node.Decode(&a)
	if err != nil {
		return err
	}
	*s = Schema(a)
	return nil
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
			case schema.Types:
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
