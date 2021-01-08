package lang

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
	When  Expression
}

type StepBlock struct {
	Statments []Statement
}

type Assignment struct {
	Lhs Expression
	Tok Token
	Rhs Expression
}

type ExprStatement struct {
	X Expression
}

type Unary struct {
	Op      Token
	Operand Expression
}

type Binary struct {
	Lhs Expression
	Op  Token
	Rhs Expression
}

type Call struct {
	Func Expression
	Args []*Argument
}

type IndexExpr struct {
	X     Expression
	Index Expression
}

type PathExpr struct {
	X    Expression
	Path []Expression
}

type Argument struct {
	Name  Expression
	Value Expression
}

type Ident struct {
	Name string
}

type Selector struct {
	X        Expression
	Selector Expression
}

type Literal struct {
	Kind  Token
	Value string
}

// exprNode() ensures only expression nodes can be assigned
func (*Ident) exprNode()     {}
func (*Selector) exprNode()  {}
func (*Literal) exprNode()   {}
func (*Call) exprNode()      {}
func (*Binary) exprNode()    {}
func (*Unary) exprNode()     {}
func (*IndexExpr) exprNode() {}
func (*PathExpr) exprNode()  {}

// stmtNode() ensures only statement nodes can be assigned
func (*Assignment) stmtNode()    {}
func (*ExprStatement) stmtNode() {}
