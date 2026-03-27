package openapi

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/providers/openapi/schema"
)

type Components struct {
	Schemas         *schema.Schemas     `yaml:"schemas,omitempty" json:"schemas,omitempty"`
	Responses       ResponseBodies      `yaml:"responses,omitempty" json:"responses,omitempty"`
	RequestBodies   RequestBodies       `yaml:"requestBodies,omitempty" json:"requestBodies,omitempty"`
	Parameters      ComponentParameters `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Examples        Examples            `yaml:"examples,omitempty" json:"examples,omitempty"`
	Headers         Headers             `yaml:"headers,omitempty" json:"headers,omitempty"`
	PathItems       PathItems           `yaml:"pathItems,omitempty" json:"pathItems,omitempty"`
	SecuritySchemes SecuritySchemes     `yaml:"securitySchemes,omitempty" json:"securitySchemes,omitempty"`
}

type ComponentParameters map[string]*ParameterRef

func (p ComponentParameters) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	for name, param := range p {
		if err := param.Parse(config, reader); err != nil {
			return fmt.Errorf("parse parameter '%v' failed: %w", name, err)
		}
	}
	return nil
}

func (c *Components) patch(patch Components) {
	if c.Schemas == nil {
		c.Schemas = patch.Schemas
	} else {
		c.Schemas.Patch(patch.Schemas)
	}
	if c.Responses == nil {
		c.Responses = patch.Responses
	} else {
		c.Responses.patch(patch.Responses)
	}
	if c.RequestBodies == nil {
		c.RequestBodies = patch.RequestBodies
	} else {
		c.RequestBodies.patch(patch.RequestBodies)
	}
	if c.Parameters == nil {
		c.Parameters = patch.Parameters
	} else {
		c.Parameters.patch(patch.Parameters)
	}
	if c.Examples == nil {
		c.Examples = patch.Examples
	} else {
		c.Examples.patch(patch.Examples)
	}
	if c.Headers == nil {
		c.Headers = patch.Headers
	} else {
		c.Headers.patch(patch.Headers)
	}
	if c.SecuritySchemes == nil {
		c.SecuritySchemes = patch.SecuritySchemes
	} else {
		c.SecuritySchemes.patch(patch.SecuritySchemes)
	}
}

func (p ComponentParameters) patch(patch ComponentParameters) {
	for name, param := range patch {
		if p1, ok := p[name]; ok {
			p1.Patch(param)
		} else {
			p[name] = param
		}
	}
}
