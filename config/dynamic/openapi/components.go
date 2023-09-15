package openapi

import (
	"mokapi/config/dynamic/openapi/parameter"
	"mokapi/config/dynamic/openapi/schema"
)

type Components struct {
	Schemas       *schema.Schemas      `yaml:"schemas,omitempty" json:"schemas,omitempty"`
	Responses     *Responses           `yaml:"responses,omitempty" json:"responses,omitempty"`
	RequestBodies RequestBodies        `yaml:"requestBodies,omitempty" json:"requestBodies,omitempty"`
	Parameters    parameter.Parameters `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Examples      Examples             `yaml:"examples,omitempty" json:"examples,omitempty"`
	Headers       Headers              `yaml:"headers,omitempty" json:"headers,omitempty"`
}

type RequestBodies map[string]*RequestBodyRef
