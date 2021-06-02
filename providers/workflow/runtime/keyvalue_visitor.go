package runtime

import (
	"mokapi/providers/workflow/ast"
)

type keyValueVisitor struct {
	outer *visitor
}

func newKeyValueVisitor(outer *visitor) *keyValueVisitor {
	return &keyValueVisitor{outer: outer}
}

func (v *keyValueVisitor) Visit(e ast.Expression) ast.Visitor {
	if e != nil {
		switch n := e.(type) {
		case *ast.Identifier:
			v.outer.stack.Push(n.Name)
			return nil
		default:
			return v.outer.Visit(e)
		}
	}

	return nil
}
