package openapitest

import (
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/parameter"
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

func WithEndpointInfo(summary, description string) EndpointOptions {
	return func(o *openapi.Endpoint) {
		o.Summary = summary
		o.Description = description
	}
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

func WithEndpointParam(name string, in string, required bool, opts ...ParamOptions) EndpointOptions {
	return func(e *openapi.Endpoint) {
		e.Parameters = append(e.Parameters, &parameter.Ref{
			Value: newParam(name, required, parameter.Location(in), opts...)})
	}
}
