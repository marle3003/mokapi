package runtime

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang"
	"mokapi/providers/pipeline/lang/types"
	"strconv"
)

type exprVisitor struct {
	visitorImpl
	scope *Scope
	stack *stack
}

func (v *exprVisitor) Visit(node lang.Node) lang.Visitor {
	if v.hasErrors() {
		return nil
	}

	switch n := node.(type) {
	case *lang.Call:
		return &callVisitor{scope: v.scope, stack: v.stack, call: n, outer: v}
	case *lang.Argument:
		return &argumentVisitor{scope: v.scope, stack: v.stack, argument: n, outer: v}
	case *lang.ExprStatement:
		v.stack.Reset()
	case *lang.Assignment:
		v.stack.Reset()
		return &assignVisitor{scope: v.scope, stack: v.stack, assign: n, outer: v}
	case *lang.Selector:
		return &selectorVisitor{scope: v.scope, stack: v.stack, outer: v}
	case *lang.IndexExpr:
		return &indexVisitor{stack: v.stack, outer: v}
	case *lang.PathExpr:
		return &pathVisitor{scope: v.scope, stack: v.stack, expr: n, outer: v}
	case *lang.Binary:
		return newBinaryVisitor(n, v.stack, v)
	case *lang.Closure:
		return &closureVisitor{stack: v.stack, scope: v.scope, outer: v, closure: n}
	case *lang.Ident:
		v.stack.Push(types.NewString(n.Name))
	case *lang.SymbolRef:
		if o, ok := v.scope.Symbol(n.Name); ok {
			v.stack.Push(o)
		} else {
			v.addError(errors.Errorf("%v is not defined", n.Name))
			return nil
		}
	case *lang.Literal:
		switch n.Kind {
		case lang.STRING:
			s, err := format(n.Value, v.scope)
			if err != nil {
				v.addError(err)
				return nil
			}
			v.stack.Push(types.NewString(s))
		case lang.RSTRING:
			v.stack.Push(types.NewString(n.Value))
		case lang.NUMBER:
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
