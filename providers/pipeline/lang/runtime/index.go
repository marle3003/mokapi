package runtime

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/ast"
)

type indexVisitor struct {
	stack *stack
	outer visitor
}

func (v *indexVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil && !v.outer.hasErrors() {
		return v.outer.Visit(node)
	}

	index := v.stack.Pop()
	if index == nil {
		v.outer.addError(errors.Errorf("index is null"))
		return nil
	}

	obj := v.stack.Pop()
	if obj == nil {
		v.outer.addError(errors.Errorf("null reference"))
		return nil
	}

	val, err := obj.GetField(index.String())
	if err != nil {
		v.outer.addError(err)
	} else {
		v.stack.Push(val)
	}

	return nil
}
