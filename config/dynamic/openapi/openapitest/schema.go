package openapitest

import "mokapi/config/dynamic/openapi"

type SchemaOptions func(s *openapi.Schema)

func NewSchema(opts ...SchemaOptions) *openapi.Schema {
	s := new(openapi.Schema)
	for _, opt := range opts {
		opt(s)
	}
	return s
}
