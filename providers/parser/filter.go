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

type ExpressionTag int

const (
	Undefined ExpressionTag = -1
	Or        ExpressionTag = 0
	And       ExpressionTag = 1
	//FilterNot            FilterTag = 2
	EqualityMatch  ExpressionTag = 3
	Greater        ExpressionTag = 4
	GreaterOrEqual ExpressionTag = 5
	Less           ExpressionTag = 6
	LessOrEqual    ExpressionTag = 7
	Like           ExpressionTag = 8

	Parameter ExpressionTag = 9
	Constant  ExpressionTag = 10
	Body      ExpressionTag = 11
	Property  ExpressionTag = 12
)

type Expression struct {
	Tag      ExpressionTag
	Children []*Expression
	Value    string
}

func ParseExpression(s string) (*Expression, error) {
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

func parseCondition(scanner *bufio.Scanner) (*Expression, error) {
	expr := &Expression{Tag: Or, Children: make([]*Expression, 0)}

	term, error := parseTerm(scanner)
	if error != nil {
		return nil, error
	}
	expr.Children = append(expr.Children, term)

	for strings.ToLower(scanner.Text()) == "or" {
		term, error := parseTerm(scanner)
		if error != nil {
			return nil, error
		}
		expr.Children = append(expr.Children, term)

		scanner.Scan()
	}

	if len(expr.Children) == 0 {
		return nil, fmt.Errorf("No expression found")
	}

	if len(expr.Children) == 1 {
		return expr.Children[0], nil
	}

	return expr, nil
}

func parseTerm(scanner *bufio.Scanner) (*Expression, error) {
	expr := &Expression{Tag: And, Children: make([]*Expression, 0)}

	term, error := parsePrimary(scanner)
	if error != nil {
		return nil, error
	}
	expr.Children = append(expr.Children, term)

	if !scanner.Scan() {
		return term, nil
	}

	for gatter := scanner.Text(); len(gatter) > 0; {
		if strings.ToLower(gatter) != "and" {
			return nil, fmt.Errorf("Expected 'and' operator but found '%v'", gatter)
		}
		term, error := parseTerm(scanner)
		if error != nil {
			return nil, error
		}
		expr.Children = append(expr.Children, term)

		scanner.Scan()
	}

	if len(expr.Children) == 0 {
		return nil, fmt.Errorf("No expression found")
	}

	if len(expr.Children) == 1 {
		return expr.Children[0], nil
	}

	return expr, nil
}

func parsePrimary(scanner *bufio.Scanner) (*Expression, error) {
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

	return &Expression{Tag: operator, Children: []*Expression{identifierLeft, identifierRight}}, nil
}

func parseOperator(scanner *bufio.Scanner) (ExpressionTag, error) {
	if !scanner.Scan() {
		return Undefined, fmt.Errorf("Syntax error: Expected identifier")
	}

	text := scanner.Text()

	switch strings.ToLower(text) {
	case "=":
		return EqualityMatch, nil
	case "like":
		return Like, nil
	}

	return Undefined, fmt.Errorf("Unsupported operator %v", text)
}

func parseFactor(scanner *bufio.Scanner) (*Expression, error) {
	if !scanner.Scan() {
		return nil, fmt.Errorf("Syntax error: Expected identifier")
	}

	text := scanner.Text()
	exp := &Expression{}

	// string constant
	if strings.HasPrefix(text, "\"") {
		exp.Value = text[1:]
		exp.Tag = Constant
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
		exp.Tag = Constant
	} else {
		paramRegex := regexp.MustCompile(`param\["(?P<name>.+)"\]`)
		match := paramRegex.FindStringSubmatch(text)
		if len(match) > 1 {
			exp.Value = match[1]
			exp.Tag = Parameter
		} else {
			bodyRegex := regexp.MustCompile(`body\["(?P<name>.+)"\]`)
			match := bodyRegex.FindStringSubmatch(text)
			if len(match) > 1 {
				exp.Value = match[1]
				exp.Tag = Body
			} else {
				exp.Value = text
				exp.Tag = Property
			}
		}
	}

	return exp, nil
}

func (exp *Expression) String() string {
	switch exp.Tag {
	case Or:
		s := ""
		for _, i := range exp.Children {
			if s != "" {
				s += " OR "
			}
			s += i.String()
		}
		return s
	case And:
		s := ""
		for _, i := range exp.Children {
			if s != "" {
				s += " AND "
			}
			s += i.String()
		}
		return s
	case EqualityMatch:
		return fmt.Sprintf("%v = %v", exp.Children[0].String(), exp.Children[1].String())
	case Greater:
		return fmt.Sprintf("%v > %v", exp.Children[0].String(), exp.Children[1].String())
	case GreaterOrEqual:
		return fmt.Sprintf("%v >= %v", exp.Children[0].String(), exp.Children[1].String())
	case Less:
		return fmt.Sprintf("%v < %v", exp.Children[0].String(), exp.Children[1].String())
	case LessOrEqual:
		return fmt.Sprintf("%v <= %v", exp.Children[0].String(), exp.Children[1].String())
	case Like:
		return fmt.Sprintf("%v LIKE %v", exp.Children[0].String(), exp.Children[1].String())
	case Property:
		return exp.Value
	case Parameter:
		return fmt.Sprintf(`param["%v"]`, exp.Value)
	case Constant:
		return exp.Value
	case Body:
		return fmt.Sprintf(`body["%v"]`, exp.Value)
	}

	return ""
}

func (exp *Expression) IsTrue(resolveFactor func(factor string, tag ExpressionTag) string) (bool, error) {
	switch exp.Tag {
	case EqualityMatch:
		left := resolveFactor(exp.Children[0].Value, exp.Children[0].Tag)
		right := resolveFactor(exp.Children[1].Value, exp.Children[1].Tag)

		return left == right, nil
	case Like:
		left := resolveFactor(exp.Children[0].Value, exp.Children[0].Tag)
		right := resolveFactor(exp.Children[1].Value, exp.Children[1].Tag)

		s := strings.ReplaceAll(right, "%", ".*")
		regex := regexp.MustCompile(s)
		match := regex.FindStringSubmatch(left)

		return len(match) > 0, nil
	}

	return false, fmt.Errorf("Unsupported expression tag %v", exp.Tag)
}
