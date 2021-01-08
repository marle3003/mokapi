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
	if node != nil && !v.outer.hasErrors() {
		switch n := node.(type) {
		case *lang.Ident:
			// if argument name is not defined, name is empty
			if len(n.Name) == 0 {
				v.stack.Push(nil)
				return nil
			} else {
				return v.outer.Visit(node)
			}
		default:
			return v.outer.Visit(node)
		}
	}

	val := v.stack.Pop()
	name := v.stack.Pop()
	argName := ""
	if name != nil {
		argName = name.String()
	}
	kv := types.NewKeyValuePair(argName, val)
	v.stack.Push(kv)

	return nil
}
