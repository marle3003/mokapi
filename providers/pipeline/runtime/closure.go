package runtime

import (
	"fmt"
	"mokapi/providers/pipeline/lang"
	"mokapi/providers/pipeline/lang/types"
)

type closureVisitor struct {
	visitorImpl
	scope   *Scope
	stack   *stack
	closure *lang.Closure
	outer   visitor
}

func (v *closureVisitor) Visit(node lang.Node) lang.Visitor {
	if v.outer.hasErrors() {
		return nil
	}
	if node != nil {
		return nil
	}

	names := make([]string, len(v.closure.Params))
	for i, p := range v.closure.Params {
		names[i] = p.Name
	}

	f := func(args []types.Object) (types.Object, error) {
		parameters := map[string]types.Object{}
		for i, n := range names {
			if i > len(args)-1 {
				return nil, fmt.Errorf("index out of range of arguments")
			}
			parameters[n] = args[i]
		}
		scope := NewScope(parameters)
		scope.outer = v.scope
		return runBlock(v.closure.Block, scope)
	}

	v.stack.Push(types.NewClosure(f))

	return nil
}
