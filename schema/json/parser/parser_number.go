package parser

import (
	"fmt"
	"mokapi/schema/json/schema"
	"strconv"
)

func (p *Parser) ParseNumber(i interface{}, s *schema.Schema) (f float64, err error) {
	switch v := i.(type) {
	case float32:
		f = float64(v)
	case float64:
		f = v

		if !p.SkipValidationFormatKeyword {
			switch s.Format {
			case "float":
				f32 := float32(v)

				if float64(f32) != v {
					return 0, &ErrorDetail{
						Message: fmt.Sprintf("number '%v' does not match format 'float'", i),
						Field:   "format",
					}
				}
			}
		}
	case string:
		if !p.ConvertStringToNumber {
			return 0, &ErrorDetail{
				Message: fmt.Sprintf("invalid type, expected number but got %v", toType(i)),
				Field:   "type",
			}
		}
		f, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, &ErrorDetail{
				Message: fmt.Sprintf("invalid type, expected number but got %v", toType(i)),
				Field:   "type",
			}
		}
	case int:
		f = float64(v)
	case int64:
		f = float64(v)
	default:
		return 0, &ErrorDetail{
			Message: fmt.Sprintf("invalid type, expected number but got %v", toType(i)),
			Field:   "type",
		}
	}

	if s.MultipleOf != nil {
		m := f / *s.MultipleOf
		if m != float64(int(m)) {
			return 0, &ErrorDetail{
				Message: fmt.Sprintf("number %v is not a multiple of %v", f, *s.MultipleOf),
				Field:   "multipleOf",
			}
		}
	}

	if err := validateNumberMinimum(f, s); err != nil {
		return 0, err
	}
	if err := validateNumberMaximum(f, s); err != nil {
		return 0, err
	}

	if len(s.Enum) > 0 {
		return f, checkValueIsInEnum(f, s.Enum, &schema.Schema{Type: schema.Types{"number"}})
	}

	return f, nil
}

func validateNumberMinimum(n float64, s *schema.Schema) error {
	if s.Minimum != nil {
		if n < *s.Minimum {
			return &ErrorDetail{
				Message: fmt.Sprintf("number %v is less than minimum value of %v", n, *s.Minimum),
				Field:   "minimum",
			}
		} else if n == *s.Minimum && s.ExclusiveMinimum != nil && s.ExclusiveMinimum.B {
			return &ErrorDetail{
				Message: fmt.Sprintf("number %v equals minimum value of %v and exclusive minimum is true", n, *s.Minimum),
				Field:   "minimum",
			}
		}
	}
	if s.ExclusiveMinimum != nil && s.ExclusiveMinimum.IsA() {
		if n < s.ExclusiveMinimum.A {
			return &ErrorDetail{
				Message: fmt.Sprintf("number %v is less than minimum value of %v", n, s.ExclusiveMinimum.A),
				Field:   "exclusiveMinimum",
			}
		} else if n == s.ExclusiveMinimum.A {
			return &ErrorDetail{
				Message: fmt.Sprintf("number %v equals minimum value of %v", n, s.ExclusiveMinimum.A),
				Field:   "exclusiveMinimum",
			}
		}
	}
	return nil
}

func validateNumberMaximum(n float64, s *schema.Schema) error {
	if s.Maximum != nil {
		if n > *s.Maximum {
			return &ErrorDetail{
				Message: fmt.Sprintf("number %v exceeds maximum value of %v", n, *s.Maximum),
				Field:   "maximum",
			}
		} else if n == *s.Maximum && s.ExclusiveMaximum != nil && s.ExclusiveMaximum.B {
			return &ErrorDetail{
				Message: fmt.Sprintf("number %v equals maximum value of %v and exclusive maximum is true", n, *s.Maximum),
				Field:   "maximum",
			}
		}
	}
	if s.ExclusiveMaximum != nil && s.ExclusiveMaximum.IsA() {
		if n > s.ExclusiveMaximum.A {
			return &ErrorDetail{
				Message: fmt.Sprintf("number %v exceeds maximum value of %v", n, s.ExclusiveMaximum.A),
				Field:   "exclusiveMaximum",
			}
		} else if n == s.ExclusiveMaximum.A {
			return &ErrorDetail{
				Message: fmt.Sprintf("number %v equals maximum value of %v", n, s.ExclusiveMaximum.A),
				Field:   "exclusiveMaximum",
			}
		}
	}
	return nil
}
