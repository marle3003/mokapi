package runtime

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang"
	"mokapi/providers/pipeline/lang/types"
)

type pathVisitor struct {
	stack *stack
	outer visitor
	expr  *lang.PathExpr
}

func (v *pathVisitor) Visit(node lang.Node) lang.Visitor {
	if node != nil && !v.outer.hasErrors() {
		return v.outer.Visit(node)
	}

	n := len(v.expr.Path)
	segments := make([]string, n)
	for i := n - 1; i >= 0; i-- {
		segments[i] = v.stack.Pop().String()
	}

	obj := v.stack.Pop()
	if obj == nil {
		v.outer.addError(errors.Errorf("null reference"))
		return nil
	}

	path := types.NewPath(obj)
	val, err := path.Resolve(segments)
	if err != nil {
		v.outer.addError(err)
	} else {
		v.stack.Push(val)
	}

	return nil
}
