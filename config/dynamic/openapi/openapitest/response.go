package openapitest

import (
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/media"
)

type ResponseOptions func(o *openapi.Response)

type ContentOptions func(c *openapi.MediaType)

func NewResponse(opts ...ResponseOptions) *openapi.Response {
	r := &openapi.Response{Content: make(map[string]*openapi.MediaType)}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func WithResponseHeader(name, description string, s *schema.Schema) ResponseOptions {
	return func(o *openapi.Response) {
		if o.Headers == nil {
			o.Headers = map[string]*openapi.HeaderRef{}
		}
		o.Headers[name] = &openapi.HeaderRef{
			Value: &openapi.Header{
				Name:        name,
				Description: description,
				Schema:      &schema.Ref{Value: s},
			},
		}
	}
}

func WithResponseDescription(description string) ResponseOptions {
	return func(o *openapi.Response) {
		o.Description = description
	}
}

func WithContent(mediaType string, opts ...ContentOptions) ResponseOptions {
	return func(o *openapi.Response) {
		ct := media.ParseContentType(mediaType)
		o.Content[mediaType] = &openapi.MediaType{ContentType: ct}
		for _, opt := range opts {
			opt(o.Content[mediaType])
		}
	}
}

func WithSchema(s *schema.Schema) ContentOptions {
	return func(c *openapi.MediaType) {
		c.Schema = &schema.Ref{Value: s}
	}
}

func WithSchemaRef(r string) ContentOptions {
	return func(c *openapi.MediaType) {
		c.Schema = &schema.Ref{Reference: ref.Reference{Ref: r}}
	}
}
