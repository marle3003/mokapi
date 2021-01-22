package ast

import "mokapi/providers/pipeline/lang/token"

type Node interface {
	Pos() token.Position
}

type Expression interface {
	Node
	exprNode()
}

type Statement interface {
	Node
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
	NamePos token.Position
	Name    string
	Stages  []*Stage
	Vars    *VarsBlock
	Scope   *Scope
}

type Stage struct {
	NamePos token.Position
	Name    string
	Steps   *StepBlock
	When    *ExprStatement
	Scope   *Scope
	Vars    *VarsBlock
}

type StepBlock struct {
	Lbrace token.Position
	Stmts  []Statement
}

type Block struct {
	Lbrace token.Position
	Stmts  []Statement
}

type Assignment struct {
	Lhs    Expression
	TokPos token.Position
	Tok    token.Token
	Rhs    Expression
}

type VarsBlock struct {
	Lbrace token.Position
	Specs  []*Assignment
}

type ExprStatement struct {
	X Expression
}

type SequenceExpr struct {
	Lbrack token.Position
	Values []Expression
	IsMap  bool
}

type Unary struct {
	OpPos token.Position
	Op    token.Token
	X     Expression
}

type Binary struct {
	Lhs        Expression
	OpPos      token.Position
	Op         token.Token
	Rhs        Expression
	Precedence int
}

type Call struct {
	FuncPos token.Position
	Func    Expression
	Args    []*Argument
}

type IndexExpr struct {
	X        Expression
	IndexPos token.Position
	Index    Expression
}

type ParenExpr struct {
	X Expression
}

type PathExpr struct {
	StartPos token.Position
	X        Expression
	Path     Expression
	Args     []*Argument
	Lhs      bool
}

type Argument struct {
	NamePos token.Position
	Name    string
	Value   Expression
}

type Ident struct {
	NamePos token.Position
	Name    string
}

type Literal struct {
	ValuePos token.Position
	Kind     token.Token
	Value    string
}

type Closure struct {
	LbracePos token.Position
	Params    []*Ident
	Block     *Block
}

type KeyValueExpr struct {
	Key   Expression
	Value Expression
}

type MapType struct {
	Map token.Position
}

// exprNode() ensures only expression nodes can be assigned
func (*Ident) exprNode()        {}
func (*Literal) exprNode()      {}
func (*Call) exprNode()         {}
func (*Binary) exprNode()       {}
func (*Unary) exprNode()        {}
func (*IndexExpr) exprNode()    {}
func (*PathExpr) exprNode()     {}
func (*Closure) exprNode()      {}
func (*ParenExpr) exprNode()    {}
func (*MapType) exprNode()      {}
func (*SequenceExpr) exprNode() {}
func (*KeyValueExpr) exprNode() {}

// stmtNode() ensures only statement nodes can be assigned
func (*Assignment) stmtNode()    {}
func (*ExprStatement) stmtNode() {}

// Pos implementations
func (f *File) Pos() token.Position {
	if len(f.Pipelines) > 0 {
		return f.Pipelines[0].NamePos
	}
	return token.Position{Line: 0, Column: 0}
}
func (p *Pipeline) Pos() token.Position      { return p.NamePos }
func (s *Stage) Pos() token.Position         { return s.NamePos }
func (s *StepBlock) Pos() token.Position     { return s.Lbrace }
func (s *Block) Pos() token.Position         { return s.Lbrace }
func (a *Assignment) Pos() token.Position    { return a.Lhs.Pos() }
func (d *VarsBlock) Pos() token.Position     { return d.Lbrace }
func (e *ExprStatement) Pos() token.Position { return e.X.Pos() }
func (u *Unary) Pos() token.Position         { return u.OpPos }
func (b *Binary) Pos() token.Position        { return b.Lhs.Pos() }
func (c *Call) Pos() token.Position          { return c.FuncPos }
func (i *IndexExpr) Pos() token.Position     { return i.X.Pos() }
func (p *ParenExpr) Pos() token.Position     { return p.X.Pos() }
func (p *PathExpr) Pos() token.Position      { return p.StartPos }
func (a *Argument) Pos() token.Position      { return a.NamePos }
func (l *Literal) Pos() token.Position       { return l.ValuePos }
func (c *Closure) Pos() token.Position       { return c.LbracePos }
func (i *Ident) Pos() token.Position         { return i.NamePos }
func (m *MapType) Pos() token.Position       { return m.Map }
func (c *SequenceExpr) Pos() token.Position  { return c.Lbrack }
func (k *KeyValueExpr) Pos() token.Position  { return k.Key.Pos() }
