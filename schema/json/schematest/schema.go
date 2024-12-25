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

func WithPatternProperty(pattern string, ps *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		if s.PatternProperties == nil {
			s.PatternProperties = map[string]*schema.Ref{}
		}
		s.PatternProperties[pattern] = &schema.Ref{Value: ps}
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

func WithUnevaluatedItems(ref *schema.Ref) SchemaOptions {
	return func(s *schema.Schema) {
		s.UnevaluatedItems = ref
	}
}

func WithPrefixItems(items ...*schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		for _, item := range items {
			s.PrefixItems = append(s.PrefixItems, &schema.Ref{Value: item})
		}

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

func WithAllOf(schemas ...*schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		for _, all := range schemas {
			s.AllOf = append(s.AllOf, &schema.Ref{Value: all})
		}
	}
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
			s.ExclusiveMinimum = schema.NewUnionTypeA[float64, bool](min)
		} else {
			s.ExclusiveMinimum.A = min
		}
	}
}

func WithExclusiveMinimumFlag(b bool) SchemaOptions {
	return func(s *schema.Schema) {
		if s.ExclusiveMinimum == nil {
			s.ExclusiveMinimum = schema.NewUnionTypeB[float64, bool](b)
		} else {
			s.ExclusiveMinimum.B = b
		}
	}
}

func WithExclusiveMaximum(max float64) SchemaOptions {
	return func(s *schema.Schema) {
		if s.ExclusiveMaximum == nil {
			s.ExclusiveMaximum = schema.NewUnionTypeA[float64, bool](max)
		} else {
			s.ExclusiveMaximum.A = max
		}
	}
}

func WithExclusiveMaximumFlag(b bool) SchemaOptions {
	return func(s *schema.Schema) {
		if s.ExclusiveMaximum == nil {
			s.ExclusiveMaximum = schema.NewUnionTypeB[float64, bool](b)
		} else {
			s.ExclusiveMaximum.B = b
		}
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

func WithExample(v interface{}) SchemaOptions {
	return func(s *schema.Schema) {
		s.Examples = append(s.Examples, v)
	}
}

func WithExamples(v ...interface{}) SchemaOptions {
	return func(s *schema.Schema) {
		s.Examples = append(s.Examples, v...)
	}
}

func WithUnevaluatedProperties(b bool) SchemaOptions {
	return func(s *schema.Schema) {
		s.UnevaluatedProperties = &schema.Ref{Boolean: &b}
	}
}

func WithPropertyNames(propSchema *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		s.PropertyNames = &schema.Ref{Value: propSchema}
	}
}

func WithContains(ref *schema.Ref) SchemaOptions {
	return func(s *schema.Schema) {
		s.Contains = ref
	}
}

func WithMinContains(n int) SchemaOptions {
	return func(s *schema.Schema) {
		s.MinContains = &n
	}
}

func WithMaxContains(n int) SchemaOptions {
	return func(s *schema.Schema) {
		s.MaxContains = &n
	}
}
