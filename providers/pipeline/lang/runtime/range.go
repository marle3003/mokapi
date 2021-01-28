package runtime

import (
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/types"
	"reflect"
)

type rangeVisitor struct {
	outer visitor
	r     *ast.RangeExpr
}

func newRangeVisitor(r *ast.RangeExpr, outer visitor) *rangeVisitor {
	return &rangeVisitor{r: r, outer: outer}
}

func (v *rangeVisitor) Visit(node ast.Node) ast.Visitor {
	if v.outer.HasErrors() {
		return nil
	}
	if node != nil {
		return v.outer.Visit(node)
	}

	end := v.outer.Stack().Pop()
	start := v.outer.Stack().Pop()

	if reflect.TypeOf(start) != reflect.TypeOf(end) {
		v.outer.AddErrorf(v.r.Pos(), "unsupported type of start '%t%' with type of end '%t'", start, end)
		return nil
	}

	a := types.NewArray()
	switch n := start.(type) {
	case *types.Number:
		s := int(n.Val())
		e := int(end.(*types.Number).Val())
		for i := s; i <= e; i++ {
			a.Add(types.NewNumber(float64(i)))
		}
	}

	v.outer.Stack().Push(a)

	return nil
}
