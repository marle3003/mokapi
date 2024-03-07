package schema

func (s *Schema) IsAny() bool {
	return s == nil || len(s.Type) == 0
}

func (s *Schema) IsString() bool {
	return s.Is("string")
}

func (s *Schema) IsInteger() bool {
	return s.Is("integer")
}

func (s *Schema) IsNumber() bool {
	return s.Is("number")
}

func (s *Schema) IsArray() bool {
	return s.Is("array")
}

func (s *Schema) IsObject() bool {
	return s.Is("object")
}

func (s *Schema) IsNullable() bool {
	return s.Is("null")
}

func (s *Schema) IsDictionary() bool {
	if s == nil {
		return false
	}
	return s.AdditionalProperties.Ref != nil && s.AdditionalProperties.Value != nil
}

func (s *Schema) Is(typeName string) bool {
	if s == nil {
		return false
	}
	for _, t := range s.Type {
		if t == typeName {
			return true
		}
	}
	return false
}

func (s *Schema) IsFreeFrom() bool {
	return s.IsObject() && s.Properties == nil || s.Properties.Len() == 0
}
