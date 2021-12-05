package openapitest

import "mokapi/config/dynamic/openapi"

type ResponseOptions func(o *openapi.Response)

type ContentOptions func(c *openapi.MediaType)

func NewResponse(opts ...ResponseOptions) *openapi.Response {
	r := &openapi.Response{Content: make(map[string]*openapi.MediaType)}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func WithContent(mime string, opts ...ContentOptions) ResponseOptions {
	return func(o *openapi.Response) {
		o.Content[mime] = &openapi.MediaType{}
		for _, opt := range opts {
			opt(o.Content[mime])
		}
	}
}

func WithSchema(s *openapi.Schema) ContentOptions {
	return func(c *openapi.MediaType) {
		c.Schema = &openapi.SchemaRef{Value: s}
	}
}
