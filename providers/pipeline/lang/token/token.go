package token

import (
	"fmt"
)

type Token int

const (
	ILLEGAL Token = iota
	EOF

	literalStart
	IDENT

	NUMBER

	RSTRING
	STRING
	literalEnd

	operatorStart
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	ADD_ASSIGN
	SUB_ASSIGN
	MUL_ASSIGN
	QUO_ASSIGN
	REM_ASSIGN

	LAND // &&
	LOR  // ||
	INC
	DEC

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

	COMMA
	PERIOD
	SEMICOLON
	COLON

	COMMENT

	// Keywords
	keywordsStart
	PIPELINE
	STAGES
	STAGE
	STEPS
	WHEN
	VARS

	keywordsEnd
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	IDENT:   "IDENT",

	NUMBER: "NUMBER",

	RSTRING: "RSTRING",
	STRING:  "STRING",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "&",

	ADD_ASSIGN: "+=",
	SUB_ASSIGN: "-=",
	MUL_ASSIGN: "*=",
	QUO_ASSIGN: "/=",
	REM_ASSIGN: "%=",

	LAND: "&&",
	LOR:  "||",
	INC:  "++",
	DEC:  "--",

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

	COMMA:     ",",
	PERIOD:    ".",
	SEMICOLON: ";",
	COLON:     ":",

	COMMENT: "//",

	PIPELINE: "pipeline",
	STAGES:   "stages",
	STAGE:    "stage",
	STEPS:    "steps",
	WHEN:     "when",
	VARS:     "vars",
}

var Keywords map[string]Token

func init() {
	Keywords = make(map[string]Token)
	for i := keywordsStart + 1; i < keywordsEnd; i++ {
		Keywords[tokens[i]] = i
	}
}

func Loockup(ident string) Token {
	if tok, b := Keywords[ident]; b {
		return tok
	}
	return IDENT
}

func (t Token) IsLiteral() bool {
	return t > literalStart && t < literalEnd
}

func (t Token) IsOperator() bool {
	return t > operatorStart && t < operatorEnd
}

func (t Token) String() string {
	if t >= 0 && t < Token(len(tokens)) {
		return tokens[t]
	}
	return fmt.Sprintf("token(%v)", int(t))
}

func (t Token) IsKeyword() bool {
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

func (t Token) IsExprEnd() bool {
	return t == SEMICOLON || t == EOF || t == RPAREN || t == RBRACE || t == RBRACK
}
