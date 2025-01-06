package openapitest

import (
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/parameter"
	"strings"
)

type PathOptions func(o *openapi.Path)

func NewPath(opts ...PathOptions) *openapi.Path {
	e := &openapi.Path{}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func WithPathInfo(summary, description string) PathOptions {
	return func(o *openapi.Path) {
		o.Summary = summary
		o.Description = description
	}
}

func AppendPath(path string, config *openapi.Config, opts ...PathOptions) *openapi.Path {
	e := &openapi.Path{}
	for _, opt := range opts {
		opt(e)
	}
	if config.Paths == nil {
		config.Paths = make(map[string]*openapi.PathRef)
	}
	config.Paths[path] = &openapi.PathRef{
		Value: e,
	}
	return e
}

func WithOperation(method string, op *openapi.Operation) PathOptions {
	return func(e *openapi.Path) {
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

func WithPathParam(name string, opts ...ParamOptions) PathOptions {
	return func(e *openapi.Path) {
		e.Parameters = append(e.Parameters, &parameter.Ref{
			Value: newParam(name, true, "path", opts...)})
	}
}

func WithPathParamRef(ref *parameter.Ref) PathOptions {
	return func(e *openapi.Path) {
		e.Parameters = append(e.Parameters, ref)
	}
}
