package parser

import (
	"fmt"
	"mokapi/providers/workflow/ast"
	"mokapi/providers/workflow/scanner"
	"mokapi/providers/workflow/token"
)

type parser struct {
	scanner *scanner.Scanner

	tok token.Token
	pos token.Position
	lit string

	errors ErrorList
}

func Parse(s string) ast.Expression {
	p := &parser{}
	eh := func(pos token.Position, msg string) { p.errors.Add(pos, msg) }
	p.scanner = scanner.NewScanner([]byte(s), eh)

	p.next()
	return p.parseBinary()
}

func (p *parser) parseBinary() ast.Expression {
	x := p.parseUnary()

	for {
		switch p.tok {
		case token.ADD, token.SUB, token.MUL, token.QUO, token.REM, token.LAND, token.LOR, token.EQL, token.LSS, token.GTR, token.NOT, token.NEQ, token.LEQ, token.GEQ:
			op := p.tok
			p.next()
			y := p.parseBinary()
			x = &ast.Binary{Lhs: x, Rhs: y, Op: op}
		default:
			return x
		}
	}
}

func (p *parser) parseUnary() ast.Expression {
	switch p.tok {
	case token.ADD, token.SUB, token.NOT:
		op := p.tok
		pos := p.pos
		p.next()
		x := p.parseUnary()
		return &ast.Unary{X: x, Op: op, OpPos: pos}
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() ast.Expression {
	o := p.parseOperand()

	for {
		switch p.tok {
		case token.PERIOD:
			p.next()
			s := p.parseIdent()
			o = &ast.Selector{X: o, Selector: s}
		case token.LPAREN:
			if ident, isIdent := o.(*ast.Identifier); !isIdent {
				p.error("unexpected token '('")
			} else {
				return p.parseCall(ident)
			}
		case token.LAMBDA:
			if ident, isIdent := o.(*ast.Identifier); !isIdent {
				p.error("unexpected token '=>'")
			} else {
				return p.parseClosure(ident)
			}
		default:
			return o
		}
	}
}

func (p *parser) parseOperand() ast.Expression {
	switch p.tok {
	case token.IDENT:
		return p.parseIdent()
	case token.STRING, token.RSTRING, token.INT, token.FLOAT:
		x := &ast.Literal{ValuePos: token.Position{}, Kind: p.tok, Value: p.lit}
		p.next()
		return x
	case token.LBRACK:
		return p.parseSequence()
	case token.LPAREN:
		lparen := p.pos
		p.next()
		x := p.parseBinary()
		p.expect(token.RPAREN)
		return &ast.ParenExpr{Lparen: lparen, X: x}
	default:
		p.error("expected operand")
		p.next()
	}
	return nil
}

func (p *parser) parseIdent() *ast.Identifier {
	x := &ast.Identifier{Name: p.lit, NamePos: p.pos}
	p.next()
	return x
}

func (p *parser) parseCall(ident *ast.Identifier) *ast.CallExpr {
	p.expect(token.LPAREN)

	var args []ast.Expression
	for p.tok != token.RPAREN && p.tok != token.EOF {
		args = append(args, p.parsePrimary())
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
	p.expect(token.RPAREN)

	return &ast.CallExpr{Fun: ident, Args: args}
}

func (p *parser) parseClosure(ident *ast.Identifier) *ast.Closure {
	closure := &ast.Closure{
		LambdaPos: p.pos,
		Args:      []*ast.Identifier{ident},
	}

	p.expect(token.LAMBDA)
	closure.Func = p.parseBinary()

	return closure
}

// TODO CHANGE MAP TO {}
func (p *parser) parseSequence() *ast.SequenceExpr {
	lbrack := p.pos
	p.expect(token.LBRACK)
	var values []ast.Expression
	isMap := false
	for p.tok != token.RBRACK && p.tok != token.EOF {
		el := p.parseElement()
		values = append(values, el)
		switch el.(type) {
		case *ast.KeyValueExpr:
			if len(values) == 1 {
				isMap = true
			} else if !isMap {
				p.expected(token.COLON)
			}
		default:
			if len(values) == 1 {
				isMap = false
			} else if isMap {
				p.expected(token.COLON)
			}
		}
		if p.tok != token.COMMA {
			break
		}
		p.expect(token.COMMA)
	}
	p.expect(token.RBRACK)
	return &ast.SequenceExpr{Values: values, Lbrack: lbrack, IsMap: isMap}
}

func (p *parser) parseElement() ast.Expression {
	x := p.parseValue()
	if p.tok == token.COLON {
		p.next()
		val := p.parseValue()
		return &ast.KeyValueExpr{Key: x, Value: val}
	} else if p.tok == token.RANGE {
		p.next()
		return &ast.RangeExpr{Start: x, End: p.parsePrimary()}
	}
	return x
}

func (p *parser) parseValue() ast.Expression {
	if p.tok == token.LBRACK {
		return p.parseSequence()
	}

	return p.parseBinary()
}

func (p *parser) next() {
	p.pos, p.tok, p.lit = p.scanner.Scan()
}

func (p *parser) expect(tok token.Token) {
	if p.tok != tok {
		p.expected(tok)
	}
	p.next()
}

func (p *parser) expected(tok token.Token) {
	msg := fmt.Sprintf("expected '%v' but found '%v'", tok.String(), p.tok.String())
	if !tok.IsKeyword() {
		msg += fmt.Sprintf(": %v", p.lit)
	}
	p.errors.Add(p.pos, msg)
}

func (p *parser) error(msg string) {
	p.errors.Add(p.pos, msg)
}
