package parser

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"mokapi/schema/json/schema"
	"mokapi/version"
	"net"
	"net/mail"
	"regexp"
	"regexp/syntax"
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
	case []byte:
		s = string(v)
	default:
		if v == nil {
			if schema.IsNullable() {
				return nil, nil
			}
			return nil, Errorf("type", "invalid type, expected string but got null")
		}
		return nil, Errorf("type", "invalid type, expected %v but got %v", schema.Type, toType(data))
	}

	return s, validateString(s, schema, p.SkipValidationFormatKeyword)
}

func validateString(str string, s *schema.Schema, skipValidationFormatKeyword bool) error {
	if !skipValidationFormatKeyword {
		switch s.Format {
		case "date":
			_, err := time.Parse("2006-01-02", str)
			if err != nil {
				return Errorf("format", "string '%v' does not match format 'date'", str)
			}
		case "date-time":
			_, err := time.Parse(time.RFC3339, str)
			if err != nil {
				return Errorf("format", "string '%v' does not match format 'date-time'", str)
			}
		case "time":
			_, err := time.Parse("15:04:05Z07:00", str)
			if err != nil {
				return Errorf("format", "string '%v' does not match format 'time'", str)
			}
		case "duration":
			err := ParseDuration(str)
			if err != nil {
				return Errorf("format", "string '%v' does not match format 'duration'", str)
			}
		case "email":
			_, err := mail.ParseAddress(str)
			if err != nil {
				return Errorf("format", "string '%v' does not match format 'email'", str)
			}
		case "uuid":
			_, err := uuid.Parse(str)
			if err != nil {
				return Errorf("format", "string '%v' does not match format 'uuid'", str)
			}
		case "ipv4":
			ip := net.ParseIP(str)
			if ip == nil {
				return Errorf("format", "string '%v' does not match format 'ipv4'", str)
			}
			if len(strings.Split(str, ".")) != 4 {
				return Errorf("format", "string '%v' does not match format 'ipv4'", str)
			}
		case "ipv6":
			ip := net.ParseIP(str)
			if ip == nil {
				return Errorf("format", "string '%v' does not match format 'ipv6'", str)
			}
			if len(strings.Split(str, ":")) != 8 {
				return Errorf("format", "string '%v' does not match format 'ipv6'", str)
			}
		}
	}

	if len(s.Pattern) > 0 {
		p, err := regexp.Compile(s.Pattern)
		if err != nil {
			var sErr *syntax.Error
			var msg string
			if errors.As(err, &sErr) {
				msg = sErr.Code.String()
			} else {
				msg = err.Error()
			}
			return Errorf("pattern", "validate string '%s' with regex pattern '%s' failed: error parsing regex: %s", str, s.Pattern, msg)
		}
		if !p.MatchString(str) {
			return Errorf("pattern", "string '%v' does not match regex pattern '%v'", str, s.Pattern)
		}
	}

	if s.MinLength != nil && *s.MinLength > len(str) {
		return Errorf("minLength", "string '%v' is less than minimum of %v", str, *s.MinLength)
	}
	if s.MaxLength != nil && *s.MaxLength < len(str) {
		return Errorf("maxLength", "string '%v' exceeds maximum of %v", str, *s.MaxLength)
	}

	if len(s.Enum) > 0 {
		return checkValueIsInEnum(str, s.Enum, &schema.Schema{Type: schema.Types{"string"}})
	}

	return nil
}
