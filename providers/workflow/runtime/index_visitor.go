package runtime

import (
	"mokapi/providers/workflow/ast"
	"mokapi/providers/workflow/path/objectpath"
)

type indexVisitor struct {
	outer *visitor
	exp   ast.Expression
}

func newIndexVisitor(outer *visitor, exp ast.Expression) *indexVisitor {
	return &indexVisitor{outer: outer, exp: exp}
}

func (v *indexVisitor) Visit(e ast.Expression) ast.Visitor {
	if e != nil {
		return v.outer.Visit(e)
	}

	index := v.outer.stack.Pop()
	source := v.outer.stack.Pop()

	m, err := objectpath.Resolve(index, source)
	if err != nil {
		v.outer.errors.Addf(v.exp.Pos(), "unable to resolve index: %v", err)
	}
	v.outer.stack.Push(m)

	return nil
}
