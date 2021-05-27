package ast

import "mokapi/providers/workflow/token"

type Expression interface {
	Pos() token.Position
}

type Literal struct {
	ValuePos token.Position
	Kind     token.Token
	Value    string
}

type Identifier struct {
	NamePos token.Position
	Name    string
}

type Selector struct {
	X        Expression
	Selector *Identifier
}

type CallExpr struct {
	Fun  *Identifier
	Args []Expression
}

type Closure struct {
	LambdaPos token.Position
	Func      Expression
	Args      []*Identifier
}

type Binary struct {
	Lhs   Expression
	OpPos token.Position
	Op    token.Token
	Rhs   Expression
}

type SequenceExpr struct {
	Lbrack token.Position
	Values []Expression
	IsMap  bool
}

type KeyValueExpr struct {
	Key   Expression
	Value Expression
}

type RangeExpr struct {
	Start Expression
	End   Expression
}

func (e *Literal) Pos() token.Position      { return e.ValuePos }
func (e *Identifier) Pos() token.Position   { return e.NamePos }
func (e *Selector) Pos() token.Position     { return e.X.Pos() }
func (e *CallExpr) Pos() token.Position     { return e.Fun.Pos() }
func (e *Binary) Pos() token.Position       { return e.Lhs.Pos() }
func (e *Closure) Pos() token.Position      { return e.LambdaPos }
func (e *SequenceExpr) Pos() token.Position { return e.Lbrack }
func (e *KeyValueExpr) Pos() token.Position { return e.Key.Pos() }
func (e *RangeExpr) Pos() token.Position    { return e.Start.Pos() }
