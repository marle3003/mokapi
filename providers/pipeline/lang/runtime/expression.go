package runtime

import (
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/token"
	"mokapi/providers/pipeline/lang/types"
	"strconv"
)

type exprVisitor struct {
	visitorErrorHandler
	stack *stack
	scope *ast.Scope
}

func newExprVisitor(stack *stack, scope *ast.Scope) *exprVisitor {
	return &exprVisitor{stack: stack, scope: scope}
}

func (v *exprVisitor) Visit(node ast.Node) ast.Visitor {
	if v.HasErrors() {
		return nil
	}

	switch n := node.(type) {
	case *ast.Call:
		return newCallVisitor(n, v)
	case *ast.Argument:
		return newArgumentVisitor(n, v)
	case *ast.ExprStatement:
		v.Stack().Reset()
	case *ast.Assignment:
		v.Stack().Reset()
		return newAssignVisitor(n, v)
	case *ast.IndexExpr:
		return newIndexVisitor(n, v)
	case *ast.PathExpr:
		return newPathVisitor(n, v)
	case *ast.Binary:
		return newBinaryVisitor(n, v)
	case *ast.Closure:
		return newClosureVisitor(n, v)
	case *ast.SequenceExpr:
		return newSequenceVisitor(n, v)
	case *ast.Ident:
		if o, ok := v.Scope().Symbol(n.Name); ok {
			v.stack.Push(o)
		} else {
			// push name onto stack
			v.stack.Push(types.NewString(n.Name))
		}
	case *ast.Literal:
		switch n.Kind {
		case token.STRING:
			s, err := format(n.Value, v.Scope())
			if err != nil {
				v.AddError(n.Pos(), err.Error())
				return nil
			}
			v.Stack().Push(types.NewString(s))
		case token.RSTRING:
			v.Stack().Push(types.NewString(n.Value))
		case token.NUMBER:
			f, err := strconv.ParseFloat(n.Value, 64)
			if err != nil {
				v.AddError(n.Pos(), err.Error())
				return nil
			}
			v.Stack().Push(types.NewNumber(f))
		}
	}

	return v
}

func (v *exprVisitor) Stack() *stack {
	return v.stack
}

func (v *exprVisitor) Scope() *ast.Scope {
	return v.scope
}

func (v *exprVisitor) CloseScope() {
	v.scope = v.scope.Outer
}
