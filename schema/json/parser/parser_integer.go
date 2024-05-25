package parser

import (
	"fmt"
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
			return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
		}
		n = int64(v)
	case int32:
		n = int64(v)
	case string:
		if !p.ConvertStringToNumber {
			return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
		}
		switch s.Format {
		case "int64":
			n, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
			}
			return n, nil
		default:
			temp, err := strconv.Atoi(v)
			if err != nil {
				return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
			}
			n = int64(temp)
		}
	default:
		return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
	}

	switch s.Format {
	case "int32":
		if n > math.MaxInt32 || n < math.MinInt32 {
			return 0, fmt.Errorf("parse '%v' failed: represents a number either less than int32 min value or greater max value, expected %v", i, s)
		}
	}

	if s.MultipleOf != nil {
		m := float64(n) / *s.MultipleOf
		if m != float64(int(m)) {
			return 0, fmt.Errorf("%v is not a multiple of %v: %v", n, *s.MultipleOf, s)
		}
	}

	return n, validateInt64(n, s)
}

func validateInt64(n int64, s *schema.Schema) error {
	if s.ExclusiveMinimum != nil {
		if n <= int64(*s.ExclusiveMinimum) {
			return fmt.Errorf("%v is lower or equal as the required minimum %v, expected %v", n, *s.ExclusiveMinimum, s)
		}
	} else if s.Minimum != nil && n < int64(*s.Minimum) {
		return fmt.Errorf("%v is lower as the required minimum %v, expected %v", n, *s.Minimum, s)
	}

	if s.ExclusiveMaximum != nil {
		if n >= int64(*s.ExclusiveMaximum) {
			return fmt.Errorf("%v is greater or equal as the required maximum %v, expected %v", n, *s.ExclusiveMaximum, s)
		}
	} else if s.Maximum != nil && n > int64(*s.Maximum) {
		return fmt.Errorf("%v is greater as the required maximum %v, expected %v", n, *s.Maximum, s)
	}

	if len(s.Enum) > 0 {
		return checkValueIsInEnum(n, s.Enum, &schema.Schema{Type: schema.Types{"integer"}, Format: "int64"})
	}

	return nil
}
