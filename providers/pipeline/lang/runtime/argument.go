package runtime

import (
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/types"
)

type argumentVisitor struct {
	argument *ast.Argument
	outer    visitor
}

func newArgumentVisitor(arg *ast.Argument, outer visitor) *argumentVisitor {
	return &argumentVisitor{argument: arg, outer: outer}
}

func (v *argumentVisitor) Visit(node ast.Node) ast.Visitor {
	if v.outer.HasErrors() {
		return nil
	}
	if node != nil {
		return v.outer.Visit(node)
	}

	val := v.outer.Stack().Pop()
	kv := types.NewKeyValuePair(v.argument.Name, val)
	v.outer.Stack().Push(kv)

	return nil
}
