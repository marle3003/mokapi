package parser

import (
	"fmt"
	"mokapi/providers/pipeline/lang/ast"
	scanner2 "mokapi/providers/pipeline/lang/scanner"
	"mokapi/providers/pipeline/lang/token"
	"mokapi/providers/pipeline/lang/types"
)

type parser struct {
	errors  ErrorList
	scanner *scanner2.Scanner
	scope   *ast.Scope

	tok token.Token
	pos scanner2.Position
	lit string
}

func ParseFile(src []byte, scope *ast.Scope) (f *ast.File, err error) {
	parser := &parser{
		errors:  ErrorList{},
		scanner: scanner2.NewScanner(src, nil),
		scope:   scope,
	}
	parser.next()

	defer func() {
		err = parser.errors.Err()
	}()

	f = parser.parseFile()

	return
}

func ParseExpr(b []byte, scope *ast.Scope) (expr ast.Expression, err error) {
	parser := &parser{
		errors:  ErrorList{},
		scanner: scanner2.NewScanner(b, nil),
		scope:   scope,
	}
	parser.scanner.InsertLineEnd = true
	parser.next()

	defer func() {
		err = parser.errors.Err()
	}()

	expr = parser.parseBinary()

	return
}

func (p *parser) parseFile() (f *ast.File) {
	f = &ast.File{Scope: p.scope}
	for p.tok != token.EOF {
		switch p.tok {
		case token.PIPELINE:
			pipeline := p.parsePipeline()
			f.AddPipeline(pipeline)
		default:
			p.expectedError(token.PIPELINE)
			p.advance(token.PIPELINE)
		}
	}
	return
}

func (p *parser) parsePipeline() *ast.Pipeline {
	pipeline := &ast.Pipeline{}
	p.expect(token.PIPELINE)
	p.expect(token.LPAREN)
	pipeline.Name = p.parseName()
	p.expect(token.RPAREN)
	p.parsePipelineBody(pipeline)
	return pipeline
}

func (p *parser) parsePipelineBody(pipeline *ast.Pipeline) {
	p.expect(token.LBRACE)
	if p.tok == token.STAGES {
		pipeline.Stages = p.parseStages()
	}
	p.expect(token.RBRACE)
}

func (p *parser) parseStages() []*ast.Stage {
	stages := make([]*ast.Stage, 0)
	p.expect(token.STAGES)
	p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		switch p.tok {
		case token.STAGE:
			stages = append(stages, p.parseStage())
		default:
			p.expectedError(token.STAGE)
			p.advance(token.STAGE)
		}
	}
	p.expect(token.RBRACE)
	return stages
}

func (p *parser) parseStage() *ast.Stage {
	s := &ast.Stage{}
	p.expect(token.STAGE)
	p.expect(token.LPAREN)
	s.Name = p.parseName()
	p.expect(token.RPAREN)
	p.expect(token.LBRACE)

	p.openScope()
	defer p.closeScope()
	s.Scope = p.scope

	for p.tok != token.RBRACE && p.tok != token.EOF {
		switch p.tok {
		case token.STEPS:
			if s.Steps != nil {
				p.error("redefine steps block")
			}
			s.Steps = p.parseSteps()
		case token.WHEN:
			if s.When != nil {
				p.error("redefine steps block")
			}
			s.When = p.parseWhen()
		default:
			p.error("expected steps or when")
			p.advance(token.STEPS, token.WHEN)
		}
	}

	p.expect(token.RBRACE)
	return s
}

func (p *parser) parseWhen() (expr *ast.ExprStatement) {
	p.expect(token.WHEN)
	p.expect(token.LBRACE)
	p.scanner.UseLineEnd(true)
	x := p.parseBinary()
	expr = &ast.ExprStatement{X: x}
	p.expect(token.SEMICOLON)
	p.scanner.UseLineEnd(false)
	p.expect(token.RBRACE)
	return
}

func (p *parser) parseSteps() *ast.StepBlock {
	p.expect(token.STEPS)
	p.expect(token.LBRACE)
	p.scanner.UseLineEnd(true)
	var list []ast.Statement
	for {
		if p.tok == token.SEMICOLON {
			p.next()
			continue
		}
		if p.tok == token.RBRACE || p.tok == token.EOF {
			break
		}
		list = append(list, p.parseStatement())
	}
	p.scanner.UseLineEnd(false)
	p.expect(token.RBRACE)
	return &ast.StepBlock{Statments: list}
}

