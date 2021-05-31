package scanner

import (
	"mokapi/providers/workflow/token"
	"testing"
)

const /* class */ (
	special = iota
	literal
	operator
	keyword
)

func tokenClass(tok token.Token) int {
	switch {
	case tok.IsLiteral():
		return literal
	case tok.IsOperator():
		return operator
	case tok.IsKeyword():
		return keyword
	}
	return special
}

type data struct {
	tok   token.Token
	lit   string
	class int
}

var tokens = [...]data{
	{token.IDENT, "foobar", literal},
	{token.IDENT, "foo_bar", literal},
	{token.IDENT, "foobar12", literal},
	{token.INT, "3", literal},
	{token.FLOAT, "3.141", literal},
	{token.RSTRING, "'foobar'", literal},
	{token.STRING, "\"foobar\"", literal},

	{token.ADD, "+", operator},
	{token.SUB, "-", operator},
	{token.MUL, "*", operator},
	{token.QUO, "/", operator},
	{token.REM, "%", operator},

	{token.LAND, "&&", operator},
	{token.LOR, "||", operator},

	{token.EQL, "==", operator},
	{token.LSS, "<", operator},
	{token.GTR, ">", operator},
	{token.NOT, "!", operator},

	{token.NEQ, "!=", operator},
	{token.LEQ, "<=", operator},
	{token.GEQ, ">=", operator},

	{token.LAMBDA, "=>", operator},
	{token.RANGE, "..", operator},

	{token.LPAREN, "(", special},
	{token.RPAREN, ")", special},

	{token.LBRACE, "{", special},
	{token.RBRACE, "}", special},

	{token.LBRACK, "[", special},
	{token.RBRACK, "]", special},

	{token.COMMA, ",", special},
	{token.PERIOD, ".", special},
	{token.COLON, ":", special},
}

const whitespace = "  \t \n\n"

var source = func() []byte {
	var src []byte
	for _, t := range tokens {
		src = append(src, t.lit...)
		src = append(src, whitespace...)
	}
	return src
}()

func TestScan(t *testing.T) {
	eh := func(_ token.Position, msg string) {
		t.Errorf("error handler called with msg %v", msg)
	}

	s := NewScanner(source, eh)

	index := 0
	for {
		_, tok, lit := s.Scan()

		expect := data{token.EOF, "", special}
		if index < len(tokens) {
			expect = tokens[index]
			index++
		}

		if tok != expect.tok {
			t.Errorf("bad token for %q: got %v, expected %v", lit, tok, expect.tok)
		}
		if tokenClass(tok) != expect.class {
			t.Errorf("bad class for %q: got %v, expected %v", lit, tokenClass(tok), expect.class)
		}

		expectLit := ""
		if expect.tok.IsLiteral() {
			expectLit = expect.lit
			if expectLit[0] == '\'' || expectLit[0] == '"' {
				expectLit = expectLit[1 : len(expectLit)-1]
			}
		} else if expect.tok.IsKeyword() {
			expectLit = expect.lit
		}

		if lit != expectLit {
			t.Errorf("bad literal for %q: got %v, expected %v", lit, lit, expectLit)
		}

		if tok == token.EOF {
			break
		}
	}
}
