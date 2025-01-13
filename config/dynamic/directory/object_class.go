package directory

import (
	"fmt"
	"regexp"
	"strings"
)

type AttributeType struct {
	Id          string
	Name        []string
	Description string
	Equality    string
	Syntax      string
}

func parseAttributeType(s string) (AttributeType, error) {
	// Regular expression to match components of the definition
	regex := regexp.MustCompile(`\(\s*([\d\.]+)\s+NAME\s+((?:\([^\)]*\)|'[^']*')(?:\s+(?:'[^']*'|\([^\)]*\)))*)(?:\s+DESC\s+'([^']+)')?(?:\s+EQUALITY\s+(\S+))?(?:\s+SYNTAX\s+(\S+))?`)

	// Match the definition string
	matches := regex.FindStringSubmatch(s)
	if matches == nil {
		return AttributeType{}, fmt.Errorf("invalid attribute type definition: %s", s)
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
	attr := AttributeType{
		Id:          matches[1],
		Name:        names,
		Description: matches[3],
		Equality:    matches[4],
		Syntax:      matches[5],
	}

	return attr, nil
}
