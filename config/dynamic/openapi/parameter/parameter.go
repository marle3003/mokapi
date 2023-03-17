package parameter

import (
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

type NamedParameters struct {
	ref.Reference
	Value map[string]*Ref
}

func (l Location) String() string {
	return string(l)
}
