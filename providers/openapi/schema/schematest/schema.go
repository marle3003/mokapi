package schematest

import (
	"mokapi/config/dynamic"
	"mokapi/providers/openapi/schema"
	jsonSchema "mokapi/schema/json/schema"
)

type SchemaOptions func(s *schema.Schema)

func New(typeName string, opts ...SchemaOptions) *schema.Schema {
	s := new(schema.Schema)
	s.Type = jsonSchema.Types{typeName}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func NewRef(typeName string, opts ...SchemaOptions) *schema.Ref {
	s := new(schema.Schema)
	s.Type = jsonSchema.Types{typeName}
	for _, opt := range opts {
		opt(s)
	}
	return &schema.Ref{Value: s}
}

func And(typeName string) SchemaOptions {
	return func(s *schema.Schema) {
		s.Type = append(s.Type, typeName)
	}
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

func WithItemsRef(ref string) SchemaOptions {
	return func(s *schema.Schema) {
		s.Items = &schema.Ref{Reference: dynamic.Reference{Ref: ref}}
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

func WithShuffleItems() SchemaOptions {
	return func(s *schema.Schema) {
		s.ShuffleItems = true
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
		s.AdditionalProperties = &schema.Ref{Value: additional}
	}
}

func WithFreeForm(allowed bool) SchemaOptions {
	return func(s *schema.Schema) {
		s.AdditionalProperties = &schema.Ref{Boolean: &allowed}
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
		if s.ExclusiveMinimum == nil {
			s.ExclusiveMinimum = jsonSchema.NewUnionTypeA[float64, bool](min)
		} else {
			s.ExclusiveMinimum.A = min
		}
	}
}

func WithExclusiveMaximum(max float64) SchemaOptions {
	return func(s *schema.Schema) {
		if s.ExclusiveMaximum == nil {
			s.ExclusiveMaximum = jsonSchema.NewUnionTypeA[float64, bool](max)
		} else {
			s.ExclusiveMaximum.A = max
		}
	}
}

func WithXml(xml *schema.Xml) SchemaOptions {
	return func(s *schema.Schema) {
		s.Xml = xml
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

func WithMultipleOf(f float64) SchemaOptions {
	return func(s *schema.Schema) {
		s.MultipleOf = &f
	}
}

func IsNullable(b bool) SchemaOptions {
	return func(s *schema.Schema) {
		s.Nullable = b
	}
}

func WithDefault(d interface{}) SchemaOptions {
	return func(s *schema.Schema) {
		s.Default = d
	}
}

func WithSchema(value string) SchemaOptions {
	return func(s *schema.Schema) {
		s.Schema = value
	}
}

func WithTitle(title string) SchemaOptions {
	return func(s *schema.Schema) {
		s.Title = title
	}
}

func WithDescription(description string) SchemaOptions {
	return func(s *schema.Schema) {
		s.Description = description
	}
}

func WithConst(c interface{}) SchemaOptions {
	return func(s *schema.Schema) {
		s.Const = c
	}
}

func WithDeprecated(b bool) SchemaOptions {
	return func(s *schema.Schema) {
		s.Deprecated = b
	}
}

func WithExample(e interface{}) SchemaOptions {
	return func(s *schema.Schema) {
		s.Example = e
	}
}

func WithExamples(e ...interface{}) SchemaOptions {
	return func(s *schema.Schema) {
		s.Examples = e
	}
}

func WithContentMediaType(value string) SchemaOptions {
	return func(s *schema.Schema) {
		s.ContentMediaType = value
	}
}

func WithContentEncoding(value string) SchemaOptions {
	return func(s *schema.Schema) {
		s.ContentEncoding = value
	}
}

func WithExclusiveMinimumFlag(b bool) SchemaOptions {
	return func(s *schema.Schema) {
		if s.ExclusiveMinimum != nil {
			s.ExclusiveMinimum.B = b
		} else {
			s.ExclusiveMinimum = jsonSchema.NewUnionTypeB[float64, bool](b)
		}
	}
}
