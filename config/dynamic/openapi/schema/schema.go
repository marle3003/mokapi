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
	Type                 string
	Format               string
	Pattern              string
	Description          string
	Properties           *SchemasRef
	AdditionalProperties *AdditionalProperties `yaml:"additionalProperties,omitempty" json:"additionalProperties,omitempty"`
	Faker                string                `yaml:"x-faker" json:"x-faker"`
	Items                *Ref
	Xml                  *Xml
	Required             []string
	Nullable             bool
	Example              interface{}
	Enum                 []interface{}
	Minimum              *float64 `yaml:"minimum,omitempty" json:"minimum,omitempty"`
	Maximum              *float64 `yaml:"maximum,omitempty" json:"maximum,omitempty"`
	ExclusiveMinimum     *bool    `yaml:"exclusiveMinimum,omitempty" json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     *bool    `yaml:"exclusiveMaximum ,omitempty" json:"exclusiveMaximum,omitempty"`
	AnyOf                []*Ref   `yaml:"anyOf" json:"anyOf"`
	AllOf                []*Ref   `yaml:"allOf" json:"allOf"`
	OneOf                []*Ref   `yaml:"oneOf" json:"oneOf"`
	UniqueItems          bool     `yaml:"uniqueItems" json:"uniqueItems"`
	MinItems             *int     `yaml:"minItems" json:"minItems"`
	MaxItems             *int     `yaml:"maxItems" json:"maxItems"`
	ShuffleItems         bool     `yaml:"x-shuffleItems" json:"x-shuffleItems"`
	MinProperties        *int     `yaml:"minProperties" json:"minProperties"`
	MaxProperties        *int     `yaml:"maxProperties" json:"maxProperties"`
}

type AdditionalProperties struct {
	*Ref
}

type Xml struct {
	Wrapped   bool
	Name      string
	Attribute bool
	Prefix    string
	Namespace string
	CData     bool `yaml:"x-cdata" json:"x-cdata"`
}

func (s *SchemasRef) Get(name string) *Ref {
	if s.Value == nil {
		return nil
	}
	r := s.Value.Get(name)
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
	return i.(*Ref).Value, nil
}

func (r *Ref) HasProperties() bool {
	return r.Value != nil && r.Value.HasProperties()
}

func (s *Schema) HasProperties() bool {
	return s.Properties != nil && s.Properties.Value != nil && s.Properties.Value.Len() > 0
}

func (s *Schema) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("schema type=%v", s.Type))
	if len(s.Format) > 0 {
		sb.WriteString(fmt.Sprintf(" format=%v", s.Format))
	}
	if len(s.Pattern) > 0 {
		sb.WriteString(fmt.Sprintf(" pattern=%v", s.Format))
	}
	if s.Minimum != nil {
		sb.WriteString(fmt.Sprintf(" minimum=%v", *s.Minimum))
	}
	if s.Maximum != nil {
		sb.WriteString(fmt.Sprintf(" maximum=%v", *s.Minimum))
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
	if len(s.Required) > 0 {
		sb.WriteString(fmt.Sprintf(" required=%v", s.Required))
	}
	if s.Type == "object" && s.IsFreeForm() {
		sb.WriteString(" free-form=true")
	}
	return sb.String()
}

func (s *Schema) IsFreeForm() bool {
	return s.Type == "object" && !s.HasProperties() ||
		s.AdditionalProperties.IsFreeForm()
}

func (s *Schema) IsDictionary() bool {
	return s.AdditionalProperties != nil && s.AdditionalProperties.Ref != nil && s.AdditionalProperties.Value != nil && s.AdditionalProperties.Value.Type != ""
}

func (ap *AdditionalProperties) IsFreeForm() bool {
	if ap == nil || ap.Ref == nil {
		return false
	}
	if ap.Value == nil && ap.Ref == nil {
		return true
	}
	if ap.Value != nil && ap.Value.Type == "" {
		return true
	}
	return false
}
