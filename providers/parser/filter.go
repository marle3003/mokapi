package parser

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type FilterTag int

const (
	FilterUndefined      FilterTag = -1
	FilterOr             FilterTag = 0
	FilterAnd            FilterTag = 1
	FilterNot            FilterTag = 2
	FilterEqualityMatch  FilterTag = 3
	FilterGreater        FilterTag = 4
	FilterGreaterOrEqual FilterTag = 5
	FilterLess           FilterTag = 6
	FilterLessOrEqual    FilterTag = 7
	FilterLike           FilterTag = 8
	FilterProperty       FilterTag = 9
	FilterParameter      FilterTag = 10
)

type FilterExp struct {
	Tag      FilterTag
	Children []*FilterExp
	Value    string
}

func ParseFilter(s string) (*FilterExp, error) {
	if len(s) == 0 {
		return nil, nil
	}

	codeReader := strings.NewReader(s)

	scanner := bufio.NewScanner(codeReader)
	scanner.Split(bufio.ScanWords)

	filter, error := parseCondition(scanner)
	if error != nil {
		return nil, error
	}

	return filter, nil
}

func parseCondition(scanner *bufio.Scanner) (*FilterExp, error) {
	filter := &FilterExp{Tag: FilterOr, Children: make([]*FilterExp, 0)}

	term, error := parseTerm(scanner)
	if error != nil {
		return nil, error
	}
	filter.Children = append(filter.Children, term)

	for strings.ToLower(scanner.Text()) == "or" {
		term, error := parseTerm(scanner)
		if error != nil {
			return nil, error
		}
		filter.Children = append(filter.Children, term)

		scanner.Scan()
	}

	if len(filter.Children) == 0 {
		return nil, fmt.Errorf("No expression found")
	}

	if len(filter.Children) == 1 {
		return filter.Children[0], nil
	}

	return filter, nil
}

func parseTerm(scanner *bufio.Scanner) (*FilterExp, error) {
	filter := &FilterExp{Tag: FilterAnd, Children: make([]*FilterExp, 0)}

	term, error := parsePrimary(scanner)
	if error != nil {
		return nil, error
	}
	filter.Children = append(filter.Children, term)

	if !scanner.Scan() {
		return term, nil
	}

	for strings.ToLower(scanner.Text()) == "and" {
		term, error := parseTerm(scanner)
		if error != nil {
			return nil, error
		}
		filter.Children = append(filter.Children, term)

		scanner.Scan()
	}

	if len(filter.Children) == 0 {
		return nil, fmt.Errorf("No expression found")
	}

	if len(filter.Children) == 1 {
		return filter.Children[0], nil
	}

	return filter, nil
}

func parsePrimary(scanner *bufio.Scanner) (*FilterExp, error) {
	identifierLeft, error := parseIdentifier(scanner)
	if error != nil {
		return nil, error
	}

	operator, error := parseOperator(scanner)
	if error != nil {
		return nil, error
	}

	identifierRight, error := parseIdentifier(scanner)
	if error != nil {
		return nil, error
	}

	return &FilterExp{Tag: operator, Children: []*FilterExp{identifierLeft, identifierRight}}, nil
}

func parseOperator(scanner *bufio.Scanner) (FilterTag, error) {
	if !scanner.Scan() {
		return FilterUndefined, fmt.Errorf("Syntax error: Expected identifier")
	}

	text := scanner.Text()

	switch strings.ToLower(text) {
	case "=":
		return FilterEqualityMatch, nil
	case "like":
		return FilterLike, nil
	}

	return FilterUndefined, fmt.Errorf("Unsupported operator %v", text)
}

func parseIdentifier(scanner *bufio.Scanner) (*FilterExp, error) {
	if !scanner.Scan() {
		return nil, fmt.Errorf("Syntax error: Expected identifier")
	}

	exp := &FilterExp{}

	text := scanner.Text()
	paramRegex := regexp.MustCompile(`param\[(?P<name>.+)\]`)
	match := paramRegex.FindStringSubmatch(text)
	if len(match) > 1 {
		exp.Value = match[1]
		exp.Tag = FilterParameter
	} else {
		exp.Value = text
		exp.Tag = FilterProperty
	}

	return exp, nil
}
