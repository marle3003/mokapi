package parser

import (
	"fmt"
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
				return false, &ErrorDetail{
					Message: fmt.Sprintf("invalid type, expected boolean but got %v", toType(i)),
					Field:   "type",
				}
			}
		} else {
			return false, &ErrorDetail{
				Message: fmt.Sprintf("invalid type, expected boolean but got %v", toType(i)),
				Field:   "type",
			}
		}
	default:
		return false, &ErrorDetail{
			Message: fmt.Sprintf("invalid type, expected boolean but got %v", toType(i)),
			Field:   "type",
		}
	}

	if len(s.Enum) > 0 {
		return b, checkValueIsInEnum(b, s.Enum, &schema.Schema{Type: schema.Types{"integer"}, Format: "int64"})
	}

	return b, nil
}
