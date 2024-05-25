package parser

import (
	"fmt"
	"mokapi/schema/json/schema"
	"strings"
)

func (p *Parser) ParseBoolean(i interface{}, s *schema.Schema) (bool, error) {
	switch v := i.(type) {
	case bool:
		return v, nil
	case string:
		if p.ConvertStringToBoolean {
			switch strings.ToLower(v) {
			case "true":
				return true, nil
			case "false":
				return false, nil
			}
		}
		return false, fmt.Errorf("parse %v failed, expected %v", i, s)
	case int, int64:
		return v != 0, nil
	}
	return false, fmt.Errorf("parse %v failed, expected %v", i, s)
}
