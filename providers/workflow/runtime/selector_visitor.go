package runtime

import (
	"mokapi/providers/workflow/ast"
	"mokapi/providers/workflow/path/objectpath"
)

type selectorVisitor struct {
	outer *visitor
	exp   ast.Expression
}

func newSelectorVisitor(outer *visitor) *selectorVisitor {
	return &selectorVisitor{outer: outer}
}

func (v *selectorVisitor) Visit(e ast.Expression) ast.Visitor {
	if e != nil {
		v.exp = e
		switch t := e.(type) {
		case *ast.Identifier:
			v.outer.stack.Push(t.Name)
		case *ast.Selector:
			if ident, ok := t.X.(*ast.Identifier); ok {
				v.outer.Visit(ident)
				v.Visit(t.Selector)
				return v.Visit(nil)
			}
			return v
		case *ast.IndexExpr:
			return v.outer.Visit(t)
		}
		return nil
	}

	selector := v.outer.stack.Pop()
	source := v.outer.stack.Pop()

	m, err := objectpath.Resolve(selector, source)
	if err != nil {
		v.outer.errors.Addf(v.exp.Pos(), "unable to resolve object path: %v", err)
	}
	v.outer.stack.Push(m)

	return nil
}
