package runtime

import (
	"fmt"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/types"
)

type closureVisitor struct {
	visitorImpl
	scope   *ast.Scope
	stack   *stack
	closure *ast.Closure
	outer   visitor
}

func (v *closureVisitor) Visit(node ast.Node) ast.Visitor {
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
		scope := ast.NewScopeWithOuter(parameters, v.scope)
		return runBlock(v.closure.Block, scope)
	}

	v.stack.Push(types.NewClosure(f))

	return nil
}
