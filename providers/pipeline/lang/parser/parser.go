package parser

import (
	"fmt"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/scanner"
	"mokapi/providers/pipeline/lang/token"
	"mokapi/providers/pipeline/lang/types"
)

type parser struct {
	errors  ErrorList
	scanner *scanner.Scanner
	scope   *ast.Scope

	tok token.Token
	pos token.Position
	lit string
}

func ParseFile(src []byte, scope *ast.Scope) (f *ast.File, err error) {
	parser := newParser(src, scope)

	defer func() {
		err = parser.errors.Err()
	}()

	f = parser.parseFile()

	return
}

func ParseExpr(src string, scope *ast.Scope) (expr ast.Expression, err error) {
	parser := newParser([]byte(src), scope)
	parser.scanner.InsertLineEnd = true

	defer func() {
		err = parser.errors.Err()
	}()

	parser.openScope()
	expr = parser.parseBinary(true)
	parser.closeScope()

	parser.expect(token.EOF)

	return
}

func newParser(b []byte, scope *ast.Scope) *parser {
	parser := &parser{
		errors: ErrorList{},
		scope:  scope,
	}
	eh := func(pos token.Position, msg string) { parser.errors.Add(pos, msg) }
	parser.scanner = scanner.NewScanner(b, eh)
	parser.next()

	return parser
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
	pipeline.NamePos = p.pos
	pipeline.Name = p.parseName()
	p.expect(token.RPAREN)
	p.parsePipelineBody(pipeline)
	return pipeline
}

func (p *parser) parsePipelineBody(pipeline *ast.Pipeline) {
	p.expect(token.LBRACE)
	p.openScope()
	defer p.closeScope()
	pipeline.Scope = p.scope
	if p.tok == token.STAGES {
		pipeline.Stages, pipeline.Vars = p.parseStages()
	}
	p.expect(token.RBRACE)
}

func (p *parser) parseStages() (stages []*ast.Stage, vars *ast.VarsBlock) {
	p.expect(token.STAGES)
	p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		switch p.tok {
		case token.STAGE:
			stages = append(stages, p.parseStage())
		case token.VARS:
			if vars != nil {
				p.error("redefine vars block")
			}
			vars = p.parseVarsBlock()
		default:
			p.expectedError(token.STAGE)
			p.advance(token.STAGE)
		}
	}
	p.expect(token.RBRACE)
	return
}

func (p *parser) parseStage() *ast.Stage {
	s := &ast.Stage{}
	p.expect(token.STAGE)
	p.expect(token.LPAREN)
	s.NamePos = p.pos
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
		case token.VARS:
			if s.Vars != nil {
				p.error("redefine vars block")
			}
			s.Vars = p.parseVarsBlock()
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
	x := p.parseBinary(true)
	expr = &ast.ExprStatement{X: x}
	p.expect(token.SEMICOLON)
	p.scanner.UseLineEnd(false)
	p.expect(token.RBRACE)
	return
}

func (p *parser) parseSteps() *ast.StepBlock {
	p.expect(token.STEPS)
	pos := p.pos
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
	return &ast.StepBlock{Stmts: list, Lbrace: pos}
}

func (p *parser) parseStatement() (stmt ast.Statement) {
	lhs := p.parseBinary(true)
	switch p.tok {
	case token.DEFINE, token.ASSIGN, token.ADD_ASSIGN, token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN, token.REM_ASSIGN:
		assignTok := p.tok
		p.next()
		rhs := p.parseBinary(false)
		stmt = &ast.Assignment{Lhs: lhs, Tok: assignTok, Rhs: rhs}
		if assignTok == token.DEFINE {
			// we do not know the type of the rhs expression
			p.varDecl(lhs, "")
			if _, isPath := lhs.(*ast.PathExpr); isPath {
				p.error("define operator on path expression")
			}
		} else {
			if path, isPath := lhs.(*ast.PathExpr); isPath {
				path.Lhs = true
			}
		}
	default:
		stmt = &ast.ExprStatement{X: lhs}
	}

	p.expectStatmentEnd()
	return
}

