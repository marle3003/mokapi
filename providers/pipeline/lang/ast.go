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
	When  *ExprStatement
}

type StepBlock struct {
	Statments []Statement
}

type Block struct {
	Stmts []Statement
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
	Op Token
	X  Expression
}

type Binary struct {
	Lhs        Expression
	Op         Token
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

type SymbolRef struct {
	Name string
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

type Closure struct {
	Params []*Ident
	Block  *Block
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
func (*Closure) exprNode()   {}
func (*ParenExpr) exprNode() {}
func (*SymbolRef) exprNode() {}

// stmtNode() ensures only statement nodes can be assigned
func (*Assignment) stmtNode()    {}
func (*ExprStatement) stmtNode() {}
