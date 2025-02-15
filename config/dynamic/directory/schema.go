package directory

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type Schema struct {
	AttributeTypes map[string]*AttributeType
	ObjectClasses  map[string]*ObjectClass

	e Entry
}

type AttributeType struct {
	Id          string
	Name        []string
	Description string
	Equality    string
	Syntax      string
}

func NewSchema(e Entry) (*Schema, error) {
	s := &Schema{e: e, AttributeTypes: make(map[string]*AttributeType)}
	for _, v := range e.Attributes["attributeTypes"] {
		a, err := parseAttributeType(v)
		if err != nil {
			return nil, err
		}
		for _, name := range a.Name {
			s.AttributeTypes[name] = a
		}
	}
	for _, v := range e.Attributes["objectClasses"] {
		c, err := parseObjectClass(v)
		if err != nil {
			return nil, err
		}
		for _, name := range c.Name {
			s.ObjectClasses[name] = c
		}
	}
	return s, nil
}

func (a *AttributeType) Validate(value string) bool {
	switch a.Syntax {
	case "1.3.6.1.4.1.1466.115.121.1.5":
		// OctetString
		return true
	case "1.3.6.1.4.1.1466.115.121.1.7":
		// TRUE, FALSE
		return strings.EqualFold(value, "true") || strings.EqualFold(value, "false")
	case "1.3.6.1.4.1.1466.115.121.1.8":
		// Certificate
		return true
	case "1.3.6.1.4.1.1466.115.121.1.9":
		// Certificate List
		return true
	case "1.3.6.1.4.1.1466.115.121.1.10":
		// 	Certificate Pair
		return true
	case "1.3.6.1.4.1.1466.115.121.1.11":
		// Country String
		return len(value) == 2 && isPrintable(value)
	case "1.3.6.1.4.1.1466.115.121.1.12":
		// Distinguished Name
		return isPrintable(value)
	case "1.3.6.1.4.1.1466.115.121.1.14":
		// Delivery Method
		return utf8.Valid([]byte(value))
	case "1.3.6.1.4.1.1466.115.121.1.15":
		// Directory String
		return utf8.Valid([]byte(value))
	case "1.3.6.1.4.1.1466.115.121.1.21":
		// Enhanced Guide
		return utf8.Valid([]byte(value))
	case "1.3.6.1.4.1.1466.115.121.1.22":
		// Facsimile Telephone Number
		return isPrintable(value)
	case "1.3.6.1.4.1.1466.115.121.1.23":
		// Fax
		return isPrintable(value)
	case "1.3.6.1.4.1.1466.115.121.1.24":
		// Generalized Time
		return isTime(value)
	case "1.3.6.1.4.1.1466.115.121.1.26":
		// ASCII-only
		for _, ch := range value {
			if ch > unicode.MaxASCII {
				return false
			}
		}
		return true
	case "1.3.6.1.4.1.1466.115.121.1.27":
		// +/- 62 digit integer
		_, err := strconv.ParseInt(value, 10, 64)
		return err == nil
	case "1.3.6.1.4.1.1466.115.121.1.28":
		// JPEG
		return true
	case "1.3.6.1.4.1.1466.115.121.1.40":
		return true
	case "1.3.6.1.4.1.1466.115.121.1.50":
		return isPrintable(value)
	}

	return true
}

func parseAttributeType(s string) (*AttributeType, error) {
	// Regular expression to match components of the definition
	regex := regexp.MustCompile(`\(\s*([\d\.]+)\s+NAME\s+((?:\([^\)]*\)|'[^']*')(?:\s+(?:'[^']*'|\([^\)]*\)))*)(?:\s+DESC\s+'([^']+)')?(?:\s+EQUALITY\s+(\S+))?(?:\s+SYNTAX\s+(\S+))?`)

	// Match the definition string
	matches := regex.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("invalid attribute type definition: %s", s)
	}

	// Parse the NAME field (single or multiple)
	var names []string
	if strings.HasPrefix(matches[2], "(") { // Multiple names
		// Trim parentheses and split by spaces
		names = strings.Fields(strings.Trim(matches[2], "()"))
		// Remove surrounding quotes from each name
		for i, name := range names {
			names[i] = strings.Trim(name, "'")
		}
	} else { // Single name
		names = []string{strings.Trim(matches[2], "'")}
	}

	// Create an AttributeType struct from the matches
	attr := &AttributeType{
		Id:          matches[1],
		Name:        names,
		Description: matches[3],
		Equality:    matches[4],
		Syntax:      matches[5],
	}

	return attr, nil
}

type ObjectClass struct {
	Id          string
	Name        []string
	Description string
	SuperClass  []string
	Type        string
	Must        []string
	May         []string
}

func parseObjectClass(input string) (*ObjectClass, error) {
	re := regexp.MustCompile(`\s*\(\s*([\d\.]+)\s*` +
		`(?:NAME\s+(?:'([^']+)'|\(\s*'([^']+(?:'\s+'[^']+)*)'\s*\)))?\s*` +
		`(?:DESC\s+'([^']+)')?\s*` +
		`(?:SUP\s+(?:([\w\-]+)|\(\s*([^)]+(?:\s+'[^)]+)*)\s*\)))?\s*` +
		`(STRUCTURAL|ABSTRACT|AUXILIARY)?\s*` +
		`(?:MUST\s+\(\s*([^()]+)\s*\))?\s*` +
		`(?:MAY\s+\(\s*([^()]+)\s*\))?\s*` +
		`\)`)
	matches := re.FindStringSubmatch(input)

	if matches == nil {
		return nil, fmt.Errorf("invalid objectClass format")
	}

	id := matches[1]
	var names []string
	if matches[2] != "" {
		names = []string{matches[2]}
	} else if matches[3] != "" {
		names = strings.Fields(strings.ReplaceAll(matches[3], "'", ""))
	}
	description := matches[4]

	var superClasses []string
	if matches[5] != "" {
		superClasses = []string{matches[5]}
	} else if matches[6] != "" {
		superClasses = strings.Fields(strings.ReplaceAll(matches[6], "$", " "))
	}

	classType := matches[7]

	// Create and return struct
	c := &ObjectClass{
		Id:          id,
		Name:        names,
		Description: description,
		SuperClass:  superClasses,
		Type:        classType,
	}

	if matches[8] != "" {
		c.Must = strings.Fields(strings.ReplaceAll(matches[8], "$", " "))
	}
	if matches[9] != "" {
		c.May = strings.Fields(strings.ReplaceAll(matches[9], "$", " "))
	}

	return c, nil
}

func extractFirstMatch(re *regexp.Regexp, input string) string {
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractList(re *regexp.Regexp, input string) []string {
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		attrs := strings.Split(matches[1], " $ ")
		for i := range attrs {
			attrs[i] = strings.TrimSpace(attrs[i])
		}
		return attrs
	}
	return nil
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func isTime(s string) bool {
	_, err := time.Parse("20060102150405.000000", s)
	if err == nil {
		return true
	}
	_, err = time.Parse("20060102150405.000000Z0700", s)
	if err == nil {
		return true
	}
	_, err = time.Parse("200601021504Z0700", s)
	return err == nil
}
