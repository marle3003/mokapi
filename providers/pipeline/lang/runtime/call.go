package runtime

import (
	"fmt"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/types"
)

type callVisitor struct {
	call  *ast.Call
	outer visitor
}

func newCallVisitor(call *ast.Call, outer visitor) *callVisitor {
	return &callVisitor{call: call, outer: outer}
}

func (v *callVisitor) Visit(node ast.Node) ast.Visitor {
	if v.outer.HasErrors() {
		return nil
	}
	if node != nil {
		switch n := node.(type) {
		//case *lang.Selector:
		//	return newSelectorVisitor(v.scope, v.stack, true, true, v.outer)
		case *ast.Ident:
			if o, ok := v.outer.Scope().Symbol(n.Name); ok {
				v.outer.Stack().Push(nil)
				v.outer.Stack().Push(o)
			} else {
				v.outer.Stack().Push(types.NewString(n.Name))
			}
			return nil
		default:
			return v.outer.Visit(n)
		}
	}

	n := len(v.call.Args)
	args := make(map[string]types.Object, n)
	for i := n - 1; i >= 0; i-- {
		kv := v.outer.Stack().Pop().(*types.KeyValuePair)
		argName := kv.Key
		if len(argName) == 0 {
			argName = fmt.Sprintf("%v", i)
		}
		args[argName] = kv.Value
	}

	f := v.outer.Stack().Pop()
	target := v.outer.Stack().Pop()
	if target == nil {
		if step, ok := f.(types.Step); ok {
			v.callStep(step, args)
		} else {
			v.outer.AddErrorf(v.call.Pos(), "unresolved func call '%v'", f)
			return nil
		}
	} else {
		val, err := target.InvokeFunc(f.String(), args)
		if err != nil {
			v.outer.AddError(v.call.Pos(), err.Error())
			return nil
		}
		v.outer.Stack().Push(val)
	}

	return nil
}
