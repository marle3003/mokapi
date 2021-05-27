package runtime

import (
	"fmt"
	"mokapi/providers/workflow/ast"
	"mokapi/providers/workflow/functions"
	"mokapi/providers/workflow/token"
	"strconv"
)

type visitor struct {
	ctx   *WorkflowContext
	stack *stack
	vars  map[string]interface{}
}

type stack struct {
	values []interface{}
}

func newVisitor(ctx *WorkflowContext) *visitor {
	stack := newStack()
	return &visitor{ctx: ctx, stack: stack, vars: make(map[string]interface{})}
}

func newStack() *stack {
	return &stack{}
}

func (v *visitor) Visit(e ast.Expression) ast.Visitor {
	switch t := e.(type) {
	case *ast.Literal:
		switch t.Kind {
		case token.INT:
			// TODO error
			i, _ := strconv.Atoi(t.Value)
			v.stack.Push(i)
		case token.FLOAT:
			// TODO error
			i, _ := strconv.ParseFloat(t.Value, 64)
			v.stack.Push(i)
		default:
			v.stack.Push(t.Value)
		}
	case *ast.Binary:
		return newBinaryVisitor(t, v)
	case *ast.Selector:
		s := newSelectorVisitor(v)
		return s.Visit(e)
	case *ast.Identifier:
		if x, ok := v.vars[t.Name]; ok {
			v.stack.Push(x)
		} else {
			i := v.ctx.Context.Get(t.Name)
			v.stack.Push(i)
		}
		return nil
	case *ast.CallExpr:
		c := newCallVisitor(v)
		return c.Visit(e)
	case *ast.Closure:
		var f functions.Function
		f = func(args ...interface{}) (interface{}, error) {
			fVisitor := newVisitor(v.ctx)
			for i, p := range t.Args {
				if i >= len(args) {
					return nil, fmt.Errorf("invalid parameter length")
				}
				fVisitor.vars[p.Name] = args[i]
			}
			ast.Walk(fVisitor, t.Func)
			return fVisitor.stack.Pop(), nil
		}
		v.stack.Push(f)
		return nil
	case *ast.SequenceExpr:
		return newSequenceVisitor(t, v)
	case *ast.RangeExpr:
		return newRangeVisitor(t, v)
	}

	return v
}

func (s *stack) Pop() (val interface{}) {
	n := len(s.values)
	if n == 0 {
		return
	}
	val = s.values[n-1]
	s.values = s.values[:n-1]
	return
}

func (s *stack) Push(val interface{}) {
	s.values = append(s.values, val)
}

func (s *stack) Size() int {
	return len(s.values)
}
