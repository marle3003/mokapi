package runtime

import (
	"fmt"
	"mokapi/providers/pipeline/lang"
	"mokapi/providers/pipeline/lang/types"
)

type callVisitor struct {
	scope *Scope
	stack *stack
	call  *lang.Call
	outer visitor
}

func (v *callVisitor) Visit(node lang.Node) lang.Visitor {
	if node != nil {
		return v.outer.Visit(node)
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
	v.callFunc(f, args)

	return nil
}

func (v *callVisitor) callFunc(obj types.Object, args map[string]types.Object) {
	if step, ok := obj.(types.Step); ok {
		v.callStep(step, args)
	}
}
