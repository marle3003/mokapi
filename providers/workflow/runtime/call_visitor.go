package runtime

import (
	"mokapi/providers/workflow/ast"
	"mokapi/providers/workflow/functions"
	"strings"
)

type callVisitor struct {
	outer *visitor
	args  int
}

func newCallVisitor(outer *visitor) *callVisitor {
	return &callVisitor{outer: outer}
}

func (v *callVisitor) Visit(e ast.Expression) ast.Visitor {
	if e != nil {
		switch t := e.(type) {
		case *ast.Identifier:
			f, ok := v.outer.ctx.Functions[t.Name]
			if !ok {
				if strings.ToLower(t.Name) == "true" {
					v.outer.stack.Push(true)
				} else if strings.ToLower(t.Name) == "false" {
					v.outer.stack.Push(false)
				} else if x, ok := v.outer.vars[t.Name]; ok {
					v.outer.stack.Push(x)
				} else {
					i := v.outer.ctx.Context.Get(t.Name)
					v.outer.stack.Push(i)
				}
			} else {
				v.outer.stack.Push(f)
			}
			return nil
		case *ast.CallExpr:
			v.args = len(t.Args)
			return v
		}
		return v.outer.Visit(e)
	}

	args := make([]interface{}, v.args)
	for i := v.args - 1; i >= 0; i-- {
		args[i] = v.outer.stack.Pop()
	}

	f := v.outer.stack.Pop().(functions.Function)
	if f == nil {
		return nil
	}
	o, _ := f(args...)
	v.outer.stack.Push(o)

	return nil
}
