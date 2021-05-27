package runtime

import (
	"mokapi/providers/workflow/ast"
	"reflect"
)

type rangeVisitor struct {
	outer *visitor
	r     *ast.RangeExpr
}

func newRangeVisitor(r *ast.RangeExpr, outer *visitor) *rangeVisitor {
	return &rangeVisitor{r: r, outer: outer}
}

func (v *rangeVisitor) Visit(e ast.Expression) ast.Visitor {
	if e != nil {
		return v.outer.Visit(e)
	}

	end := v.outer.stack.Pop()
	start := v.outer.stack.Pop()

	if reflect.TypeOf(start) != reflect.TypeOf(end) {
		//v.outer.AddErrorf(v.r.Pos(), "unsupported type of start '%t%' with type of end '%t'", start, end)
		return nil
	}

	a := make([]interface{}, 0)
	switch n := start.(type) {
	case int:
		s := n
		for i := s; i <= end.(int); i++ {
			a = append(a, i)
		}
	}

	v.outer.stack.Push(a)

	return nil
}
