package runtime

import (
	"mokapi/providers/workflow/ast"
	"mokapi/providers/workflow/functions"
	"strings"
)

type callVisitor struct {
	outer *visitor
	args  int
	call  *ast.CallExpr
}

func newCallVisitor(outer *visitor) *callVisitor {
	return &callVisitor{outer: outer}
}

func (v *callVisitor) Visit(e ast.Expression) ast.Visitor {
	if len(v.outer.errors) > 0 {
		return nil
	}

	if e != nil {
		switch t := e.(type) {
		case *ast.Identifier:
			f, ok := v.outer.ctx.Functions[t.Name]
			if !ok {
				if strings.ToLower(t.Name) == "true" {
					v.outer.stack.Push(true)
				} else if strings.ToLower(t.Name) == "false" {
					v.outer.stack.Push(false)
				} else if strings.ToLower(t.Name) == "null" {
					v.outer.stack.Push(nil)
				} else if x, ok := v.outer.vars[t.Name]; ok {
					v.outer.stack.Push(x)
				} else {
					i, err := v.outer.ctx.Context.Get(t.Name)
					if err != nil {
						v.outer.errors.Add(t.Pos(), err.Error())
						return nil
					}
					v.outer.stack.Push(i)
				}
			} else {
				v.outer.stack.Push(f)
			}
			return nil
		case *ast.CallExpr:
			if v.call == nil {
				v.call = t
				v.args = len(t.Args)
				return v
			} else {
				c := newCallVisitor(v.outer)
				return c.Visit(t)
			}
		}
		return v.outer.Visit(e)
	}

	args := make([]interface{}, v.args)
	for i := v.args - 1; i >= 0; i-- {
		args[i] = v.outer.stack.Pop()
	}

	f := v.outer.stack.Pop().(functions.Function)
	if f == nil {
		v.outer.errors.Addf(v.call.Pos(), "expected function")
		return nil
	}
	o, _ := f(args...)
	v.outer.stack.Push(o)

	return nil
}