func (p *parser) parseStatement() (stmt ast.Statement) {
	if p.tok == token.VAR {
		stmt = p.parseVarDecl()
	} else {
		lhs := p.parseBinary()
		switch p.tok {
		case token.DEFINE, token.ASSIGN, token.ADD_ASSIGN, token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN, token.REM_ASSIGN:
			assignTok := p.tok
			p.next()
			rhs := p.parseBinary()
			stmt = &ast.Assignment{Lhs: lhs, Tok: assignTok, Rhs: rhs}
			if assignTok == token.DEFINE {
				// we do not know the type of the rhs expression
				p.varDecl(lhs, "")
			}
		default:
			stmt = &ast.ExprStatement{X: lhs}
		}
	}
	p.expectStatmentEnd()
	return
}

func (p *parser) varDecl(x ast.Expression, typeName string) {
	if ident, isIdent := x.(*ast.Ident); isIdent {
		if _, identExists := p.scope.Symbol(ident.Name); identExists {
			p.error(fmt.Sprintf("identifier '%v' already defined", identExists))
		}

		switch t := typeName; {
		case t == "Expando":
			p.scope.SetSymbol(ident.Name, types.NewExpando())
		case len(t) == 0:
			p.scope.SetSymbol(ident.Name, nil)
		default:
			p.error(fmt.Sprintf("unexpected type %v", t))
		}

	} else {
		p.error("identifier on left side of := expected")
	}
}

func (p *parser) parseVarDecl() *ast.DeclStmt {
	p.expect(token.VAR)
	if p.tok != token.IDENT {
		p.expectedError(token.IDENT)
	}
	ident := &ast.Ident{Name: p.lit}
	p.next()
	if p.tok != token.IDENT {
		p.expectedError(token.IDENT)
	}
	typeName := p.lit
	p.next()

	p.varDecl(ident, typeName)

	return &ast.DeclStmt{Name: ident, Type: typeName}
}

func (p *parser) parseBinary() ast.Expression {
	x := p.parseUnary()

	// todo: consider correct operator order () before */ before +-...
	for {
		switch p.tok {
		case token.ADD, token.SUB, token.MUL, token.QUO, token.REM, token.LAND, token.LOR, token.EQL, token.LSS, token.GTR, token.NOT, token.NEQ, token.LEQ, token.GEQ:
			op := p.tok
			p.next()
			y := p.parseBinary()
			x = &ast.Binary{Lhs: x, Rhs: y, Op: op, Precedence: op.Precedence()}
		default:
			return x
		}
	}

	return x
}

