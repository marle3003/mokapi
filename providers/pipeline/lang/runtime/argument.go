package runtime

import (
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/types"
)

type argumentVisitor struct {
	scope    *ast.Scope
	stack    *stack
	argument *ast.Argument
	outer    visitor
}

func (v *argumentVisitor) Visit(node ast.Node) ast.Visitor {
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