func (p *parser) varDecl(x ast.Expression, typeName string) {
	if ident, isIdent := x.(*ast.Ident); isIdent {
		if _, identExists := p.scope.Symbol(ident.Name); identExists {
			p.error(fmt.Sprintf("identifier '%v' already defined", ident.Name))
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

func (p *parser) parseVarsBlock() *ast.VarsBlock {
	p.expect(token.VARS)
	var list []*ast.Assignment
	p.expect(token.LBRACE)
	p.scanner.UseLineEnd(true)
	for {
		if p.tok == token.RBRACE || p.tok == token.EOF {
			break
		}
		stmt := p.parseStatement()
		if a, isAssign := stmt.(*ast.Assignment); isAssign {
			list = append(list, a)
		} else {
			p.error("expected assign statement in vars block")
		}

	}
	p.scanner.UseLineEnd(false)
	p.expect(token.RBRACE)
	return &ast.VarsBlock{Specs: list}
}

func (p *parser) parseBinary(lhs bool) ast.Expression {
	x := p.parseUnary(lhs)

	// todo: consider correct operator order () before */ before +-...
	for {
		switch p.tok {
		case token.ADD, token.SUB, token.MUL, token.QUO, token.REM, token.LAND, token.LOR, token.EQL, token.LSS, token.GTR, token.NOT, token.NEQ, token.LEQ, token.GEQ:
			if lhs {
				p.resolve(x)
				lhs = false
			}
			op := p.tok
			p.next()
			y := p.parseBinary(false)
			x = &ast.Binary{Lhs: x, Rhs: y, Op: op, Precedence: op.Precedence()}
		default:
			return x
		}
	}

	return x
}

func (p *parser) parseUnary(lhs bool) ast.Expression {
	switch p.tok {
	case token.NOT:
		op := p.tok
		pos := p.pos
		p.next()
		x := p.parseUnary(false)
		return &ast.Unary{X: x, Op: op, OpPos: pos}
	}
	return p.parsePrimary(lhs)
}

func (p *parser) parsePrimary(lhs bool) ast.Expression {
	x := p.parseOperand(lhs)

	switch p.tok {
	case token.PERIOD:
		p.resolve(x)
		path := &ast.PathExpr{X: x, StartPos: x.Pos()}
		x = p.parsePath(path)
	}

	if !p.tok.IsExprEnd() && !p.tok.IsOperator() && p.tok != token.COLON && p.tok != token.COMMA {
		x = p.parseCall(x)
	}

	return x
}

func (p *parser) isLiteralType(x ast.Expression) bool {
	switch x.(type) {
	case *ast.MapType:
		return true
	}
	return false
}

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
				p.expectedError(token.COLON)
			}
		default:
			if len(values) == 1 {
				isMap = false
			} else if isMap {
				p.expectedError(token.COLON)
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
		p.resolve(val)
		return &ast.KeyValueExpr{Key: x, Value: val}
	} else {
		p.resolve(x)
	}
	return x
}

func (p *parser) parseValue() ast.Expression {
	if p.tok == token.LBRACK {
		return p.parseSequence()
	}

	return p.parseBinary(true)
}

func (p *parser) parsePath(path *ast.PathExpr) ast.Expression {
	current := path
	for {

		if p.tok == token.LBRACK {
			// if index used like findAll...}[0]
			current.Path = p.parseIndex()
		} else {
			p.next()
			if p.tok == token.PERIOD {
				p.next()
				return &ast.RangeExpr{Start: path.X, End: p.parseOperand(false)}
			}
			current.Path = p.parsePathOperand()
		}

		if p.tok == token.LBRACK {
			// if index used like list[0]
			current = &ast.PathExpr{X: current}
			current.Path = p.parseIndex()
		} else if !p.tok.IsExprEnd() && p.tok != token.PERIOD && !p.tok.IsOperator() && p.tok != token.COMMA {
			c := p.parseCall(current.Path)
			current.Path = c.Func
			current.Args = c.Args
		}

		if p.tok != token.PERIOD && p.tok != token.LBRACK {
			return current
		}
		current = &ast.PathExpr{X: current, StartPos: current.Pos()}
	}
}

func (p *parser) parsePathOperand() (x ast.Expression) {
	switch p.tok {
	case token.IDENT:
		x = &ast.Ident{Name: p.lit, NamePos: p.pos}
		p.next()
	case token.STRING, token.RSTRING:
		x = &ast.Literal{Kind: p.tok, Value: p.lit, ValuePos: p.pos}
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
		x = &ast.Literal{Kind: p.tok, Value: p.lit, ValuePos: p.pos}
		p.next()
	case token.NUMBER:
		x = &ast.Literal{Kind: p.tok, Value: p.lit, ValuePos: p.pos}
		p.next()
	}
	p.expect(token.RBRACK)
	return
}

func (p *parser) parseCall(f ast.Expression) *ast.Call {
	pos := p.pos
	list := p.parseArgList()
	return &ast.Call{Args: list, Func: f, FuncPos: pos}
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
		x = p.parseIdent(lhs)
	case token.STRING, token.RSTRING:
		x = &ast.Literal{Kind: p.tok, Value: p.lit, ValuePos: p.pos}
		p.next()
	case token.NUMBER:
		x = &ast.Literal{Kind: p.tok, Value: p.lit, ValuePos: p.pos}
		p.next()
	case token.LPAREN:
		p.next()
		x = &ast.ParenExpr{X: p.parseBinary(false)}
		p.expect(token.RPAREN)
	case token.LBRACK:
		x = p.parseSequence()
	default:

		p.error("expected operand")
		p.next()
	}
	return
}

func (p *parser) parseIdent(lhs bool) ast.Expression {
	x := &ast.Ident{Name: p.lit, NamePos: p.pos}
	if !lhs {
		p.resolve(x)
	}
	p.next()
	return x
}

func (p *parser) parseClosure() *ast.Closure {
	closure := &ast.Closure{Block: &ast.Block{}, LbracePos: p.pos}
	p.expect(token.LBRACE)
	inInput := false
	p.openScope()
	for p.tok != token.RBRACE && p.tok != token.SEMICOLON && p.tok != token.EOF {
		x := p.parseBinary(true)
		if p.tok == token.COMMA || p.tok == token.LAMBDA {
			inInput = p.tok == token.COMMA
			ident := x.(*ast.Ident)
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
		expr = p.parseBinary(true)
	}

	if p.tok == token.COLON {
		p.expect(token.COLON)
		arg.Value = p.parseBinary(false)

		arg.Name = expr.(*ast.Ident).Name
		arg.NamePos = expr.Pos()
	} else {
		p.resolve(expr)
		arg.Name = ""
		arg.NamePos = expr.Pos()
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
		if p.tok == token.RBRACE {
			// end of block semicolon not needed
			return
		}
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
