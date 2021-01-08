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

	LPAREN // (
	RPAREN // )

	LBRACE
	RBRACE

	LBRACK
	RBRACK
	operatorEnd

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
