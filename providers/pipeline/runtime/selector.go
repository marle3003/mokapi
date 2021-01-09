package runtime

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang"
)

type selectorVisitor struct {
	stack  *stack
	outer  visitor
	inCall bool
}

func (v *selectorVisitor) Visit(node lang.Node) lang.Visitor {
	if node != nil && !v.outer.hasErrors() {
		return v.outer.Visit(node)
	}

	selector := v.stack.Pop()
	if selector == nil {
		v.outer.addError(errors.Errorf("null reference"))
		return nil
	}

	obj := v.stack.Pop()
	if obj == nil {
		v.outer.addError(errors.Errorf("null reference"))
		return nil
	}

	name := selector.String()
	if obj.HasField(name) {
		val, err := obj.GetField(name)
		if err != nil {
			v.outer.addError(err)
		} else {
			v.stack.Push(val)
		}
	} else {
		if v.inCall {
			v.stack.Push(obj)
			v.stack.Push(selector)
		} else {
			val, err := obj.InvokeFunc(name, nil)
			if err != nil {
				v.outer.addError(err)
			} else {
				v.stack.Push(val)
			}
		}
	}

	return nil
}
