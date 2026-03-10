package openapitest

import (
	"mokapi/config/dynamic"
	"mokapi/media"
	"mokapi/providers/openapi"
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
				Parameter: openapi.Parameter{
					Name:        name,
					Description: description,
					Schema:      s,
				},
			},
		}
	}
}

func WithResponseHeaderRef(name string, ref string) ResponseOptions {
	return func(o *openapi.Response) {
		if o.Headers == nil {
			o.Headers = map[string]*openapi.HeaderRef{}
		}
		o.Headers[name] = &openapi.HeaderRef{Reference: dynamic.Reference{Ref: ref}}
	}
}

func UseResponseHeaderRef(name string, ref *openapi.HeaderRef) ResponseOptions {
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

func WithContent(mediaType string, opts ...ContentOptions) ResponseOptions {
	return func(o *openapi.Response) {
		ct := media.ParseContentType(mediaType)
		content := NewContent(opts...)
		o.Content[mediaType] = content
		if content != nil {
			content.ContentType = ct
		}
	}
}

func UseContent(mediaType string, mt *openapi.MediaType) ResponseOptions {
	return func(o *openapi.Response) {
		if o.Content == nil {
			o.Content = map[string]*openapi.MediaType{}
		}
		o.Content[mediaType] = mt
		if mt != nil {
			mt.ContentType = media.ParseContentType(mediaType)
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

func WithExampleValue(example interface{}) ContentOptions {
	return func(c *openapi.MediaType) {
		c.Example = &openapi.ExampleValue{Value: example}
	}
}

func WithExampleRef(name, ref string) ContentOptions {
	return func(c *openapi.MediaType) {
		if c.Examples == nil {
			c.Examples = map[string]*openapi.ExampleRef{}
		}
		c.Examples[name] = &openapi.ExampleRef{Reference: dynamic.Reference{Ref: ref}}
	}
}

func WithSchema(s *schema.Schema) ContentOptions {
	return func(c *openapi.MediaType) {
		c.Schema = s
	}
}

func WithSchemaRef(r string) ContentOptions {
	return func(c *openapi.MediaType) {
		c.Schema = &schema.Schema{Ref: r}
	}
}
