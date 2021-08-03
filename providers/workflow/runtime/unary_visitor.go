package runtime

import (
	"mokapi/providers/workflow/ast"
	"mokapi/providers/workflow/token"
)

type unaryVisitor struct {
	outer *visitor
	u     *ast.Unary
}

func newUnaryVisitor(u *ast.Unary, outer *visitor) *unaryVisitor {
	return &unaryVisitor{u: u, outer: outer}
}

func (v *unaryVisitor) Visit(e ast.Expression) ast.Visitor {
	if e != nil {
		return v.outer.Visit(e)
	}

	x := v.outer.stack.Pop()

	switch v.u.Op {
	case token.SUB:
		switch n := x.(type) {
		case int:
			x = n * -1
		case float64:
			x = n * -1
		default:
			v.outer.errors.Addf(v.u.Pos(), "operator %v not supported for type %t", v.u.Op, x)
		}
	case token.NOT:
		switch n := x.(type) {
		case bool:
			x = !n
		}
	}

	v.outer.stack.Push(x)

	return nil
}
