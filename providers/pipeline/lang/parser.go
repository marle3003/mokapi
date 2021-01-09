package lang

import (
	"fmt"
)

/*
	EBNF

	pipeline = 'pipeline(' [ identifier ] ')' '{' { NEWLINE }-


	identifier = '[^'|\\']*'|"[^"|\\"]*"
*/

type parser struct {
	errors  ErrorList
	scanner *Scanner

	tok Token
	pos Position
	lit string

	inParamList bool
}

func ParseFile(src []byte) (f *File, err error) {
	parser := &parser{
		errors:  ErrorList{},
		scanner: NewScanner(src, nil)}
	parser.next()

	defer func() {
		err = parser.errors.Err()
	}()

	f = parser.parseFile()

	return
}

func ParseExpr(b []byte) (expr Expression, err error) {
	parser := &parser{
		errors:  ErrorList{},
		scanner: NewScanner(b, nil)}
	parser.scanner.insertLineEnd = true
	parser.next()

	defer func() {
		err = parser.errors.Err()
	}()

	expr = parser.parseBinary()

	return
}

func (p *parser) parseFile() (f *File) {
	f = &File{}
	for p.tok != EOF {
		switch p.tok {
		case PIPELINE:
			pipeline := p.parsePipeline()
			f.AddPipeline(pipeline)
		default:
			p.expectedError(PIPELINE)
			p.advance(PIPELINE)
		}

	}
	return
}

func (p *parser) parsePipeline() *Pipeline {
	pipeline := &Pipeline{}
	p.expect(PIPELINE)
	p.expect(LPAREN)
	pipeline.Name = p.parseName()
	p.expect(RPAREN)
	p.parsePipelineBody(pipeline)
	return pipeline
}

func (p *parser) parsePipelineBody(pipeline *Pipeline) {
	p.expect(LBRACE)
	if p.tok == STAGES {
		pipeline.Stages = p.parseStages()
	}
	p.expect(RBRACE)
}

func (p *parser) parseStages() []*Stage {
	stages := make([]*Stage, 0)
	p.expect(STAGES)
	p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
		switch p.tok {
		case STAGE:
			stages = append(stages, p.parseStage())
		default:
			p.expectedError(STAGE)
			p.advance(STAGE)
		}
	}
	p.expect(RBRACE)
	return stages
}

func (p *parser) parseStage() *Stage {
	s := &Stage{}
	p.expect(STAGE)
	p.expect(LPAREN)
	s.Name = p.parseName()
	p.expect(RPAREN)
	p.expect(LBRACE)

	for p.tok != RBRACE && p.tok != EOF {
		switch p.tok {
		case STEPS:
			if s.Steps != nil {
				p.error("redefine steps block")
			}
			s.Steps = p.parseSteps()
		case WHEN:
			if s.When != nil {
				p.error("redefine steps block")
			}
			s.When = p.parseWhen()
		default:
			p.error("expected steps or when")
			p.advance(STEPS, WHEN)
		}
	}

	p.expect(RBRACE)
	return s
}

func (p *parser) parseWhen() (expr *ExprStatement) {
	p.expect(WHEN)
	p.expect(LBRACE)
	p.scanner.UseLineEnd(true)
	x := p.parseBinary()
	expr = &ExprStatement{X: x}
	p.expect(SEMICOLON)
	p.scanner.UseLineEnd(false)
	p.expect(RBRACE)
	return
}

func (p *parser) parseSteps() *StepBlock {
	p.expect(STEPS)
	p.expect(LBRACE)
	p.scanner.UseLineEnd(true)
	var list []Statement
	for {
		list = append(list, p.parseStatement())
		if p.tok == RBRACE || p.tok == EOF {
			break
		}
	}
	p.scanner.UseLineEnd(false)
	p.expect(RBRACE)
	return &StepBlock{Statments: list}
}

func (p *parser) parseStatement() (stmt Statement) {
	lhs := p.parseBinary()
	switch p.tok {
	case DEFINE, ASSIGN:
		assignTok := p.tok
		p.next()
		rhs := p.parseBinary()
		stmt = &Assignment{Lhs: lhs, Tok: assignTok, Rhs: rhs}
	default:
		stmt = &ExprStatement{X: lhs}
	}
	p.expectStatmentEnd()
	return
}

func (p *parser) parseBinary() Expression {
	x := p.parseUnary()

	// todo: consider correct operator order () before */ before +-...
	for {
		switch p.tok {
		case ADD, SUB, MUL, QUO, REM, LAND, LOR, EQL, LSS, GTR, NOT, NEQ, LEQ, GEQ:
			op := p.tok
			p.next()
			y := p.parseBinary()
			x = &Binary{Lhs: x, Rhs: y, Op: op, Precedence: op.Precedence()}
		default:
			return x
		}
	}

	return x
}

