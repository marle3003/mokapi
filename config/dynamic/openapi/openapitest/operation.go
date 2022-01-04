package openapitest

import "mokapi/config/dynamic/openapi"

type OperationOptions func(o *openapi.Operation)

func NewOperation(opts ...OperationOptions) *openapi.Operation {
	o := &openapi.Operation{Responses: new(openapi.Responses)}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func WithResponse(status openapi.HttpStatus, opts ...ResponseOptions) OperationOptions {
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

type ParamOptions func(p *openapi.Parameter)

func WithPathParam(name string, required bool, opts ...ParamOptions) OperationOptions {
	return func(o *openapi.Operation) {

		o.Parameters = append(o.Parameters, &openapi.ParameterRef{
			Value: newParam(name, required, openapi.PathParameter, opts...)})
	}
}

func WithQueryParam(name string, required bool, opts ...ParamOptions) OperationOptions {
	return func(o *openapi.Operation) {

		o.Parameters = append(o.Parameters, &openapi.ParameterRef{
			Value: newParam(name, required, openapi.QueryParameter, opts...)})
	}
}

func WithCookieParam(name string, required bool, opts ...ParamOptions) OperationOptions {
	return func(o *openapi.Operation) {
		o.Parameters = append(o.Parameters, &openapi.ParameterRef{
			Value: newParam(name, required, openapi.CookieParameter, opts...)})
	}
}

func WithHeaderParam(name string, required bool, opts ...ParamOptions) OperationOptions {
	return func(o *openapi.Operation) {
		o.Parameters = append(o.Parameters, &openapi.ParameterRef{
			Value: newParam(name, required, openapi.HeaderParameter, opts...)})
	}
}

func newParam(name string, required bool, t openapi.ParameterLocation, opts ...ParamOptions) *openapi.Parameter {
	p := &openapi.Parameter{
		Name:     name,
		Type:     t,
		Required: required,
	}

	for _, opt := range opts {
		opt(p)
	}
	return p
}

func WithParamSchema(s *openapi.Schema) ParamOptions {
	return func(p *openapi.Parameter) {
		p.Schema = &openapi.SchemaRef{Value: s}
	}
}
