package parameter

import (
	"fmt"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/config/dynamic/openapi/schema"
)

const (
	Path   Location = "path"
	Query  Location = "query"
	Header Location = "header"
	Cookie Location = "cookie"
)

type Parameters []*Ref

type Ref struct {
	ref.Reference
	Value *Parameter
}

type Parameter struct {
	// The name of the parameter. Parameter names are case-sensitive.
	Name string `yaml:"name" json:"name"`

	// The location of the parameter
	Type Location `yaml:"in" json:"in"`

	// The schema defining the type used for the parameter
	Schema *schema.Ref `yaml:"schema" json:"schema"`

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
}

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
		case Query:
			p.Style = "form"
		case Path:
			p.Style = "simple"
		case Header:
			p.Style = "simple"
		case Cookie:
			p.Style = "form"
		}
	}
}

func (p Parameters) Parse(config *common.Config, reader common.Reader) error {
	for index, param := range p {
		if err := param.Parse(config, reader); err != nil {
			return fmt.Errorf("parse parameter index '%v' failed: %w", index, err)
		}
	}

	return nil
}

func (r *Ref) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 && r.Value == nil {
		return common.Resolve(r.Ref, &r.Value, config, reader)
	}

	if r.Value == nil {
		return nil
	}

	if err := r.Value.Schema.Parse(config, reader); err != nil {
		return err
	}

	return nil
}

func (p *Parameter) Parse(_ *common.Config, _ common.Reader) error {
	return nil
}
