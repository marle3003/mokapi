package runtime

import (
	"fmt"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/types"
)

type closureVisitor struct {
	closure *ast.Closure
	outer   visitor
}

func newClosureVisitor(closure *ast.Closure, outer visitor) *closureVisitor {
	return &closureVisitor{closure: closure, outer: outer}
}

func (v *closureVisitor) Visit(node ast.Node) ast.Visitor {
	if v.outer.HasErrors() {
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
		scope := ast.NewScopeWithOuter(parameters, v.outer.Scope())
		return runBlock(v.closure.Block, scope)
	}

	v.outer.Stack().Push(types.NewClosure(f))

	return nil
}
