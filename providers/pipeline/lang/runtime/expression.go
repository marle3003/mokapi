package runtime

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/token"
	"mokapi/providers/pipeline/lang/types"
	"strconv"
)

type exprVisitor struct {
	visitorImpl
	scope *ast.Scope
	stack *stack
}

func (v *exprVisitor) Visit(node ast.Node) ast.Visitor {
	if v.hasErrors() {
		return nil
	}

	switch n := node.(type) {
	case *ast.Call:
		return &callVisitor{scope: v.scope, stack: v.stack, call: n, outer: v}
	case *ast.Argument:
		return &argumentVisitor{scope: v.scope, stack: v.stack, argument: n, outer: v}
	case *ast.ExprStatement:
		v.stack.Reset()
	case *ast.Assignment:
		v.stack.Reset()
		return &assignVisitor{scope: v.scope, stack: v.stack, assign: n, outer: v}
	case *ast.IndexExpr:
		return &indexVisitor{stack: v.stack, outer: v}
	case *ast.PathExpr:
		return &pathVisitor{scope: v.scope, stack: v.stack, expr: n, outer: v}
	case *ast.Binary:
		return newBinaryVisitor(n, v.stack, v)
	case *ast.Closure:
		return &closureVisitor{stack: v.stack, scope: v.scope, outer: v, closure: n}
	case *ast.Ident:
		if o, ok := v.scope.Symbol(n.Name); ok {
			v.stack.Push(o)
		} else {
			v.addError(errors.Errorf("Unresolved symbol '%v'", n.Name))
		}
	case *ast.Literal:
		switch n.Kind {
		case token.STRING:
			s, err := format(n.Value, v.scope)
			if err != nil {
				v.addError(err)
				return nil
			}
			v.stack.Push(types.NewString(s))
		case token.RSTRING:
			v.stack.Push(types.NewString(n.Value))
		case token.NUMBER:
			f, err := strconv.ParseFloat(n.Value, 64)
			if err != nil {
				v.addError(err)
				return nil
			}
			v.stack.Push(types.NewNumber(f))
		}
	}

	return v
}
