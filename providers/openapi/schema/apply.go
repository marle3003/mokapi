package schema

import "strings"

func (s *Schema) apply(ref *Schema) {
	if ref == nil {
		return
	}

	if s.isEmpty() && s.Boolean == nil && ref.Boolean != nil {
		s.Boolean = ref.Boolean
		return
	}

	if !s.isSet("type") {
		s.Type = ref.Type
	}
	if !s.isSet("enum") {
		s.Enum = ref.Enum
	}
	if !s.isSet("const") {
		s.Const = ref.Const
	}

	if !s.isSet("multipleOf") {
		s.MultipleOf = ref.MultipleOf
	}
	if !s.isSet("minimum") {
		s.Minimum = ref.Minimum
	}
	if !s.isSet("maximum") {
		s.Maximum = ref.Maximum
	}
	if !s.isSet("exclusiveMinimum") {
		s.ExclusiveMinimum = ref.ExclusiveMinimum
	}
	if !s.isSet("exclusiveMaximum") {
		s.ExclusiveMaximum = ref.ExclusiveMaximum
	}

	if !s.isSet("pattern") {
		s.Pattern = ref.Pattern
	}
	if !s.isSet("minLength") {
		s.MinLength = ref.MinLength
	}
	if !s.isSet("maxLength") {
		s.MaxLength = ref.MaxLength
	}
	if !s.isSet("format") {
		s.Format = ref.Format
	}

	if !s.isSet("items") {
		s.Items = ref.Items
	}
	if !s.isSet("prefixItems") {
		s.PrefixItems = ref.PrefixItems
	}
	if !s.isSet("unevaluatedItems") {
		s.UnevaluatedItems = ref.UnevaluatedItems
	}
	if !s.isSet("contains") {
		s.Contains = ref.Contains
	}
	if !s.isSet("maxContains") {
		s.MaxContains = ref.MaxContains
	}
	if !s.isSet("minContains") {
		s.MinContains = ref.MinContains
	}
	if !s.isSet("minContains") {
		s.MinContains = ref.MinContains
	}
	if !s.isSet("minItems") {
		s.MinItems = ref.MinItems
	}
	if !s.isSet("maxItems") {
		s.MaxItems = ref.MaxItems
	}
	if !s.isSet("uniqueItems") {
		s.UniqueItems = ref.UniqueItems
	}
	if !s.isSet("shuffleItems") {
		s.ShuffleItems = ref.ShuffleItems
	}

	if !s.isSet("properties") {
		s.Properties = ref.Properties
	}
	if !s.isSet("patternProperties") {
		s.PatternProperties = ref.PatternProperties
	}
	if !s.isSet("minProperties") {
		s.MinProperties = ref.MinProperties
	}
	if !s.isSet("maxProperties") {
		s.MaxProperties = ref.MaxProperties
	}
	if !s.isSet("required") {
		s.Required = ref.Required
	}
	if !s.isSet("dependentRequired") {
		s.DependentRequired = ref.DependentRequired
	}
	if !s.isSet("dependentSchemas") {
		s.DependentSchemas = ref.DependentSchemas
	}
	if !s.isSet("additionalProperties") {
		s.AdditionalProperties = ref.AdditionalProperties
	}
	if !s.isSet("unevaluatedProperties") {
		s.UnevaluatedProperties = ref.UnevaluatedProperties
	}
	if !s.isSet("propertyNames") {
		s.PropertyNames = ref.PropertyNames
	}

	if !s.isSet("anyOf") {
		s.AnyOf = ref.AnyOf
	}
	if !s.isSet("allOf") {
		s.AllOf = ref.AllOf
	}
	if !s.isSet("oneOf") {
		s.OneOf = ref.OneOf
	}
	if !s.isSet("not") {
		s.Not = ref.Not
	}

	if !s.isSet("if") {
		s.If = ref.If
	}
	if !s.isSet("then") {
		s.Then = ref.Then
	}
	if !s.isSet("else") {
		s.Else = ref.Else
	}

	if !s.isSet("title") {
		s.Title = ref.Title
	}
	if !s.isSet("description") {
		s.Description = ref.Description
	}
	if !s.isSet("default") {
		s.Default = ref.Default
	}
	if !s.isSet("deprecated") {
		s.Deprecated = ref.Deprecated
	}
	if !s.isSet("examples") {
		s.Examples = ref.Examples
	}
	if !s.isSet("example") {
		s.Example = ref.Example
	}

	if !s.isSet("contentMediaType") {
		s.ContentMediaType = ref.ContentMediaType
	}
	if !s.isSet("contentEncoding") {
		s.ContentEncoding = ref.ContentEncoding
	}

	if !s.isSet("xml") {
		s.Xml = ref.Xml
	}
	if !s.isSet("nullable") {
		s.Nullable = ref.Nullable
	}
}

func (s *Schema) isEmpty() bool {
	for k := range s.m {
		if !strings.HasPrefix(k, "$") {
			return false
		}
	}
	return true
}

func (s *Schema) isSet(name string) bool {
	return s.m[name]
}
