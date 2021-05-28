package ast

type Visitor interface {
	Visit(e Expression) Visitor
}

func Walk(v Visitor, e Expression) {
	if v = v.Visit(e); v == nil {
		return
	}

	switch n := e.(type) {
	case *Selector:
		Walk(v, n.X)
		Walk(v, n.Selector)
	case *Unary:
		Walk(v, n.X)
	case *Binary:
		Walk(v, n.Lhs)
		Walk(v, n.Rhs)
	case *CallExpr:
		Walk(v, n.Fun)
		for _, a := range n.Args {
			Walk(v, a)
		}
	case *Closure:
		Walk(v, n.Func)
		for _, a := range n.Args {
			Walk(v, a)
		}
	case *SequenceExpr:
		for _, val := range n.Values {
			Walk(v, val)
		}
	case *RangeExpr:
		Walk(v, n.Start)
		Walk(v, n.End)
	}
	v.Visit(nil)
}
