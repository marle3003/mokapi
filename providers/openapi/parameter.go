package openapi

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/providers/openapi/schema"
)

const (
	ParameterPath        Location = "path"
	ParameterQuery       Location = "query"
	ParameterHeader      Location = "header"
	ParameterCookie      Location = "cookie"
	ParameterQueryString Location = "querystring"
)

type ParameterRef struct {
	dynamic.Reference
	Value *Parameter
}

type Parameter struct {
	// The name of the parameter. Parameter names are case-sensitive.
	Name string `yaml:"name" json:"name"`

	// The location of the parameter
	Type Location `yaml:"in" json:"in"`

	// The schema defining the type used for the parameter
	Schema *schema.Schema `yaml:"schema" json:"schema"`

	// Determines whether the parameter is mandatory.
	// If the location of the parameter is "path", this property
	// is required and its value MUST be true
	Required bool `yaml:"required" json:"required"`

	// A brief description of the parameter. This could contain examples
	// of use.
	Description string `yaml:"description" json:"description"`

	Deprecated bool `yaml:"deprecated" json:"deprecated"`

	// Defines how multiple values are delimited. Possible styles depend on
	// the parameter location
	Style string `yaml:"style" json:"style"`

	// specifies whether arrays and objects should generate separate
	// parameters for each array item or object property
	Explode *bool `yaml:"explode" json:"explode"`

	AllowReserved bool `yaml:"allowReserved" json:"allowReserved"`

	Content Content `yaml:"content" json:"content"`
}

type Parameters []*ParameterRef

type Location string

func (l Location) String() string {
	return string(l)
}

func (p *Parameter) IsExplode() bool {
	if p.Explode != nil {
		return *p.Explode
	}
	if p.Style == "form" {
		return true
	}
	return false
}

func (p *Parameter) SetDefaultStyle() {
	if p.Style == "" {
		switch p.Type {
		case ParameterQuery:
			p.Style = "form"
		case ParameterPath:
			p.Style = "simple"
		case ParameterHeader:
			p.Style = "simple"
		case ParameterCookie:
			p.Style = "form"
		}
	}
}

func (p *Parameters) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	for index, param := range *p {
		if err := param.Parse(config, reader); err != nil {
			return fmt.Errorf("parse parameter index '%v' failed: %w", index, err)
		}
	}

	return nil
}

func (r *ParameterRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 && r.Value == nil {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
	}

	if r.Value == nil {
		return nil
	}

	if err := r.Value.Schema.Parse(config, reader); err != nil {
		return fmt.Errorf("parse schema failed: %w", err)
	}

	return nil
}

func (p *Parameter) Parse(_ *dynamic.Config, _ dynamic.Reader) error {
	return nil
}
