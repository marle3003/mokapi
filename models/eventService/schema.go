package event

import (
	"fmt"
	"mokapi/config/dynamic/asyncApi"
	"strings"
)

func buildSchemaFromComponents(name string, config *asyncApi.Schema, ctx *context) (schema *Schema) {
	if s, exists := ctx.service.Models[name]; exists && s.isResolved {
		return s
	} else if exists {
		schema = s
		s = newSchema(config)
		*schema = *s
		schema.Name = name
	} else {
		schema = newSchema(config)
		schema.Name = name
		ctx.service.Models[name] = schema
	}

	schema.Reference = "#/components/schemas/" + name

	for n, p := range config.Properties {
		if schema.Properties == nil {
			schema.Properties = make(map[string]*Schema)
		}
		schema.Properties[n] = createSchema(p, ctx)
	}

	if config.Items != nil {
		schema.Items = createSchema(config.Items, ctx)
	}

	return
}

func createSchema(config *asyncApi.Schema, ctx *context) *Schema {
	if config == nil {
		return nil
	}

	if len(config.Reference) > 0 {
		switch r := config.Reference; {
		case strings.HasPrefix(r, "#/components/schema"):
			seg := strings.Split(r, "/")
			name := seg[len(seg)-1]
			if s, exists := ctx.service.Models[name]; exists {
				return s
			} else {
				s := &Schema{Reference: r, isResolved: false}
				ctx.service.Models[name] = s
				return s
			}
		default:
			ctx.error(fmt.Sprintf("$ref '%v' is not supported", r))
			return nil
		}
	}

	s := newSchema(config)

	if config.Items != nil {
		s.Items = createSchema(config.Items, ctx)
	}

	for n, p := range config.Properties {
		if s.Properties == nil {
			s.Properties = make(map[string]*Schema)
		}
		s.Properties[n] = createSchema(p, ctx)
	}

	return s
}

func newSchema(config *asyncApi.Schema) *Schema {
	schema := &Schema{
		Description: config.Description,
		Type:        config.Type,
		Reference:   config.Reference,
	}

	return schema
}
