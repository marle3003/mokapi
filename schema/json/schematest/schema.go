package schematest

import (
	"mokapi/schema/json/schema"
)

type SchemaOptions func(s *schema.Schema)

func New(typeName string, opts ...SchemaOptions) *schema.Schema {
	s := new(schema.Schema)
	s.Type = append(s.Type, typeName)
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func NewTypes(typeNames []string, opts ...SchemaOptions) *schema.Schema {
	s := new(schema.Schema)
	s.Type = append(s.Type, typeNames...)
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func NewRef(typeName string, opts ...SchemaOptions) *schema.Ref {
	s := new(schema.Schema)
	s.Type = append(s.Type, typeName)
	for _, opt := range opts {
		opt(s)
	}
	return &schema.Ref{Value: s}
}

func WithProperty(name string, ps *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		if s.Properties == nil {
			s.Properties = &schema.Schemas{}
		}
		s.Properties.Set(name, &schema.Ref{Value: ps})
	}
}

func WithItems(typeName string, opts ...SchemaOptions) SchemaOptions {
	return func(s *schema.Schema) {
		s.Items = NewRef(typeName, opts...)
	}
}

func WithItemsRef(ref *schema.Ref) SchemaOptions {
	return func(s *schema.Schema) {
		s.Items = ref
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

func NewAny(schemas ...*schema.Schema) *schema.Schema {
	s := &schema.Schema{}
	for _, any := range schemas {
		s.AnyOf = append(s.AnyOf, &schema.Ref{Value: any})
	}
	return s
}

func NewAnyRef(schemas ...*schema.Ref) *schema.Schema {
	s := &schema.Schema{}
	for _, any := range schemas {
		s.AnyOf = append(s.AnyOf, any)
	}
	return s
}

func OneOf(schemas ...*schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		for _, one := range schemas {
			s.OneOf = append(s.OneOf, &schema.Ref{Value: one})
		}
	}
}

func NewOneOf(schemas ...*schema.Schema) *schema.Schema {
	s := &schema.Schema{}
	for _, one := range schemas {
		s.OneOf = append(s.OneOf, &schema.Ref{Value: one})
	}
	return s
}

func NewOneOfRef(schemas ...*schema.Ref) *schema.Schema {
	s := &schema.Schema{}
	for _, one := range schemas {
		s.OneOf = append(s.OneOf, one)
	}
	return s
}

func AllOf(schemas ...*schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		for _, all := range schemas {
			s.AllOf = append(s.AllOf, &schema.Ref{Value: all})
		}
	}
}

func NewAllOf(schemas ...*schema.Schema) *schema.Schema {
	s := &schema.Schema{}
	for _, all := range schemas {
		s.AllOf = append(s.AllOf, &schema.Ref{Value: all})
	}
	return s
}

func NewAllOfRefs(schemas ...*schema.Ref) *schema.Schema {
	s := &schema.Schema{}
	for _, all := range schemas {
		s.AllOf = append(s.AllOf, all)
	}
	return s
}

func WithFormat(format string) SchemaOptions {
	return func(s *schema.Schema) {
		s.Format = format
	}
}

func WithMinItems(n int) SchemaOptions {
	return func(s *schema.Schema) {
		s.MinItems = &n
	}
}

func WithMaxItems(n int) SchemaOptions {
	return func(s *schema.Schema) {
		s.MaxItems = &n
	}
}

func WithMinProperties(n int) SchemaOptions {
	return func(s *schema.Schema) {
		s.MinProperties = &n
	}
}

func WithMaxProperties(n int) SchemaOptions {
	return func(s *schema.Schema) {
		s.MaxProperties = &n
	}
}

func WithAdditionalProperties(additional *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		s.AdditionalProperties = schema.AdditionalProperties{Ref: &schema.Ref{Value: additional}}
	}
}

func WithFreeForm(allowed bool) SchemaOptions {
	return func(s *schema.Schema) {
		if allowed {
			s.AdditionalProperties = schema.AdditionalProperties{}
		} else {
			s.AdditionalProperties = schema.AdditionalProperties{Forbidden: true}
		}
	}
}

func WithMinimum(min float64) SchemaOptions {
	return func(s *schema.Schema) {
		s.Minimum = &min
	}
}

func WithMaximum(max float64) SchemaOptions {
	return func(s *schema.Schema) {
		s.Maximum = &max
	}
}

func WithExclusiveMinimum(min float64) SchemaOptions {
	return func(s *schema.Schema) {
		s.ExclusiveMinimum = schema.NewUnionTypeA[float64, bool](min)
	}
}

func WithExclusiveMaximum(max float64) SchemaOptions {
	return func(s *schema.Schema) {
		s.ExclusiveMaximum = schema.NewUnionTypeA[float64, bool](max)
	}
}

func WithMultipleOf(n float64) SchemaOptions {
	return func(s *schema.Schema) {
		s.MultipleOf = &n
	}
}

func WithPattern(p string) SchemaOptions {
	return func(s *schema.Schema) {
		s.Pattern = p
	}
}

func WithEnum(e []interface{}) SchemaOptions {
	return func(s *schema.Schema) {
		s.Enum = e
	}
}

func WithMinLength(n int) SchemaOptions {
	return func(s *schema.Schema) {
		s.MinLength = &n
	}
}

func WithMaxLength(n int) SchemaOptions {
	return func(s *schema.Schema) {
		s.MaxLength = &n
	}
}

func IsNullable(b bool) SchemaOptions {
	return func(s *schema.Schema) {
		s.Type = append(s.Type, "null")
	}
}

func WithDefault(v interface{}) SchemaOptions {
	return func(s *schema.Schema) {
		s.Default = v
	}
}
