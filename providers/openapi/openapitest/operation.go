package openapitest

import (
	"mokapi/media"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema"
	"strconv"
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

		o.Responses.Set(strconv.Itoa(status), &openapi.ResponseRef{Value: r})
	}
}

func WithResponseRef(status int, ref *openapi.ResponseRef) OperationOptions {
	return func(o *openapi.Operation) {
		o.Responses.Set(strconv.Itoa(status), ref)
	}
}

type ParamOptions func(p *openapi.Parameter)

func WithOperationParam(name string, required bool, opts ...ParamOptions) OperationOptions {
	return func(o *openapi.Operation) {

		o.Parameters = append(o.Parameters, &openapi.ParameterRef{
			Value: newParam(name, required, openapi.ParameterPath, opts...)})
	}
}

func WithOperationParamRef(ref *openapi.ParameterRef) OperationOptions {
	return func(o *openapi.Operation) {

		o.Parameters = append(o.Parameters, ref)
	}
}

func WithQueryParam(name string, required bool, opts ...ParamOptions) OperationOptions {
	return func(o *openapi.Operation) {

		o.Parameters = append(o.Parameters, &openapi.ParameterRef{
			Value: newParam(name, required, openapi.ParameterQuery, opts...)})
	}
}

func WithCookieParam(name string, required bool, opts ...ParamOptions) OperationOptions {
	return func(o *openapi.Operation) {
		o.Parameters = append(o.Parameters, &openapi.ParameterRef{
			Value: newParam(name, required, openapi.ParameterCookie, opts...)})
	}
}

func WithHeaderParam(name string, required bool, opts ...ParamOptions) OperationOptions {
	return func(o *openapi.Operation) {
		o.Parameters = append(o.Parameters, &openapi.ParameterRef{
			Value: newParam(name, required, openapi.ParameterHeader, opts...)})
	}
}

func WithParamInfo(description string) ParamOptions {
	return func(p *openapi.Parameter) {
		p.Description = description
	}
}

func WithRequestBody(description string, required bool, opts ...RequestBodyOptions) OperationOptions {
	return func(o *openapi.Operation) {
		body := &openapi.RequestBody{
			Description: description,
			Required:    required,
		}

		for _, opt := range opts {
			opt(body)
		}

		o.RequestBody = &openapi.RequestBodyRef{
			Value: body,
		}
	}
}

func newParam(name string, required bool, t openapi.Location, opts ...ParamOptions) *openapi.Parameter {
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

func WithOperationInfo(summary, description, operationId string, deprecated bool) OperationOptions {
	return func(o *openapi.Operation) {
		o.Summary = summary
		o.Description = description
		o.OperationId = operationId
		o.Deprecated = deprecated
	}
}

type RequestBodyOptions func(o *openapi.RequestBody)

func WithRequestContent(mediaType string, content *openapi.MediaType) RequestBodyOptions {
	return func(rb *openapi.RequestBody) {
		ct := media.ParseContentType(mediaType)
		if rb.Content == nil {
			rb.Content = map[string]*openapi.MediaType{}
		}
		rb.Content[mediaType] = content
		if content != nil {
			content.ContentType = ct
		}
	}
}

func WithParamSchema(s *schema.Schema) ParamOptions {
	return func(p *openapi.Parameter) {
		p.Schema = s
	}
}

func WithSecurity(s map[string][]string) OperationOptions {
	return func(o *openapi.Operation) {
		o.Security = append(o.Security, s)
	}
}

func WithGlobalSecurity(s map[string][]string) ConfigOptions {
	return func(c *openapi.Config) {
		c.Security = append(c.Security, s)
	}
}

func WithTagName(name string) OperationOptions {
	return func(o *openapi.Operation) {
		o.Tags = append(o.Tags, name)
	}
}
