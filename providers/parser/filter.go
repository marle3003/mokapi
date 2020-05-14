package parser

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

/*
	EBNF

	condition = term { 'OR' term } ;
	term = primary { 'AND' primary } ;
	primary = factor ( comparison_operator) factor ;
	factor = identifier | parameter | body | constant ;
	parameter = 'param["', identifier, '"]' ;
	body = 'body["', all_characters, '"]' ;
	identifier = alphabetic_character, { alphabetic_character | digit } ;
	constant =  number | string ;
	number = [ '-' ], digit, { digit } ;
	string = '"', { all_characters - '"' }, '"' ;
	comparison_operator = '=' | '!=' | '>' | '<' | '>=' | '<=' | 'like' ;
	alphabetic_character = "A" | "B" | "C" | "D" | "E" | "F" | "G"
						| "H" | "I" | "J" | "K" | "L" | "M" | "N"
						| "O" | "P" | "Q" | "R" | "S" | "T" | "U"
						| "V" | "W" | "X" | "Y" | "Z"
						| "a" | "b" | "c" | "d" | "e" | "f" | "g"
						| "h" | "i" | "j" | "k" | "l" | "m" | "n"
						| "o" | "p" | "q" | "r" | "s" | "t" | "u"
						| "v" | "w" | "x" | "y" | "z"  ;
	digit = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;
	all_characters = ? all visible characters ? ;
*/

type FilterTag int

const (
	FilterUndefined FilterTag = -1
	FilterOr        FilterTag = 0
	FilterAnd       FilterTag = 1
	//FilterNot            FilterTag = 2
	FilterEqualityMatch  FilterTag = 3
	FilterGreater        FilterTag = 4
	FilterGreaterOrEqual FilterTag = 5
	FilterLess           FilterTag = 6
	FilterLessOrEqual    FilterTag = 7
	FilterLike           FilterTag = 8

	FilterParameter FilterTag = 9
	FilterConstant  FilterTag = 10
	FilterBody      FilterTag = 11
	FilterProperty  FilterTag = 12
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
	identifierLeft, error := parseFactor(scanner)
	if error != nil {
		return nil, error
	}

	operator, error := parseOperator(scanner)
	if error != nil {
		return nil, error
	}

	identifierRight, error := parseFactor(scanner)
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

func parseFactor(scanner *bufio.Scanner) (*FilterExp, error) {
	if !scanner.Scan() {
		return nil, fmt.Errorf("Syntax error: Expected identifier")
	}

	text := scanner.Text()
	exp := &FilterExp{}

	// string constant
	if strings.HasPrefix(text, "\"") {
		exp.Value = text[1:]
		exp.Tag = FilterConstant
		if !strings.HasSuffix(text, "\"") {
			for scanner.Scan() {
				text = scanner.Text()

				if strings.HasSuffix(text, "\"") {
					exp.Value += text[:len(text)-1] // remove " at the end
					break
				}
				exp.Value += text
			}
		} else {
			exp.Value = exp.Value[:len(exp.Value)-1] // remove " at the end
		}

	} else if unicode.IsDigit(rune(text[0])) || text[0] == '-' { // number
		exp.Value = text
		exp.Tag = FilterConstant
	} else {
		paramRegex := regexp.MustCompile(`param\["(?P<name>.+)"\]`)
		match := paramRegex.FindStringSubmatch(text)
		if len(match) > 1 {
			exp.Value = match[1]
			exp.Tag = FilterParameter
		} else {
			bodyRegex := regexp.MustCompile(`body\["(?P<name>.+)"\]`)
			match := bodyRegex.FindStringSubmatch(text)
			if len(match) > 1 {
				exp.Value = match[1]
				exp.Tag = FilterBody
			} else {
				exp.Value = text
				exp.Tag = FilterProperty
			}
		}
	}

	return exp, nil
}

func (exp *FilterExp) String() string {
	switch exp.Tag {
	case FilterOr:
		s := ""
		for _, i := range exp.Children {
			if s != "" {
				s += " OR "
			}
			s += i.String()
		}
		return s
	case FilterAnd:
		s := ""
		for _, i := range exp.Children {
			if s != "" {
				s += " AND "
			}
			s += i.String()
		}
		return s
	case FilterEqualityMatch:
		return fmt.Sprintf("%v = %v", exp.Children[0].String(), exp.Children[1].String())
	case FilterGreater:
		return fmt.Sprintf("%v > %v", exp.Children[0].String(), exp.Children[1].String())
	case FilterGreaterOrEqual:
		return fmt.Sprintf("%v >= %v", exp.Children[0].String(), exp.Children[1].String())
	case FilterLess:
		return fmt.Sprintf("%v < %v", exp.Children[0].String(), exp.Children[1].String())
	case FilterLessOrEqual:
		return fmt.Sprintf("%v <= %v", exp.Children[0].String(), exp.Children[1].String())
	case FilterLike:
		return fmt.Sprintf("%v LIKE %v", exp.Children[0].String(), exp.Children[1].String())
	case FilterProperty:
		return exp.Value
	case FilterParameter:
		return fmt.Sprintf("param[\"%v\"]", exp.Value)
	case FilterConstant:
		return exp.Value
	case FilterBody:
		return fmt.Sprintf("body[\"%v\"]", exp.Value)
	}

	return ""
}

func (exp *FilterExp) IsTrue(resolveFactor func(factor string, tag FilterTag) string) (bool, error) {
	switch exp.Tag {
	case FilterEqualityMatch:
		left := resolveFactor(exp.Children[0].Value, exp.Children[0].Tag)
		right := resolveFactor(exp.Children[1].Value, exp.Children[1].Tag)

		return left == right, nil
	case FilterLike:
		left := resolveFactor(exp.Children[0].Value, exp.Children[0].Tag)
		right := resolveFactor(exp.Children[1].Value, exp.Children[1].Tag)

		s := strings.ReplaceAll(right, "%", ".*")
		regex := regexp.MustCompile(s)
		match := regex.FindStringSubmatch(left)

		return len(match) > 0, nil
	}

	return false, fmt.Errorf("Unsupported filter tag %v", exp.Tag)
}
