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

	INT
	FLOAT

	RSTRING
	STRING
	literalEnd

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

	LAMBDA // =>
	RANGE  // ..
	operatorEnd

	LPAREN // (
	RPAREN // )

	LBRACE // {
	RBRACE // }

	LBRACK // [
	RBRACK // ]

	COMMA
	PERIOD
	COLON

	// Keywords
	keywordsStart

	keywordsEnd
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	IDENT:   "IDENT",

	INT:   "INT",
	FLOAT: "FLOAT",

	RSTRING: "RSTRING",
	STRING:  "STRING",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",

	LAND: "&&",
	LOR:  "||",

	EQL: "==",
	LSS: "<",
	GTR: ">",
	NOT: "!",

	NEQ: "!=",
	LEQ: "<=",
	GEQ: ">=",

	LAMBDA: "=>",
	RANGE:  "..",

	LPAREN: "(",
	RPAREN: ")",

	LBRACE: "{",
	RBRACE: "}",

	LBRACK: "[",
	RBRACK: "]",

	COMMA:  ",",
	PERIOD: ".",
	COLON:  ":",
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
