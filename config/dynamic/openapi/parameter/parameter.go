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
	Name string

	// The location of the parameter
	Type Location `yaml:"in" json:"in"`

	// The schema defining the type used for the parameter
	Schema *schema.Ref

	// Determines whether the parameter is mandatory.
	// If the location of the parameter is "path", this property
	// is required and its value MUST be true
	Required bool

	// A brief description of the parameter. This could contain examples
	// of use.
	Description string

	Deprecated bool `yaml:"deprecated" json:"deprecated"`

	// Defines how multiple values are delimited. Possible styles depend on
	// the parameter location
	Style string

	// specifies whether arrays and objects should generate separate
	// parameters for each array item or object property
	Explode bool
}

type Location string

func (l Location) String() string {
	return string(l)
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
		if err := common.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
	}

	if r.Value == nil {
		return nil
	}

	if err := r.Value.Schema.Parse(config, reader); err != nil {
		return err
	}

	return nil
}

func (p *Parameter) Parse(config *common.Config, reader common.Reader) error {
	return nil
}
