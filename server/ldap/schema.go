package ldap

import (
	"bufio"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Schema struct {
	attributes []*Attribute
}

type Attribute struct {
	names       []string
	description string
	syntax      string
	usage       string
}

func (s *Schema) getAttribute(name string) *Attribute {
	for _, a := range s.attributes {
		for _, n := range a.names {
			if n == name {
				return a
			}
		}
	}
	return nil
}

func (s *Server) getSchema() (*Schema, error) {
	attrList, ok := s.root.Attributes["subSchemaSubentry"]
	if !ok || len(attrList) == 0 {
		return nil, nil
	}

	dn := attrList[0]

	schemaEntry := s.getEntry(dn)
	if schemaEntry == nil {
		return nil, fmt.Errorf("No entry with dn %v found", dn)
	}

	schema := &Schema{attributes: make([]*Attribute, 0)}

	attributeTypes, ok := schemaEntry.Attributes["attributeTypes"]
	if !ok || len(attributeTypes) == 0 {
		log.Errorf("No attribute 'attributeTypes' found in schema")
	}

	for _, attr := range attributeTypes {
		attribute, error := parseABNF(attr)
		if error != nil {
			log.Error(error.Error)
			continue
		}
		schema.attributes = append(schema.attributes, &attribute)
	}

	return schema, nil
}

func parseABNF(s string) (Attribute, error) {
	codeReader := strings.NewReader(s)

	scanner := bufio.NewScanner(codeReader)
	scanner.Split(bufio.ScanWords)

	var attribute Attribute

	for scanner.Scan() {
		text := scanner.Text()
		switch strings.ToLower(text) {
		case "(":
		case ")":
			return attribute, nil
		case "name":
			names, error := parseName(scanner)
			if error != nil {
				return attribute, error
			}
			attribute.names = names
		case "desc":
			attribute.description = parseDesc(scanner)
		case "syntax":
			attribute.syntax = parseSyntax(scanner)
		case "usage":
			attribute.usage = parseUsage(scanner)
		}
	}

	return attribute, fmt.Errorf("Syntax error missing closing ')'")
}

func parseName(scanner *bufio.Scanner) ([]string, error) {
	names := make([]string, 0)
	for scanner.Scan() {
		text := scanner.Text()
		switch strings.ToLower(text) {
		case "(":
		case ")":
			return names, nil
		default:
			names = append(names, strings.Trim(text, "'"))
		}
	}

	return nil, fmt.Errorf("Syntax error in NAME")
}

func parseDesc(scanner *bufio.Scanner) string {
	d := ""
	for scanner.Scan() {
		text := scanner.Text()

		if strings.HasPrefix(text, "'") {
			d += text[1:]
		} else if strings.HasSuffix(text, "'") {
			last := len(text) - 1
			d += text[:last]
			break
		} else {
			d += " " + text
		}
	}

	return d
}

func parseSyntax(scanner *bufio.Scanner) string {
	if scanner.Scan() {
		t := scanner.Text()
		return t
	}

	return ""
}

func parseUsage(scanner *bufio.Scanner) string {
	if scanner.Scan() {
		t := scanner.Text()
		return t
	}

	return ""
}
