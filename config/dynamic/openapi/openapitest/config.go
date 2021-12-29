package openapitest

import "mokapi/config/dynamic/openapi"

type ConfigOptions func(c *openapi.Config)

func NewConfig(version string, opts ...ConfigOptions) *openapi.Config {
	c := &openapi.Config{
		OpenApi:   version,
		Servers:   nil,
		EndPoints: make(map[string]*openapi.EndpointRef),
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
		c.EndPoints[path] = &openapi.EndpointRef{Value: endpoint}
	}
}

func WithEndpointRef(path string, endpoint *openapi.EndpointRef) ConfigOptions {
	return func(c *openapi.Config) {
		c.EndPoints[path] = endpoint
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
