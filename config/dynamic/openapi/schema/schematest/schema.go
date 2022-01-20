package schematest

import (
	"mokapi/config/dynamic/openapi/schema"
)

type SchemaOptions func(s *schema.Schema)

func New(typeName string, opts ...SchemaOptions) *schema.Schema {
	s := new(schema.Schema)
	s.Type = typeName
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithProperty(name string, ps *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		if s.Properties == nil {
			s.Properties = &schema.SchemasRef{}
		}
		if s.Properties.Value == nil {
			s.Properties.Value = &schema.Schemas{}
		}
		s.Properties.Value.Set(name, &schema.Ref{Value: ps})
	}
}

func WithItems(items *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		s.Items = &schema.Ref{Value: items}
	}
}

func WithRequired(names ...string) SchemaOptions {
	return func(s *schema.Schema) {
		for _, n := range names {
			s.Required = append(s.Required, n)
		}
	}
}

func WithUniqueItems() SchemaOptions {
	return func(s *schema.Schema) {
		s.UniqueItems = true
	}
}

func Any(schemas ...*schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		for _, any := range schemas {
			s.AnyOf = append(s.AnyOf, &schema.Ref{Value: any})
		}
	}
}

func OneOf(schemas ...*schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		for _, one := range schemas {
			s.OneOf = append(s.OneOf, &schema.Ref{Value: one})
		}
	}
}

func AllOf(schemas ...*schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		for _, all := range schemas {
			s.AllOf = append(s.AllOf, &schema.Ref{Value: all})
		}
	}
}

func WithFormat(format string) SchemaOptions {
	return func(s *schema.Schema) {
		s.Format = format
	}
}
