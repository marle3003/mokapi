package parser

import (
	"mokapi/schema/json/schema"
	"strings"
)

func (p *Parser) parseBoolean(i interface{}, s *schema.Schema) (bool, error) {
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
				return false, Errorf("type", "invalid type, expected boolean but got %v", toType(i))
			}
		} else {
			return false, Errorf("type", "invalid type, expected boolean but got %v", toType(i))
		}
	default:
		return false, Errorf("type", "invalid type, expected boolean but got %v", toType(i))
	}

	if len(s.Enum) > 0 {
		return b, checkValueIsInEnum(b, s.Enum, &schema.Schema{Type: schema.Types{"integer"}, Format: "int64"})
	}

	return b, nil
}
