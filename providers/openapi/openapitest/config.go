package openapitest

import (
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/parameter"
	"mokapi/providers/openapi/schema"
	"mokapi/version"
)

type ConfigOptions func(c *openapi.Config)

func NewConfig(versionString string, opts ...ConfigOptions) *openapi.Config {
	c := &openapi.Config{
		OpenApi: version.New(versionString),
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
	return WithComponentSchemaRef(name, &schema.Ref{Value: s})
}

func WithComponentSchemaRef(name string, s *schema.Ref) ConfigOptions {
	return func(c *openapi.Config) {
		if c.Components.Schemas == nil {
			c.Components.Schemas = &schema.Schemas{}
		}
		c.Components.Schemas.Set(name, s)
	}
}

func WithComponentResponse(name string, r *openapi.Response) ConfigOptions {
	return WithComponentResponseRef(name, &openapi.ResponseRef{Value: r})
}

func WithComponentResponseRef(name string, r *openapi.ResponseRef) ConfigOptions {
	return func(c *openapi.Config) {
		if c.Components.Responses == nil {
			c.Components.Responses = openapi.ResponseBodies{}
		}
		c.Components.Responses[name] = r
	}
}

func WithComponentRequestBody(name string, r *openapi.RequestBody) ConfigOptions {
	return WithComponentRequestBodyRef(name, &openapi.RequestBodyRef{Value: r})
}

func WithComponentRequestBodyRef(name string, r *openapi.RequestBodyRef) ConfigOptions {
	return func(c *openapi.Config) {
		if c.Components.RequestBodies == nil {
			c.Components.RequestBodies = openapi.RequestBodies{}
		}
		c.Components.RequestBodies[name] = r
	}
}

func WithComponentParameter(name string, p *parameter.Parameter) ConfigOptions {
	return WithComponentParameterRef(name, &parameter.Ref{Value: p})
}

func WithComponentParameterRef(name string, r *parameter.Ref) ConfigOptions {
	return func(c *openapi.Config) {
		if c.Components.Parameters == nil {
			c.Components.Parameters = openapi.Parameters{}
		}
		c.Components.Parameters[name] = r
	}
}

func WithComponentExample(name string, e *openapi.Example) ConfigOptions {
	return WithComponentExampleRef(name, &openapi.ExampleRef{Value: e})
}

func WithComponentExampleRef(name string, r *openapi.ExampleRef) ConfigOptions {
	return func(c *openapi.Config) {
		if c.Components.Examples == nil {
			c.Components.Examples = openapi.Examples{}
		}
		c.Components.Examples[name] = r
	}
}

func WithComponentHeader(name string, h *openapi.Header) ConfigOptions {
	return WithComponentHeaderRef(name, &openapi.HeaderRef{Value: h})
}

func WithComponentHeaderRef(name string, r *openapi.HeaderRef) ConfigOptions {
	return func(c *openapi.Config) {
		if c.Components.Headers == nil {
			c.Components.Headers = openapi.Headers{}
		}
		c.Components.Headers[name] = r
	}
}

func WithComponentPathItem(name string, r *openapi.Path) ConfigOptions {
	return WithComponentPathItemRef(name, &openapi.PathRef{Value: r})
}

func WithComponentPathItemRef(name string, r *openapi.PathRef) ConfigOptions {
	return func(c *openapi.Config) {
		if c.Components.PathItems == nil {
			c.Components.PathItems = openapi.PathItems{}
		}
		c.Components.PathItems[name] = r
	}
}
