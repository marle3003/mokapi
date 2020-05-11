package ldap

import (
	"bufio"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Attribute struct {
	names []string
}

func (s *Server) getSchema() {
	attrList, ok := s.root.Attributes["subSchemaSubentry"]
	if !ok || len(attrList) == 0 {
		return
	}

	dn := attrList[0]

	schemaEntry := s.getEntry(dn)
	if schemaEntry == nil {
		log.Errorf("No entry with dn %v found", dn)
	}

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
		fmt.Printf("%v", attribute)
	}
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
			parseDesc(scanner)
		case "syntax":
			parseSyntax(scanner)
		case "usage":
			parseUsage(scanner)
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

func parseDesc(scanner *bufio.Scanner) {

}

func parseSyntax(scanner *bufio.Scanner) {

}

func parseUsage(scanner *bufio.Scanner) {

}
