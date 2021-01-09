package lang

import "fmt"

type Token int

const (
	ILLEGAL Token = iota
	EOF
	IDENT

	NUMBER

	operatorStart
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	LAND // &&
	LOR  // ||

	EQL // ==
	LSS // <
	GTR // >
	NOT // !

	NEQ // !=
	LEQ // <=
	GEQ // >=

	ASSIGN // =
	DEFINE // :=
	LAMBDA // =>
	operatorEnd

	LPAREN // (
	RPAREN // )

	LBRACE
	RBRACE

	LBRACK
	RBRACK

	RSTRING
	STRING

	COMMA
	PERIOD
	SEMICOLON

	COMMENT

	// Keywords
	keywordsStart
	PIPELINE
	STAGES
	STAGE
	STEPS
	WHEN
	keywordsEnd
)

var tokens = []string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	IDENT:   "IDENT",

	NUMBER: "NUMBER",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "&",

	LAND: "&&",
	LOR:  "||",

	EQL: "==",
	LSS: "<",
	GTR: ">",
	NOT: "!",

	NEQ: "!=",
	LEQ: "<=",
	GEQ: ">=",

	ASSIGN: "=",
	DEFINE: ":=",
	LAMBDA: "=>",

	LPAREN: "(",
	RPAREN: ")",

	LBRACE: "{",
	RBRACE: "}",

	LBRACK: "[",
	RBRACK: "]",

	RSTRING: "RSTRING",
	STRING:  "STRING",

	COMMA:     ",",
	PERIOD:    ".",
	SEMICOLON: ";",

	COMMENT: "//",

	PIPELINE: "pipeline",
	STAGES:   "stages",
	STAGE:    "stage",
	STEPS:    "steps",
	WHEN:     "when",
}

var keywords map[string]Token

func lockup(ident string) Token {
	if tok, b := keywords[ident]; b {
		return tok
	}
	return IDENT
}

func (t Token) isOperator() bool {
	return t > operatorStart && t < operatorEnd
}

func (t Token) String() string {
	if t >= 0 && t < Token(len(tokens)) {
		return tokens[t]
	}
	return fmt.Sprintf("token(%v)", int(t))
}

func (t Token) isKeyword() bool {
	return t > keywordsStart && t < keywordsEnd
}

func (t Token) Precedence() int {
	switch t {
	case LOR:
		return 1
	case LAND:
		return 2
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return 3
	case ADD, SUB:
		return 4
	case MUL, QUO, REM:
		return 5
	default:
		return 0
	}
}
