package schema

import "encoding/json"

type Types []string

func (t *Types) UnmarshalJSON(b []byte) error {
	var v interface{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	if str, ok := v.(string); ok {
		*t = append(*t, str)
	} else if arr, ok := v.([]interface{}); ok {
		for _, i := range arr {
			if str, ok := i.(string); ok {
				*t = append(*t, str)
			} else {
				return &UnmarshalError{Value: i, Field: "type"}
			}
		}
	} else {
		return &UnmarshalError{Value: v, Field: "type"}
	}
	return nil
}

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

func (s *Schema) HasProperties() bool {
	if s == nil {
		return false
	}
	return s.Properties != nil && s.Properties.Len() > 0
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

func (s *Schema) IsAnyString() bool {
	if !s.IsString() {
		return false
	}
	return s.Pattern == "" && s.Format == "" && s.MinLength == nil && s.MaxLength == nil
}
