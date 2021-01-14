package runtime

import (
	"mokapi/providers/pipeline/lang/ast"
)

type indexVisitor struct {
	outer visitor
	index *ast.IndexExpr
}

func newIndexVisitor(index *ast.IndexExpr, outer visitor) *indexVisitor {
	return &indexVisitor{index: index, outer: outer}
}

func (v *indexVisitor) Visit(node ast.Node) ast.Visitor {
	if v.outer.HasErrors() {
		return nil
	}
	if node != nil {
		return v.outer.Visit(node)
	}

	index := v.outer.Stack().Pop()
	if index == nil {
		v.outer.AddError(v.index.Index.Pos(), "index is nil")
		return nil
	}

	obj := v.outer.Stack().Pop()
	if obj == nil {
		v.outer.AddError(v.index.Pos(), "nil reference")
		return nil
	}

	val, err := obj.GetField(index.String())
	if err != nil {
		v.outer.AddError(v.index.Pos(), err.Error())
	} else {
		v.outer.Stack().Push(val)
	}

	return nil
}
