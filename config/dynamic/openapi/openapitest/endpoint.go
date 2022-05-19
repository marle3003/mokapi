package openapitest

import (
	"mokapi/config/dynamic/openapi"
	"strings"
)

type EndpointOptions func(o *openapi.Endpoint)

func NewEndpoint(opts ...EndpointOptions) *openapi.Endpoint {
	e := &openapi.Endpoint{}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func AppendEndpoint(path string, config *openapi.Config, opts ...EndpointOptions) *openapi.Endpoint {
	e := &openapi.Endpoint{}
	for _, opt := range opts {
		opt(e)
	}
	if config.Paths.Value == nil {
		config.Paths.Value = make(map[string]*openapi.EndpointRef)
	}
	config.Paths.Value[path] = &openapi.EndpointRef{
		Value: e,
	}
	return e
}

func WithOperation(method string, op *openapi.Operation) EndpointOptions {
	return func(e *openapi.Endpoint) {
		switch strings.ToUpper(method) {
		case "GET":
			e.Get = op
		case "POST":
			e.Post = op
		case "PUT":
			e.Put = op
		case "PATCH":
			e.Patch = op
		case "DELETE":
			e.Delete = op
		case "HEAD":
			e.Head = op
		case "OPTIONS":
			e.Options = op
		case "TRACE":
			e.Trace = op
		}
	}
}
