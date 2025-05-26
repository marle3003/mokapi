package parser

import (
	"fmt"
	"math"
	"mokapi/schema/json/schema"
	"strconv"
)

func (p *Parser) ParseInteger(i interface{}, s *schema.Schema) (interface{}, error) {
	var n int64
	switch v := i.(type) {
	case int:
		n = int64(v)
	case int64:
		n = v
	case float64:
		if math.Trunc(v) != v {
			return 0, &ErrorDetail{
				Message: fmt.Sprintf("invalid type, expected %v but got %v", s.Type, toType(i)),
				Field:   "type",
			}
		}
		n = int64(v)
	case int32:
		i = v
		n = int64(v)
	case string:
		if !p.ConvertStringToNumber {
			return 0, &ErrorDetail{
				Message: fmt.Sprintf("invalid type, expected %v but got %v", s.Type, toType(i)),
				Field:   "type",
			}
		}
		switch s.Format {
		case "int32":
			n32, err := strconv.Atoi(v)
			if err != nil {
				return 0, &ErrorDetail{
					Message: fmt.Sprintf("invalid type, expected %v but got %v", s.Type, toType(i)),
					Field:   "type",
				}
			}
			n = int64(n32)
		default:
			var err error
			n, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return 0, &ErrorDetail{
					Message: fmt.Sprintf("invalid type, expected %v but got %v", s.Type, toType(i)),
					Field:   "type",
				}
			}
		}
	default:
		return 0, &ErrorDetail{
			Message: fmt.Sprintf("invalid type, expected %v but got %v", s.Type, toType(i)),
			Field:   "type",
		}
	}

	if !p.SkipValidationFormatKeyword {
		switch s.Format {
		case "int32":
			if n < math.MinInt32 || n > math.MaxInt32 {
				return 0, &ErrorDetail{
					Message: fmt.Sprintf("integer '%v' does not match format 'int32'", i),
					Field:   "format",
				}
			}
		}
	}

	if s.MultipleOf != nil {
		m := float64(n) / *s.MultipleOf
		if m != float64(int(m)) {
			return 0, &ErrorDetail{
				Message: fmt.Sprintf("integer %v is not a multiple of %v", n, *s.MultipleOf),
				Field:   "multipleOf",
			}
		}
	}

	if err := validateIntegerMinimum(n, s); err != nil {
		return 0, err
	}
	if err := validateIntegerMaximum(n, s); err != nil {
		return 0, err
	}

	if s.Format == "int32" && !p.SkipValidationFormatKeyword {
		i = int32(n)
	} else {
		i = n
	}

	if len(s.Enum) > 0 {
		return i, checkValueIsInEnum(n, s.Enum, &schema.Schema{Type: schema.Types{"integer"}})
	}

	return i, nil
}

func validateIntegerMinimum(n int64, s *schema.Schema) error {
	if s.Minimum != nil {
		if n < int64(*s.Minimum) {
			return &ErrorDetail{
				Message: fmt.Sprintf("integer %v is less than minimum value of %v", n, *s.Minimum),
				Field:   "minimum",
			}
		} else if n == int64(*s.Minimum) && s.ExclusiveMinimum != nil && s.ExclusiveMinimum.B {
			return &ErrorDetail{
				Message: fmt.Sprintf("integer %v equals minimum value of %v and exclusive minimum is true", n, *s.Minimum),
				Field:   "minimum",
			}
		}
	}
	if s.ExclusiveMinimum != nil && s.ExclusiveMinimum.IsA() {
		if n < int64(s.ExclusiveMinimum.A) {
			return &ErrorDetail{
				Message: fmt.Sprintf("integer %v is less than minimum value of %v", n, s.ExclusiveMinimum.A),
				Field:   "exclusiveMinimum",
			}
		} else if n == int64(s.ExclusiveMinimum.A) {
			return &ErrorDetail{
				Message: fmt.Sprintf("integer %v equals minimum value of %v", n, s.ExclusiveMinimum.A),
				Field:   "exclusiveMinimum",
			}
		}
	}
	return nil
}

func validateIntegerMaximum(n int64, s *schema.Schema) error {
	if s.Maximum != nil {
		if n > int64(*s.Maximum) {
			return &ErrorDetail{
				Message: fmt.Sprintf("integer %v exceeds maximum value of %v", n, *s.Maximum),
				Field:   "maximum",
			}
		} else if n == int64(*s.Maximum) && s.ExclusiveMaximum != nil && s.ExclusiveMaximum.B {
			return &ErrorDetail{
				Message: fmt.Sprintf("integer %v equals maximum value of %v and exclusive maximum is true", n, *s.Maximum),
				Field:   "maximum",
			}
		}
	}
	if s.ExclusiveMaximum != nil && s.ExclusiveMaximum.IsA() {
		if n > int64(s.ExclusiveMaximum.A) {
			return &ErrorDetail{
				Message: fmt.Sprintf("integer %v exceeds maximum value of %v", n, s.ExclusiveMaximum.A),
				Field:   "exclusiveMaximum",
			}
		} else if n == int64(s.ExclusiveMaximum.A) {
			return &ErrorDetail{
				Message: fmt.Sprintf("integer %v equals maximum value of %v", n, s.ExclusiveMaximum.A),
				Field:   "exclusiveMaximum",
			}
		}
	}
	return nil
}
