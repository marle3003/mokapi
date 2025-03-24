package schema

import "strings"

func (s *Schema) apply(ref *Schema) {
	if ref == nil {
		return
	}
	if s.SubSchema == nil {
		s.SubSchema = ref.SubSchema
		return
	}
	if len(s.m) == 0 {
		s.SubSchema = ref.SubSchema
		return
	}

	if s.isEmpty() && s.Boolean == nil && ref.Boolean != nil {
		s.Boolean = ref.Boolean
		return
	}

	if len(s.Type) == 0 {
		s.Type = ref.Type
	}
	if s.Enum == nil {
		s.Enum = ref.Enum
	}
	if s.Const == nil {
		s.Const = ref.Const
	}
	if len(s.Format) == 0 {
		s.Format = ref.Format
	}

	if len(s.Pattern) == 0 {
		s.Pattern = ref.Pattern
	}
	if s.MinLength == nil {
		s.MinLength = ref.MinLength
	}
	if s.MaxLength == nil {
		s.MaxLength = ref.MaxLength
	}
	if s.MultipleOf == nil {
		s.MultipleOf = ref.MultipleOf
	}
	if s.Minimum == nil {
		s.Minimum = ref.Minimum
	}
	if s.Maximum == nil {
		s.Maximum = ref.Maximum
	}
	if s.ExclusiveMinimum == nil {
		s.ExclusiveMinimum = ref.ExclusiveMinimum
	}
	if s.ExclusiveMaximum == nil {
		s.ExclusiveMaximum = ref.ExclusiveMaximum
	}
	if s.Items == nil {
		s.Items = ref.Items
	}

	if _, ok := s.m["uniqueItems"]; !ok {
		s.UniqueItems = ref.UniqueItems
	}

	if s.MinItems == nil {
		s.MinItems = ref.MinItems
	}
	if s.MaxItems == nil {
		s.MaxItems = ref.MaxItems
	}

	if _, ok := s.m["shuffleItems"]; !ok {
		s.ShuffleItems = ref.ShuffleItems
	}

	if s.Properties == nil {
		s.Properties = ref.Properties
	}

	if s.Required == nil {
		s.Required = ref.Required
	}

	if s.AdditionalProperties == nil {
		s.AdditionalProperties = ref.AdditionalProperties
	} else {
		s.AdditionalProperties.apply(ref.AdditionalProperties)
	}

	if s.MinProperties == nil {
		s.MinProperties = ref.MinProperties
	}
	if s.MaxProperties == nil {
		s.MaxProperties = ref.MaxProperties
	}
	if len(s.Title) == 0 {
		s.Title = ref.Title
	}
	if len(s.Description) == 0 {
		s.Description = ref.Description
	}
	if s.Default == nil {
		s.Default = ref.Default
	}

	if _, ok := s.m["deprecated"]; !ok {
		s.Deprecated = ref.Deprecated
	}

	if s.Examples == nil {
		s.Examples = ref.Examples
	}
	if len(s.ContentMediaType) == 0 {
		s.ContentMediaType = ref.ContentMediaType
	}
	if len(s.ContentEncoding) == 0 {
		s.ContentEncoding = ref.ContentEncoding
	}

	if s.Definitions == nil {
		s.Definitions = ref.Definitions
	} else {
		for k, v := range ref.Definitions {
			if def, ok := s.Definitions[k]; ok {
				def.apply(v)
			} else {
				s.Definitions[k] = v
			}
		}
	}

	if s.Defs == nil {
		s.Defs = ref.Defs
	} else {
		for k, v := range ref.Defs {
			if def, ok := s.Defs[k]; ok {
				def.apply(v)
			} else {
				s.Defs[k] = v
			}
		}
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
