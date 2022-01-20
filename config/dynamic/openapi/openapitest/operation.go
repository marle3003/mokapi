package openapitest

import (
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/parameter"
	"mokapi/config/dynamic/openapi/schema"
)

type OperationOptions func(o *openapi.Operation)

func NewOperation(opts ...OperationOptions) *openapi.Operation {
	o := &openapi.Operation{Responses: new(openapi.Responses)}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func WithResponse(status int, opts ...ResponseOptions) OperationOptions {
	return func(o *openapi.Operation) {
		r := &openapi.Response{
			Content: make(map[string]*openapi.MediaType),
			Headers: make(map[string]*openapi.HeaderRef)}
		for _, opt := range opts {
			opt(r)
		}

		o.Responses.Set(status, &openapi.ResponseRef{Value: r})
	}
}

type ParamOptions func(p *parameter.Parameter)

func WithPathParam(name string, required bool, opts ...ParamOptions) OperationOptions {
	return func(o *openapi.Operation) {

		o.Parameters = append(o.Parameters, &parameter.Ref{
			Value: newParam(name, required, parameter.Path, opts...)})
	}
}

func WithQueryParam(name string, required bool, opts ...ParamOptions) OperationOptions {
	return func(o *openapi.Operation) {

		o.Parameters = append(o.Parameters, &parameter.Ref{
			Value: newParam(name, required, parameter.Query, opts...)})
	}
}

func WithCookieParam(name string, required bool, opts ...ParamOptions) OperationOptions {
	return func(o *openapi.Operation) {
		o.Parameters = append(o.Parameters, &parameter.Ref{
			Value: newParam(name, required, parameter.Cookie, opts...)})
	}
}

func WithHeaderParam(name string, required bool, opts ...ParamOptions) OperationOptions {
	return func(o *openapi.Operation) {
		o.Parameters = append(o.Parameters, &parameter.Ref{
			Value: newParam(name, required, parameter.Header, opts...)})
	}
}

func newParam(name string, required bool, t parameter.Location, opts ...ParamOptions) *parameter.Parameter {
	p := &parameter.Parameter{
		Name:     name,
		Type:     t,
		Required: required,
	}

	for _, opt := range opts {
		opt(p)
	}
	return p
}

func WithParamSchema(s *schema.Schema) ParamOptions {
	return func(p *parameter.Parameter) {
		p.Schema = &schema.Ref{Value: s}
	}
}
