package ast

import "mokapi/providers/pipeline/lang/token"

type Node interface {
}

type Expression interface {
	Node
	exprNode()
}

type Statement interface {
	stmtNode()
}

type File struct {
	Pipelines []*Pipeline
	Scope     *Scope
}

func (f *File) AddPipeline(p *Pipeline) {
	f.Pipelines = append(f.Pipelines, p)
}

type Pipeline struct {
	Name   string
	Stages []*Stage
}

type Stage struct {
	Name  string
	Steps *StepBlock
	When  *ExprStatement
	Scope *Scope
}

type StepBlock struct {
	Statments []Statement
}

type Block struct {
	Stmts []Statement
}

type Assignment struct {
	Lhs Expression
	Tok token.Token
	Rhs Expression
}

type DeclStmt struct {
	Name *Ident
	Type string
}

type ExprStatement struct {
	X Expression
}

type Unary struct {
	Op token.Token
	X  Expression
}

type Binary struct {
	Lhs        Expression
	Op         token.Token
	Rhs        Expression
	Precedence int
}

type Call struct {
	Func Expression
	Args []*Argument
}

type IndexExpr struct {
	X     Expression
	Index Expression
}

type ParenExpr struct {
	X Expression
}

type PathExpr struct {
	X    Expression
	Path Expression
	Args []*Argument
}

type Argument struct {
	Name  string
	Value Expression
}

type Ident struct {
	Name string
}

type Literal struct {
	Kind  token.Token
	Value string
}

type Closure struct {
	Params []*Ident
	Block  *Block
}

// exprNode() ensures only expression nodes can be assigned
func (*Ident) exprNode() {}

//unc (*Selector) exprNode()  {}
func (*Literal) exprNode()   {}
func (*Call) exprNode()      {}
func (*Binary) exprNode()    {}
func (*Unary) exprNode()     {}
func (*IndexExpr) exprNode() {}
func (*PathExpr) exprNode()  {}
func (*Closure) exprNode()   {}
func (*ParenExpr) exprNode() {}

// stmtNode() ensures only statement nodes can be assigned
func (*Assignment) stmtNode()    {}
func (*ExprStatement) stmtNode() {}
func (*DeclStmt) stmtNode()      {}
