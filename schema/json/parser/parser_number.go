package parser

import (
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
					return 0, Errorf("format", "number '%v' does not match format 'float'", i)
				}
			}
		}
	case string:
		if !p.ConvertStringToNumber {
			return 0, Errorf("type", "invalid type, expected number but got %v", toType(i))
		}
		f, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, Errorf("type", "invalid type, expected number but got %v", toType(i))
		}
	case int:
		f = float64(v)
	case int64:
		f = float64(v)
	default:
		return 0, Errorf("type", "invalid type, expected number but got %v", toType(i))
	}

	if s.MultipleOf != nil {
		m := f / *s.MultipleOf
		if m != float64(int(m)) {
			return 0, Errorf("multipleOf", "number %v is not a multiple of %v", f, *s.MultipleOf)
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
			return Errorf("minimum", "number %v is less than minimum value of %v", n, *s.Minimum)
		} else if n == *s.Minimum && s.ExclusiveMinimum != nil && s.ExclusiveMinimum.B {
			return Errorf("minimum", "number %v equals minimum value of %v and exclusive minimum is true", n, *s.Minimum)
		}
	}
	if s.ExclusiveMinimum != nil && s.ExclusiveMinimum.IsA() {
		if n < s.ExclusiveMinimum.A {
			return Errorf("exclusiveMinimum", "number %v is less than minimum value of %v", n, s.ExclusiveMinimum.A)
		} else if n == s.ExclusiveMinimum.A {
			return Errorf("exclusiveMinimum", "number %v equals minimum value of %v", n, s.ExclusiveMinimum.A)
		}
	}
	return nil
}

func validateNumberMaximum(n float64, s *schema.Schema) error {
	if s.Maximum != nil {
		if n > *s.Maximum {
			return Errorf("maximum", "number %v exceeds maximum value of %v", n, *s.Maximum)
		} else if n == *s.Maximum && s.ExclusiveMaximum != nil && s.ExclusiveMaximum.B {
			return Errorf("maximum", "number %v equals maximum value of %v and exclusive maximum is true", n, *s.Maximum)
		}
	}
	if s.ExclusiveMaximum != nil && s.ExclusiveMaximum.IsA() {
		if n > s.ExclusiveMaximum.A {
			return Errorf("exclusiveMaximum", "number %v exceeds maximum value of %v", n, s.ExclusiveMaximum.A)
		} else if n == s.ExclusiveMaximum.A {
			return Errorf("exclusiveMaximum", "number %v equals maximum value of %v", n, s.ExclusiveMaximum.A)
		}
	}
	return nil
}
