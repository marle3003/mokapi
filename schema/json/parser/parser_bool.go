package parser

import (
	"fmt"
	"mokapi/schema/json/schema"
	"strings"
)

func (p *Parser) ParseBoolean(i interface{}, s *schema.Schema) (bool, error) {
	var b bool
	switch v := i.(type) {
	case bool:
		b = v
	case string:
		if p.ConvertStringToBoolean {
			switch strings.ToLower(v) {
			case "true":
				b = true
			case "false":
				b = false
			default:
				return false, fmt.Errorf("parse '%v' (string) failed, expected %v", i, s)
			}
		} else {
			return false, fmt.Errorf("parse '%v' (string) failed, expected %v", i, s)
		}
	case int, int64:
		return false, fmt.Errorf("parse '%v' (int) failed, expected %v: invalid type", i, s)
	default:
		return false, fmt.Errorf("parse %v failed, expected %v", i, s)
	}

	if len(s.Enum) > 0 {
		return b, checkValueIsInEnum(b, s.Enum, &schema.Schema{Type: schema.Types{"integer"}, Format: "int64"})
	}

	return b, nil
}
