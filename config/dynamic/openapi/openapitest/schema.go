package openapitest

import "mokapi/config/dynamic/openapi"

type SchemaOptions func(s *openapi.Schema)

func NewSchema(typeName string, opts ...SchemaOptions) *openapi.Schema {
	s := new(openapi.Schema)
	s.Type = typeName
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithProperty(name string, schema *openapi.Schema) SchemaOptions {
	return func(s *openapi.Schema) {
		if s.Properties == nil {
			s.Properties = &openapi.SchemasRef{}
		}
		if s.Properties.Value == nil {
			s.Properties.Value = &openapi.Schemas{}
		}
		s.Properties.Value.Set(name, &openapi.SchemaRef{Value: schema})
	}
}

func WithItems(schema *openapi.Schema) SchemaOptions {
	return func(s *openapi.Schema) {
		s.Items = &openapi.SchemaRef{Value: schema}
	}
}

func WithRequired(names ...string) SchemaOptions {
	return func(s *openapi.Schema) {
		for _, n := range names {
			s.Required = append(s.Required, n)
		}
	}
}

func WithUniqueItems() SchemaOptions {
	return func(s *openapi.Schema) {
		s.UniqueItems = true
	}
}

func Any(schemas ...*openapi.Schema) SchemaOptions {
	return func(s *openapi.Schema) {
		for _, any := range schemas {
			s.AnyOf = append(s.AnyOf, &openapi.SchemaRef{Value: any})
		}
	}
}

func OneOf(schemas ...*openapi.Schema) SchemaOptions {
	return func(s *openapi.Schema) {
		for _, one := range schemas {
			s.OneOf = append(s.OneOf, &openapi.SchemaRef{Value: one})
		}
	}
}

func AllOf(schemas ...*openapi.Schema) SchemaOptions {
	return func(s *openapi.Schema) {
		for _, all := range schemas {
			s.AllOf = append(s.AllOf, &openapi.SchemaRef{Value: all})
		}
	}
}

func WithFormat(format string) SchemaOptions {
	return func(s *openapi.Schema) {
		s.Format = format
	}
}