func (p *parser) parseUnary() Expression {
	return p.parsePrimary()
}

func (p *parser) parsePrimary() Expression {
	operand := p.parseOperand()

L:
	for {
		switch p.tok {
		case PERIOD:
			p.next()
			switch p.tok {
			case IDENT:
				s := p.parseOperand()
				operand = &Selector{X: operand, Selector: s}
			case STRING, RSTRING:
				path := &PathExpr{X: operand}
				path = p.parsePath(path)
				operand = path
			}
		case LBRACK:
			p.next()
			index := p.parseOperand()
			operand = &IndexExpr{X: operand, Index: index}
			p.expect(RBRACK)
		default:
			break L
		}
	}

	if p.tok != SEMICOLON && p.tok != EOF && !p.tok.isOperator() && !p.inParamList && p.tok != RPAREN {
		return p.parseCall(operand)
	}
	return operand
}

func (p *parser) parsePath(path *PathExpr) *PathExpr {
	current := path
	for {
		current.Path = p.parseOperand()
		if p.tok == PERIOD {
			current = &PathExpr{X: current}
			p.next()
		} else if p.tok != SEMICOLON && p.tok != EOF && !p.tok.isOperator() && !p.inParamList && p.tok != RPAREN {
			current.Args = p.parseArgList()
			if p.tok != RPAREN {
				return current
			}
			p.expect(RPAREN)
		} else {
			return current
		}
	}
}

func (p *parser) parseCall(call Expression) *Call {
	list := p.parseArgList()
	return &Call{Args: list, Func: call}
}

func (p *parser) parseArgList() []*Argument {
	var list []*Argument
	p.inParamList = true
	for p.tok != SEMICOLON && p.tok != EOF {
		list = append(list, p.parseArgument())
		if p.tok != COMMA {
			if p.tok != SEMICOLON {
				p.expect(COMMA)
			}
			break
		}
		p.next()
	}
	p.inParamList = false
	return list
}

func (p *parser) parseOperand() (x Expression) {
	switch p.tok {
	case IDENT:
		x = &Ident{Name: p.lit}
		p.next()
	case STRING, RSTRING:
		x = &Literal{Kind: p.tok, Value: p.lit}
		p.next()
	case NUMBER:
		x = &Literal{Kind: p.tok, Value: p.lit}
		p.next()
	case LBRACE:
		x = p.parseClosure()
	default:
		p.error("expected operand")
		p.next()
	}
	return
}

func (p *parser) parseClosure() *Closure {
	p.expect(LBRACE)
	closure := &Closure{Block: &Block{}}
	inInput := false
	for p.tok != RBRACE && p.tok != SEMICOLON && p.tok != EOF {
		x := p.parseBinary()
		if p.tok == COMMA {
			inInput = true
			closure.Params = append(closure.Params, x.(*Ident))
			p.next()
		} else if p.tok == LAMBDA {
			closure.Params = append(closure.Params, x.(*Ident))
			inInput = false
			p.next()
		} else {
			closure.Block.Stmts = append(closure.Block.Stmts, &ExprStatement{X: x})
		}
	}
	if inInput {
		p.expectedError(LAMBDA)
	}
	p.expect(RBRACE)
	return closure
}

func (p *parser) parseArgument() *Argument {
	arg := &Argument{}

	expr := p.parseBinary()
	if p.tok != COMMA && p.tok != SEMICOLON {
		arg.Value = p.parseBinary()
		arg.Name = expr
	} else {
		arg.Name = &Ident{}
		arg.Value = expr
	}

	return arg
}

func (p *parser) parseName() (name string) {
	if p.tok == STRING || p.tok == RSTRING {
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

func (p *parser) expect(tok Token) {
	if p.tok != tok {
		p.expectedError(tok)
	}
	p.next()
}

func (p *parser) expectStatmentEnd() {
	if p.tok != SEMICOLON && p.tok != EOF {
		p.expectedError(SEMICOLON)
	}
	p.next()
}

func (p *parser) expectedError(tok Token) {
	msg := fmt.Sprintf("expected '%v' but found '%v'", tok.String(), p.tok.String())
	if !tok.isKeyword() {
		msg += fmt.Sprintf(": %v", p.lit)
	}
	p.errors.Add(p.pos, msg)
}

func (p *parser) error(msg string) {
	p.errors.Add(p.pos, msg)
}

func (p *parser) advance(tok ...Token) {
	tokens := make(map[Token]bool)
	for _, t := range tok {
		tokens[t] = true
	}
	p.advanceOneOf(tokens)
}

func (p *parser) advanceOneOf(to map[Token]bool) {
	for p.tok != EOF {
		if _, ok := to[p.tok]; ok {
			return
		}
		p.next()
	}
}
