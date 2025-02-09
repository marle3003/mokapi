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

func NewBool(b bool) *schema.Schema {
	s := new(schema.Schema)
	s.Boolean = &b
	return s
}

func WithProperty(name string, ps *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		if s.Properties == nil {
			s.Properties = &schema.Schemas{}
		}
		s.Properties.Set(name, ps)
	}
}

func WithPropertyRef(name string, r string) SchemaOptions {
	return func(s *schema.Schema) {
		if s.Properties == nil {
			s.Properties = &schema.Schemas{}
		}
		s.Properties.Set(name, &schema.Schema{Ref: r})
	}
}

func WithPropertyNew(name string, prop *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		if s.Properties == nil {
			s.Properties = &schema.Schemas{}
		}
		s.Properties.Set(name, prop)
	}
}

func WithPatternProperty(pattern string, ps *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		if s.PatternProperties == nil {
			s.PatternProperties = map[string]*schema.Schema{}
		}
		s.PatternProperties[pattern] = ps
	}
}

func WithItems(typeName string, opts ...SchemaOptions) SchemaOptions {
	return func(s *schema.Schema) {
		s.Items = New(typeName, opts...)
	}
}

func WithItemsNew(items *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		s.Items = items
	}
}

func WithItemsRefString(r string) SchemaOptions {
	return func(s *schema.Schema) {
		s.Items = &schema.Schema{Ref: r}
	}
}

func WithUnevaluatedItems(items *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		s.UnevaluatedItems = items
	}
}

func WithPrefixItems(items ...*schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		for _, item := range items {
			s.PrefixItems = append(s.PrefixItems, item)
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
			s.AnyOf = append(s.AnyOf, any)
		}
	}
}

func NewAny(schemas ...*schema.Schema) *schema.Schema {
	s := &schema.Schema{}
	for _, any := range schemas {
		s.AnyOf = append(s.AnyOf, any)
	}
	return s
}

func OneOf(schemas ...*schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		for _, one := range schemas {
			s.OneOf = append(s.OneOf, one)
		}
	}
}

func NewOneOf(schemas ...*schema.Schema) *schema.Schema {
	s := &schema.Schema{}
	for _, one := range schemas {
		s.OneOf = append(s.OneOf, one)
	}
	return s
}

func AllOf(schemas ...*schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		for _, all := range schemas {
			s.AllOf = append(s.AllOf, all)
		}
	}
}

func NewAllOf(schemas ...*schema.Schema) *schema.Schema {
	s := &schema.Schema{}
	for _, all := range schemas {
		s.AllOf = append(s.AllOf, all)
	}
	return s
}

func WithAllOf(schemas ...*schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		for _, all := range schemas {
			s.AllOf = append(s.AllOf, all)
		}
	}
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
		s.AdditionalProperties = additional
	}
}

func WithFreeForm(allowed bool) SchemaOptions {
	return func(s *schema.Schema) {
		s.AdditionalProperties = &schema.Schema{Boolean: &allowed}
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

func WithEnumValues(e ...interface{}) SchemaOptions {
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

func WithUnevaluatedProperties(ref *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		s.UnevaluatedProperties = ref
	}
}

func WithPropertyNames(propSchema *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		s.PropertyNames = propSchema
	}
}

func WithContains(ref *schema.Schema) SchemaOptions {
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

func WithConst(c interface{}) SchemaOptions {
	return func(s *schema.Schema) {
		s.Const = &c
	}
}

func WithNot(not *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		s.Not = not
	}
}

func WithDependentRequired(prop string, required ...string) SchemaOptions {
	return func(s *schema.Schema) {
		if s.DependentRequired == nil {
			s.DependentRequired = map[string][]string{}
		}
		s.DependentRequired[prop] = required
	}
}

func WithDependentSchemas(prop string, dependentSchema *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		if s.DependentSchemas == nil {
			s.DependentSchemas = map[string]*schema.Schema{}
		}
		s.DependentSchemas[prop] = dependentSchema
	}
}

func WithIf(condition *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		s.If = condition
	}
}

func WithThen(condition *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		s.Then = condition
	}
}

func WithElse(condition *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		s.Else = condition
	}
}

func WithId(id string) SchemaOptions {
	return func(s *schema.Schema) {
		s.Id = id
	}
}

func WithAnchor(anchor string) SchemaOptions {
	return func(s *schema.Schema) {
		s.Anchor = anchor
	}
}

func WithDynamicAnchor(anchor string) SchemaOptions {
	return func(s *schema.Schema) {
		s.DynamicAnchor = anchor
	}
}

func WithDef(name string, def *schema.Schema) SchemaOptions {
	return func(s *schema.Schema) {
		if s.Defs == nil {
			s.Defs = map[string]*schema.Schema{}
		}
		s.Defs[name] = def
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