func (p *parser) parseUnary() ast.Expression {
	switch p.tok {
	case token.NOT:
		op := p.tok
		p.next()
		x := p.parseUnary()
		return &ast.Unary{X: x, Op: op}
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() ast.Expression {
	operand := p.parseOperand(true)

	if p.tok == token.PERIOD {
		p.resolve(operand)
		path := &ast.PathExpr{X: operand}
		operand = p.parsePath(path)
	} else if !p.tok.IsExprEnd() && !p.tok.IsOperator() && p.tok != token.COLON && p.tok != token.COMMA {
		operand = p.parseCall(operand)
	}

	return operand
}

func (p *parser) parsePath(path *ast.PathExpr) ast.Expression {
	current := path
	for {

		if p.tok == token.LBRACK {
			// if index used like findAll...}[0]
			current.Path = p.parseIndex()
		} else {
			p.next()
			current.Path = p.parsePathOperand()
		}

		if p.tok == token.LBRACK {
			// if index used like list[0]
			current = &ast.PathExpr{X: current}
			current.Path = p.parseIndex()
		} else if !p.tok.IsExprEnd() && p.tok != token.PERIOD && !p.tok.IsOperator() {
			c := p.parseCall(current.Path)
			current.Path = c.Func
			current.Args = c.Args
		}

		if p.tok != token.PERIOD && p.tok != token.LBRACK {
			return current
		}
		current = &ast.PathExpr{X: current}
	}
}

func (p *parser) parsePathOperand() (x ast.Expression) {
	switch p.tok {
	case token.IDENT:
		x = &ast.Ident{Name: p.lit}
		p.next()
	case token.STRING, token.RSTRING:
		x = &ast.Literal{Kind: p.tok, Value: p.lit}
		p.next()
	default:
		p.error("expected operand")
		p.next()
	}
	return
}

func (p *parser) parseIndex() (x ast.Expression) {
	p.expect(token.LBRACK)
	switch p.tok {
	case token.STRING, token.RSTRING:
		x = &ast.Literal{Kind: p.tok, Value: p.lit}
		p.next()
	case token.NUMBER:
		x = &ast.Literal{Kind: p.tok, Value: p.lit}
		p.next()
	}
	p.expect(token.RBRACK)
	return
}

func (p *parser) parseCall(f ast.Expression) *ast.Call {
	list := p.parseArgList()
	return &ast.Call{Args: list, Func: f}
}

func (p *parser) parseArgList() []*ast.Argument {
	var list []*ast.Argument
	for !p.tok.IsExprEnd() {
		list = append(list, p.parseArgument())
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
	return list
}

func (p *parser) parseOperand(lhs bool) (x ast.Expression) {
	switch p.tok {
	case token.IDENT:
		x = &ast.Ident{Name: p.lit}
		if !lhs {
			p.resolve(x)
		}
		p.next()
	case token.STRING, token.RSTRING:
		x = &ast.Literal{Kind: p.tok, Value: p.lit}
		p.next()
	case token.NUMBER:
		x = &ast.Literal{Kind: p.tok, Value: p.lit}
		p.next()
	case token.LPAREN:
		p.next()
		x = &ast.ParenExpr{X: p.parseBinary()}
		p.expect(token.RPAREN)
	default:
		p.error("expected operand")
		p.next()
	}
	return
}

func (p *parser) parseClosure() *ast.Closure {
	p.expect(token.LBRACE)
	closure := &ast.Closure{Block: &ast.Block{}}
	inInput := false
	p.openScope()
	for p.tok != token.RBRACE && p.tok != token.SEMICOLON && p.tok != token.EOF {
		x := p.parseBinary()
		if p.tok == token.COMMA || p.tok == token.LAMBDA {
			inInput = p.tok == token.COMMA
			ident := &ast.Ident{Name: x.(*ast.Ident).Name}
			closure.Params = append(closure.Params, ident)
			p.varDecl(ident, "")
			p.next()
		} else {
			closure.Block.Stmts = append(closure.Block.Stmts, &ast.ExprStatement{X: x})
		}
	}
	if inInput {
		p.expectedError(token.LAMBDA)
	}
	p.expect(token.RBRACE)
	p.closeScope()
	return closure
}

func (p *parser) parseArgument() *ast.Argument {
	arg := &ast.Argument{}

	var expr ast.Expression
	if p.tok == token.LBRACE {
		expr = p.parseClosure()
	} else {
		expr = p.parseBinary()
	}

	if p.tok == token.COLON {
		p.expect(token.COLON)
		arg.Value = p.parseBinary()

		arg.Name = expr.(*ast.Ident).Name
	} else {
		arg.Name = ""
		arg.Value = expr
	}

	return arg
}

func (p *parser) parseName() (name string) {
	if p.tok == token.STRING || p.tok == token.RSTRING {
		name = p.lit
		p.next()
	} else {
		name = ""
	}
	return
}

func (p *parser) next() {
	p.pos, p.tok, p.lit = p.scanner.Scan()
}

func (p *parser) expect(tok token.Token) {
	if p.tok != tok {
		p.expectedError(tok)
	}
	p.next()
}

func (p *parser) expectStatmentEnd() {
	if p.tok != token.SEMICOLON && p.tok != token.EOF {
		p.expectedError(token.SEMICOLON)
	}
	p.next()
}

func (p *parser) expectedError(tok token.Token) {
	msg := fmt.Sprintf("expected '%v' but found '%v'", tok.String(), p.tok.String())
	if !tok.IsKeyword() {
		msg += fmt.Sprintf(": %v", p.lit)
	}
	p.errors.Add(p.pos, msg)
}

func (p *parser) error(msg string) {
	p.errors.Add(p.pos, msg)
}

func (p *parser) advance(tok ...token.Token) {
	tokens := make(map[token.Token]bool)
	for _, t := range tok {
		tokens[t] = true
	}
	p.advanceOneOf(tokens)
}

func (p *parser) advanceOneOf(to map[token.Token]bool) {
	for p.tok != token.EOF {
		if _, ok := to[p.tok]; ok {
			return
		}
		p.next()
	}
}

func (p *parser) resolve(x ast.Expression) {
	if ident, isIdent := x.(*ast.Ident); isIdent {
		if _, exists := p.scope.Symbol(ident.Name); !exists {
			p.error(fmt.Sprintf("identifier '%v' does not exists", ident.Name))
		}
	}
}

func (p *parser) openScope() {
	p.scope = ast.OpenScope(p.scope)
}

func (p *parser) closeScope() {
	p.scope = p.scope.Outer
}
