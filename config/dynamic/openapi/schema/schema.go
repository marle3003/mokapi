package schema

import (
	"fmt"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/sortedmap"
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
	AdditionalProperties *Ref   // TODO custom marshal for bool, {} etc. Should it be a schema reference?
	Faker                string `yaml:"x-faker" json:"x-faker"`
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
}

type AdditionalProperties struct {
	Schema *Schema
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
