package runtime

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang"
)

type selectorVisitor struct {
	scope  *Scope
	stack  *stack
	outer  visitor
	inCall bool
}

func newSelectorVisitor(scope *Scope, stack *stack, inCall bool, root bool, outer visitor) *selectorVisitor {
	return &selectorVisitor{
		scope:  scope,
		stack:  stack,
		outer:  outer,
		inCall: inCall,
	}
}

func (v *selectorVisitor) Visit(node lang.Node) lang.Visitor {
	if v.outer.hasErrors() {
		return nil
	}

	if node != nil {
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
	//if v.root {
	//	var ok bool
	//	obj, ok = v.scope.Symbol(obj.String())
	//	if !ok {
	//		v.outer.addError(errors.Errorf("%v is not defined", obj))
	//		return nil
	//	}
	//}

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
				v.outer.addError(errors.Errorf("field or function %v not found", name))
			} else {
				v.stack.Push(val)
			}
		}
	}

	return nil
}
