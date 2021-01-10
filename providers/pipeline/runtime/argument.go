package runtime

import (
	"mokapi/providers/pipeline/lang"
	"mokapi/providers/pipeline/lang/types"
)

type argumentVisitor struct {
	scope    *Scope
	stack    *stack
	argument *lang.Argument
	outer    visitor
}

func (v *argumentVisitor) Visit(node lang.Node) lang.Visitor {
	if v.outer.hasErrors() {
		return nil
	}
	if node != nil {
		return v.outer.Visit(node)
	}

	val := v.stack.Pop()
	kv := types.NewKeyValuePair(v.argument.Name, val)
	v.stack.Push(kv)

	return nil
}
