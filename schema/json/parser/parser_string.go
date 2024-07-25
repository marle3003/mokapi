package parser

import (
	"fmt"
	"github.com/google/uuid"
	"mokapi/schema/json/schema"
	"mokapi/version"
	"net"
	"net/mail"
	"reflect"
	"regexp"
	"strings"
	"time"
)

func (p *Parser) ParseString(data interface{}, schema *schema.Schema) (interface{}, error) {
	var s string
	switch v := data.(type) {
	case string:
		s = v
	case version.Version:
		s = v.String()
	default:
		if v == nil && schema.IsNullable() {
			return nil, nil
		}
		return nil, fmt.Errorf("parse %v failed, expected %v", data, schema)
	}

	return s, validateString(s, schema)
}

func validateString(i interface{}, s *schema.Schema) error {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.String {
		return fmt.Errorf("validation error on %v, expected %v", v.Kind(), s)
	}

	str := i.(string)
	switch s.Format {
	case "date":
		_, err := time.Parse("2006-01-02", str)
		if err != nil {
			return fmt.Errorf("value '%v' does not match format 'date' (RFC3339), expected %v", str, s)
		}
		return nil
	case "date-time":
		_, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return fmt.Errorf("value '%v' does not match format 'date-time' (RFC3339), expected %v", str, s)
		}
		return nil
	case "email":
		_, err := mail.ParseAddress(str)
		if err != nil {
			return fmt.Errorf("value '%v' does not match format 'email', expected %v", str, s)
		}
		return nil
	case "uuid":
		_, err := uuid.Parse(str)
		if err != nil {
			return fmt.Errorf("value '%v' does not match format 'uuid', expected %v", str, s)
		}
		return nil
	case "ipv4":
		ip := net.ParseIP(str)
		if ip == nil {
			return fmt.Errorf("value '%v' does not match format 'ipv4', expected %v", str, s)
		}
		if len(strings.Split(str, ".")) != 4 {
			return fmt.Errorf("value '%v' does not match format 'ipv4', expected %v", str, s)
		}
		return nil
	case "ipv6":
		ip := net.ParseIP(str)
		if ip == nil {
			return fmt.Errorf("value '%v' does not match format 'ipv6', expected %v", str, s)
		}
		if len(strings.Split(str, ":")) != 8 {
			return fmt.Errorf("value '%v' does not match format 'ipv6', expected %v", str, s)
		}
		return nil
	}

	if len(s.Pattern) > 0 {
		p, err := regexp.Compile(s.Pattern)
		if err != nil {
			return err
		}
		if p.MatchString(str) {
			return nil
		}
		return fmt.Errorf("value '%v' does not match pattern, expected %v", str, s)
	}

	if s.MinLength != nil && *s.MinLength > len(str) {
		return fmt.Errorf("length of '%v' is too short, expected %v", str, s)
	}
	if s.MaxLength != nil && *s.MaxLength < len(str) {
		return fmt.Errorf("length of '%v' is too long, expected %v", str, s)
	}

	if len(s.Enum) > 0 {
		return checkValueIsInEnum(str, s.Enum, &schema.Schema{Type: schema.Types{"string"}})
	}

	return nil
}
