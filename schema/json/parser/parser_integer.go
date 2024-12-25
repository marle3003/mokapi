package parser

import (
	"math"
	"mokapi/schema/json/schema"
	"strconv"
)

func (p *Parser) ParseInteger(i interface{}, s *schema.Schema) (n int64, err error) {
	switch v := i.(type) {
	case int:
		n = int64(v)
	case int64:
		n = v
	case float64:
		if math.Trunc(v) != v {
			return 0, Errorf("type", "invalid type, expected %v but got %v", s.Type, toType(i))
		}
		n = int64(v)
	case int32:
		n = int64(v)
	case string:
		if !p.ConvertStringToNumber {
			return 0, Errorf("type", "invalid type, expected %v but got %v", s.Type, toType(i))
		}
		switch s.Format {
		case "int32":
			n32, err := strconv.Atoi(v)
			if err != nil {
				return 0, Errorf("type", "invalid type, expected %v but got %v", s.Type, toType(i))
			}
			n = int64(n32)
		default:
			n, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return 0, Errorf("type", "invalid type, expected %v but got %v", s.Type, toType(i))
			}
		}
	default:
		return 0, Errorf("type", "invalid type, expected %v but got %v", s.Type, toType(i))
	}

	if !p.SkipValidationFormatKeyword {
		switch s.Format {
		case "int32":
			if n < math.MinInt32 {
				return 0, Errorf("format", "integer '%v' does not match format 'int32': value is lower than int32 min value", i)
			}
			if n > math.MaxInt32 {
				return 0, Errorf("format", "integer '%v' does not match format 'int32': value is greater than int32 max value", i)
			}
		}
	}

	if s.MultipleOf != nil {
		m := float64(n) / *s.MultipleOf
		if m != float64(int(m)) {
			return 0, Errorf("multipleOf", "integer %v is not a multiple of %v", n, *s.MultipleOf)
		}
	}

	if err := validateIntegerMinimum(n, s); err != nil {
		return 0, err
	}
	if err := validateIntegerMaximum(n, s); err != nil {
		return 0, err
	}

	if len(s.Enum) > 0 {
		return n, checkValueIsInEnum(n, s.Enum, &schema.Schema{Type: schema.Types{"integer"}})
	}

	return n, nil
}

func validateIntegerMinimum(n int64, s *schema.Schema) error {
	if s.Minimum != nil {
		if n < int64(*s.Minimum) {
			return Errorf("minimum", "integer %v is less than minimum value of %v", n, *s.Minimum)
		} else if n == int64(*s.Minimum) && s.ExclusiveMinimum != nil && s.ExclusiveMinimum.B {
			return Errorf("minimum", "integer %v equals minimum value of %v and exclusive minimum is true", n, *s.Minimum)
		}
	}
	if s.ExclusiveMinimum != nil && s.ExclusiveMinimum.IsA() {
		if n < int64(s.ExclusiveMinimum.A) {
			return Errorf("exclusiveMinimum", "integer %v is less than minimum value of %v", n, s.ExclusiveMinimum.A)
		} else if n == int64(s.ExclusiveMinimum.A) {
			return Errorf("exclusiveMinimum", "integer %v equals minimum value of %v", n, s.ExclusiveMinimum.A)
		}
	}
	return nil
}

func validateIntegerMaximum(n int64, s *schema.Schema) error {
	if s.Maximum != nil {
		if n > int64(*s.Maximum) {
			return Errorf("maximum", "integer %v exceeds maximum value of %v", n, *s.Maximum)
		} else if n == int64(*s.Maximum) && s.ExclusiveMaximum != nil && s.ExclusiveMaximum.B {
			return Errorf("maximum", "integer %v equals maximum value of %v and exclusive maximum is true", n, *s.Maximum)
		}
	}
	if s.ExclusiveMaximum != nil && s.ExclusiveMaximum.IsA() {
		if n > int64(s.ExclusiveMaximum.A) {
			return Errorf("exclusiveMaximum", "integer %v exceeds maximum value of %v", n, s.ExclusiveMaximum.A)
		} else if n == int64(s.ExclusiveMaximum.A) {
			return Errorf("exclusiveMaximum", "integer %v equals maximum value of %v", n, s.ExclusiveMaximum.A)
		}
	}
	return nil
}
