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
		if n <= *s.ExclusiveMinimum {
			return fmt.Errorf("%v is lower or equal as the required minimum %v, expected %v", n, *s.ExclusiveMinimum, s)
		}
	} else if s.Minimum != nil && n < *s.Minimum {
		return fmt.Errorf("%v is lower as the required minimum %v, expected %v", n, *s.Minimum, s)
	}

	if s.ExclusiveMaximum != nil {
		if n >= *s.ExclusiveMaximum {
			return fmt.Errorf("%v is greater or equal as the required maximum %v, expected %v", n, *s.ExclusiveMaximum, s)
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
