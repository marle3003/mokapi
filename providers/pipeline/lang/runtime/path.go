package runtime

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/types"
)

type pathVisitor struct {
	scope *ast.Scope
	stack *stack
	outer visitor
	expr  *ast.PathExpr
}

func (v *pathVisitor) Visit(node ast.Node) ast.Visitor {
	if v.outer.hasErrors() {
		return nil
	}
	if node != nil {
		switch n := node.(type) {
		case *ast.Ident:
			v.stack.Push(types.NewString(n.Name))
			return nil
		case *ast.IndexExpr:
			return v
		}
		return v.outer.Visit(node)
	}

	n := len(v.expr.Args)
	args := make(map[string]types.Object, n)
	for i := n - 1; i >= 0; i-- {
		kv := v.stack.Pop().(*types.KeyValuePair)
		argName := kv.Key
		if len(argName) == 0 {
			argName = fmt.Sprintf("%v", i)
		}
		args[argName] = kv.Value
	}

	path := v.stack.Pop()

	obj := v.stack.Pop()
	if obj == nil {
		v.outer.addError(errors.Errorf("null reference"))
		return nil
	}
	if ident, isIdent := v.expr.X.(*ast.Ident); isIdent {
		if o, exists := v.scope.Symbol(ident.Name); exists {
			obj = o
		} else {
			v.outer.addError(errors.Errorf("unable to resolve identifier %v", obj.String()))
		}
	}

	p, ok := obj.(*types.Path)
	if !ok {
		p = types.NewPath(obj)
	}

	val, err := p.Resolve(path.String(), args)
	if err != nil {
		// we can not resolve, push it back to stack
		v.stack.Push(obj)
		v.stack.Push(path)
	} else {
		v.stack.Push(val)
	}

	return nil
}
