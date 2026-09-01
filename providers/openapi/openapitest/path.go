package openapitest

import (
	"mokapi/config/dynamic"
	"mokapi/providers/openapi"
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
	e := &openapi.Path{Path: path}
	for _, opt := range opts {
		opt(e)
	}
	if config.Paths == nil {
		config.Paths = &openapi.PathItems{}
	}
	config.Paths.Set(path, &openapi.PathRef{
		Value: e,
	})
	return e
}

func UseOperation(method string, op *openapi.Operation) PathOptions {
	op.Method = strings.ToUpper(method)
	return func(e *openapi.Path) {
		switch op.Method {
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
		case "QUERY":
			e.Query = op
		default:
			if e.AdditionalOperations == nil {
				e.AdditionalOperations = make(map[string]*openapi.Operation)
			}
			e.AdditionalOperations[method] = op
		}
		op.Path = e
		e.MethodOrder = append(e.MethodOrder, op.Method)
	}
}

func WithOperation(method string, opts ...OperationOptions) PathOptions {
	op := NewOperation(opts...)
	return UseOperation(method, op)
}

func WithPathParam(name string, opts ...ParamOptions) PathOptions {
	return func(e *openapi.Path) {
		e.Parameters = append(e.Parameters, &openapi.ParameterRef{
			Value: newParam(name, true, "path", opts...)})
	}
}

func WithPathParamRef(ref string) PathOptions {
	return func(e *openapi.Path) {
		e.Parameters = append(e.Parameters, &openapi.ParameterRef{
			Reference: dynamic.Reference[*openapi.ParameterRef]{Ref: ref},
		})
	}
}

func WithPathErrors(err ...openapi.Error) PathOptions {
	return func(p *openapi.Path) {
		p.Status = openapi.StatusInvalid
		p.Errors = err
	}
}

func WithPathSummary(summary string) PathOptions {
	return func(p *openapi.Path) {
		p.Summary = summary
	}
}

func WithPathDescription(description string) PathOptions {
	return func(p *openapi.Path) {
		p.Description = description
	}
}
