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
		Paths:   make(map[string]*openapi.PathRef),
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

func WithContact(name, url, email string) ConfigOptions {
	return func(c *openapi.Config) {
		c.Info.Contact = &openapi.Contact{
			Name:  name,
			Url:   url,
			Email: email,
		}
	}
}

func WithPath(name string, path *openapi.Path) ConfigOptions {
	return func(c *openapi.Config) {
		c.Paths[name] = &openapi.PathRef{Value: path}
	}
}

func WithPathRef(name string, ref *openapi.PathRef) ConfigOptions {
	return func(c *openapi.Config) {
		c.Paths[name] = ref
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

func WithComponentSchema(name string, s *schema.Schema) ConfigOptions {
	return func(c *openapi.Config) {
		if c.Components.Schemas == nil {
			c.Components.Schemas = &schema.SchemasRef{Value: &schema.Schemas{}}
		}
		c.Components.Schemas.Value.Set(name, &schema.Ref{Value: s})
	}
}

func WithComponentSchemaRef(name string, s *schema.Ref) ConfigOptions {
	return func(c *openapi.Config) {
		if c.Components.Schemas == nil {
			c.Components.Schemas = &schema.SchemasRef{Value: &schema.Schemas{}}
		}
		c.Components.Schemas.Value.Set(name, s)
	}
}
