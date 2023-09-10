package schema

import (
	"fmt"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/sortedmap"
	"strings"
)

type Ref struct {
	ref.Reference
	Value *Schema
}

type SchemasRef struct {
	ref.Reference
	Value *Schemas
}

type Schemas struct {
	sortedmap.LinkedHashMap
}

type Schema struct {
	Description string `yaml:"description" json:"description"`

	Type       string        `yaml:"type" json:"type"`
	AnyOf      []*Ref        `yaml:"anyOf" json:"anyOf"`
	AllOf      []*Ref        `yaml:"allOf" json:"allOf"`
	OneOf      []*Ref        `yaml:"oneOf" json:"oneOf"`
	Deprecated bool          `yaml:"deprecated" json:"deprecated"`
	Example    interface{}   `yaml:"example" json:"example"`
	Enum       []interface{} `yaml:"enum" json:"enum"`
	Xml        *Xml          `yaml:"xml" json:"xml"`
	Format     string        `yaml:"format" json:"format"`
	Nullable   bool          `yaml:"nullable" json:"nullable"`

	// String
	Pattern   string `yaml:"pattern" json:"pattern"`
	MinLength *int   `yaml:"minLength" json:"minLength"`
	MaxLength *int   `yaml:"maxLength" json:"maxLength"`

	// Numbers
	Minimum          *float64 `yaml:"minimum,omitempty" json:"minimum,omitempty"`
	Maximum          *float64 `yaml:"maximum,omitempty" json:"maximum,omitempty"`
	ExclusiveMinimum *bool    `yaml:"exclusiveMinimum,omitempty" json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *bool    `yaml:"exclusiveMaximum ,omitempty" json:"exclusiveMaximum,omitempty"`

	// Array
	Items        *Ref `yaml:"items" json:"items"`
	UniqueItems  bool `yaml:"uniqueItems" json:"uniqueItems"`
	MinItems     *int `yaml:"minItems" json:"minItems"`
	MaxItems     *int `yaml:"maxItems" json:"maxItems"`
	ShuffleItems bool `yaml:"x-shuffleItems" json:"x-shuffleItems"`

	// Object
	Properties           *SchemasRef           `yaml:"properties" json:"properties"`
	Required             []string              `yaml:"required" json:"required"`
	AdditionalProperties *AdditionalProperties `yaml:"additionalProperties,omitempty" json:"additionalProperties,omitempty"`
	MinProperties        *int                  `yaml:"minProperties" json:"minProperties"`
	MaxProperties        *int                  `yaml:"maxProperties" json:"maxProperties"`
}

type AdditionalProperties struct {
	*Ref
	Forbidden bool `yaml:"forbidden" json:"forbidden"`
}

type Xml struct {
	Wrapped   bool   `yaml:"wrapped" json:"wrapped"`
	Name      string `yaml:"name" json:"name"`
	Attribute bool   `yaml:"attribute" json:"attribute"`
	Prefix    string `yaml:"prefix" json:"prefix"`
	Namespace string `yaml:"namespace" json:"namespace"`
	CData     bool   `yaml:"x-cdata" json:"x-cdata"`
}

func (s *SchemasRef) Get(name string) *Ref {
	if s == nil && s.Value == nil {
		return nil
	}
	return s.Value.Get(name)
}

func (s *Schemas) Get(name string) *Ref {
	if s == nil {
		return nil
	}
	r := s.LinkedHashMap.Get(name)
	if r == nil {
		return nil
	}
	return r.(*Ref)
}

func (s *Schemas) Resolve(token string) (interface{}, error) {
	i := s.Get(token)
	if i == nil {
		return nil, fmt.Errorf("unable to resolve %v", token)
	}
	return i.Value, nil
}

func (r *Ref) HasProperties() bool {
	return r.Value != nil && r.Value.HasProperties()
}

func (s *Schema) HasProperties() bool {
	return s.Properties != nil && s.Properties.Value != nil && s.Properties.Value.Len() > 0
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
		sb.WriteString(fmt.Sprintf("schema type=%v", s.Type))
	}
	if len(s.Format) > 0 {
		sb.WriteString(fmt.Sprintf(" format=%v", s.Format))
	}
	if len(s.Pattern) > 0 {
		sb.WriteString(fmt.Sprintf(" pattern=%v", s.Pattern))
	}
	if s.Minimum != nil {
		sb.WriteString(fmt.Sprintf(" minimum=%v", *s.Minimum))
	}
	if s.Maximum != nil {
		sb.WriteString(fmt.Sprintf(" maximum=%v", *s.Maximum))
	}
	if s.ExclusiveMinimum != nil && *s.ExclusiveMinimum {
		sb.WriteString(" exclusiveMinimum")
	}
	if s.ExclusiveMaximum != nil && *s.ExclusiveMaximum {
		sb.WriteString(" exclusiveMaximum")
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

	if s.Type == "object" && s.Properties != nil && s.Properties.Value != nil {
		var sbProp strings.Builder
		for _, p := range s.Properties.Value.Keys() {
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
	if s.Type == "object" && s.IsFreeForm() {
		sb.WriteString(" free-form=true")
	}

	return sb.String()
}

func (s *Schema) IsFreeForm() bool {
	free := s.Type == "object" && (s.Properties == nil || s.Properties.Value.Len() == 0)
	if s.AdditionalProperties == nil {
		return free
	}
	return s.AdditionalProperties.IsFreeForm()
}

func (s *Schema) IsDictionary() bool {
	return s.AdditionalProperties != nil && s.AdditionalProperties.Ref != nil && s.AdditionalProperties.Value != nil && s.AdditionalProperties.Value.Type != ""
}

func (ap *AdditionalProperties) IsFreeForm() bool {
	if ap == nil {
		return false
	}
	if ap.Ref == nil || ap.Value == nil {
		return !ap.Forbidden
	}
	if ap.Value != nil && ap.Value.Type == "" {
		return true
	}
	return false
}
