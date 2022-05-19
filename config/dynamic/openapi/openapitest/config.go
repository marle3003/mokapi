package openapitest

import (
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/schema"
)

type ConfigOptions func(c *openapi.Config)

func NewConfig(version string, opts ...ConfigOptions) *openapi.Config {
	c := &openapi.Config{
		OpenApi: version,
		Servers: nil,
		Paths:   openapi.EndpointsRef{Value: make(map[string]*openapi.EndpointRef)},
	}

	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithInfo(name, version, description string) ConfigOptions {
	return func(c *openapi.Config) {
		c.Info.Name = name
		c.Info.Version = version
		c.Info.Description = description
	}
}

func WithEndpoint(path string, endpoint *openapi.Endpoint) ConfigOptions {
	return func(c *openapi.Config) {
		c.Paths.Value[path] = &openapi.EndpointRef{Value: endpoint}
	}
}

func WithEndpointsRef(ref string) ConfigOptions {
	return func(c *openapi.Config) {
		c.Paths.Reference.Ref = ref
	}
}

func WithEndpointRef(path string, endpoint *openapi.EndpointRef) ConfigOptions {
	return func(c *openapi.Config) {
		c.Paths.Value[path] = endpoint
	}
}

func WithServer(url, description string) ConfigOptions {
	return func(c *openapi.Config) {
		c.Servers = append(c.Servers, &openapi.Server{
			Url:         url,
			Description: description,
		})
	}
}

func WithComponentSchema(name string, ref *schema.Ref) ConfigOptions {
	return func(c *openapi.Config) {
		if c.Components.Schemas == nil {
			c.Components.Schemas = &schema.SchemasRef{
				Value: &schema.Schemas{},
			}
		}
		c.Components.Schemas.Value.Set(name, ref)
	}
}
