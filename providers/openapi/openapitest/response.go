package openapitest

import (
	"mokapi/json/ref"
	"mokapi/media"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/parameter"
	"mokapi/providers/openapi/schema"
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
				Parameter: parameter.Parameter{
					Name:        name,
					Description: description,
					Schema:      &schema.Ref{Value: s},
				},
			},
		}
	}
}

func WithResponseHeaderRef(name string, ref *openapi.HeaderRef) ResponseOptions {
	return func(o *openapi.Response) {
		if o.Headers == nil {
			o.Headers = map[string]*openapi.HeaderRef{}
		}
		o.Headers[name] = ref
	}
}

func WithResponseDescription(description string) ResponseOptions {
	return func(o *openapi.Response) {
		o.Description = description
	}
}

func WithContent(mediaType string, content *openapi.MediaType) ResponseOptions {
	return func(o *openapi.Response) {
		ct := media.ParseContentType(mediaType)
		o.Content[mediaType] = content
		if content != nil {
			content.ContentType = ct
		}
	}
}

func NewContent(opts ...ContentOptions) *openapi.MediaType {
	mt := &openapi.MediaType{}
	for _, opt := range opts {
		opt(mt)
	}
	return mt
}

func WithEncoding(propName string, encoding *openapi.Encoding) ContentOptions {
	return func(c *openapi.MediaType) {
		if c.Encoding == nil {
			c.Encoding = map[string]*openapi.Encoding{}
		}

		c.Encoding[propName] = encoding
	}
}

func WithExample(example interface{}) ContentOptions {
	return func(c *openapi.MediaType) {
		c.Example = example
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
