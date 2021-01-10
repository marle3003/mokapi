package runtime

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang"
	"mokapi/providers/pipeline/lang/types"
)

type callVisitor struct {
	visitorImpl
	scope *Scope
	stack *stack
	call  *lang.Call
	outer visitor
}

func (v *callVisitor) Visit(node lang.Node) lang.Visitor {
	if v.outer.hasErrors() {
		return nil
	}
	if node != nil {
		switch n := node.(type) {
		case *lang.Selector:
			return newSelectorVisitor(v.scope, v.stack, true, true, v.outer)
		case *lang.Ident:
			if o, ok := v.scope.Symbol(n.Name); ok {
				v.stack.Push(nil)
				v.stack.Push(o)
			} else {
				v.stack.Push(types.NewString(n.Name))
			}
			return nil
		default:
			return v.outer.Visit(n)
		}
	}

	n := len(v.call.Args)
	args := make(map[string]types.Object, n)
	for i := n - 1; i >= 0; i-- {
		kv := v.stack.Pop().(*types.KeyValuePair)
		argName := kv.Key
		if len(argName) == 0 {
			argName = fmt.Sprintf("%v", i)
		}
		args[argName] = kv.Value
	}

	f := v.stack.Pop()
	target := v.stack.Pop()
	if target == nil {
		if step, ok := f.(types.Step); ok {
			v.callStep(step, args)
		} else {
			v.outer.addError(errors.Errorf("unresovled func call '%v'", f))
			return nil
		}
	} else {
		val, err := target.InvokeFunc(f.String(), args)
		if err != nil {
			v.outer.addError(err)
			return nil
		}
		v.stack.Push(val)
	}

	return nil
}
