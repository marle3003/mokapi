package openapitest

import "mokapi/config/dynamic/openapi"

type OperationOptions func(o *openapi.Operation)

func NewOperation(opts ...OperationOptions) *openapi.Operation {
	o := &openapi.Operation{Responses: make(map[openapi.HttpStatus]*openapi.ResponseRef)}
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

		o.Responses[status] = &openapi.ResponseRef{Value: r}
	}
}

func WithPathParam(name string, required bool) OperationOptions {
	return func(o *openapi.Operation) {
		o.Parameters = append(o.Parameters, &openapi.ParameterRef{
			Value: &openapi.Parameter{
				Name:     name,
				Type:     openapi.PathParameter,
				Required: required,
			}})
	}
}

func WithQueryParam(name string, required bool) OperationOptions {
	return func(o *openapi.Operation) {
		o.Parameters = append(o.Parameters, &openapi.ParameterRef{
			Value: &openapi.Parameter{
				Name:     name,
				Type:     openapi.QueryParameter,
				Required: required,
			}})
	}
}

func WithCookieParam(name string, required bool) OperationOptions {
	return func(o *openapi.Operation) {
		o.Parameters = append(o.Parameters, &openapi.ParameterRef{
			Value: &openapi.Parameter{
				Name:     name,
				Type:     openapi.CookieParameter,
				Required: required,
			}})
	}
}

func WithHeaderParam(name string, required bool) OperationOptions {
	return func(o *openapi.Operation) {
		o.Parameters = append(o.Parameters, &openapi.ParameterRef{
			Value: &openapi.Parameter{
				Name:     name,
				Type:     openapi.HeaderParameter,
				Required: required,
			}})
	}
}
