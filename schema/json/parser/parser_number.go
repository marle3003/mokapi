package parser

import (
	"fmt"
	"math"
	"mokapi/schema/json/schema"
	"strconv"
)

func (p *Parser) ParseNumber(i interface{}, s *schema.Schema) (f float64, err error) {
	switch v := i.(type) {
	case float64:
		f = v
	case string:
		if !p.ConvertStringToNumber {
			return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
		}
		f, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
		}
	case int:
		f = float64(v)
	case int64:
		f = float64(v)
	default:
		return 0, fmt.Errorf("parse '%v' failed, expected %v", v, s)
	}

	switch s.Format {
	case "float":
		if f > math.MaxFloat32 {
			return 0, fmt.Errorf("parse %v failed, expected %v", i, s)
		}
	}

	return f, validateFloat64(f, s)
}

func validateFloat64(n float64, s *schema.Schema) error {
	if s.ExclusiveMinimum != nil {
		var min float64
		if s.ExclusiveMinimum.IsA() && (s.ExclusiveMinimum.IsA() || s.ExclusiveMinimum.B) {
			min = s.ExclusiveMinimum.A
		} else {
			if s.Minimum == nil {
				return fmt.Errorf("exclusiveMinimum is set to true but no minimum value is specified")
			}
			min = *s.Minimum
		}
		if n <= min {
			return fmt.Errorf("%v is lower or equal as the required minimum %v, expected %v", n, *s.ExclusiveMinimum, s)
		}
	} else if s.Minimum != nil && n < *s.Minimum {
		return fmt.Errorf("%v is lower as the required minimum %v, expected %v", n, *s.Minimum, s)
	}

	if s.ExclusiveMaximum != nil && (s.ExclusiveMaximum.IsA() || s.ExclusiveMaximum.B) {
		var max float64
		if s.ExclusiveMaximum.IsA() {
			max = s.ExclusiveMaximum.A
		} else {
			if s.Maximum == nil {
				return fmt.Errorf("exclusiveMaximum is set to true but no maximum value is specified")
			}
			max = *s.Maximum
		}
		if n >= max {
			return fmt.Errorf("%v is greater or equal as the required maximum %v, expected %v", n, max, s)
		}
	} else if s.Maximum != nil && n > *s.Maximum {
		return fmt.Errorf("%v is greater as the required maximum %v, expected %v", n, *s.Maximum, s)
	}

	if len(s.Enum) > 0 {
		return checkValueIsInEnum(n, s.Enum, &schema.Schema{Type: schema.Types{"number"}, Format: "double"})
	}

	if s.MultipleOf != nil {
		m := n / *s.MultipleOf
		if m != float64(int(m)) {
			return fmt.Errorf("%v is not a multiple of %v: %v", n, *s.MultipleOf, s)
		}
	}

	return nil
}
