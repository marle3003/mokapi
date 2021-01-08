package lang

type Visitor interface {
	Visit(n Node) Visitor
}

func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *File:
		for _, p := range n.Pipelines {
			Walk(v, p)
		}
	case *Pipeline:
		for _, s := range n.Stages {
			Walk(v, s)
		}
	case *Stage:
		Walk(v, n.Steps)
	case *StepBlock:
		for _, s := range n.Statments {
			Walk(v, s)
		}
	case *Assignment:
		Walk(v, n.Lhs)
		Walk(v, n.Rhs)
	case *ExprStatement:
		Walk(v, n.X)
	case *Binary:
		Walk(v, n.Lhs)
		Walk(v, n.Rhs)
	case *Unary:
		Walk(v, n.Operand)
	case *Call:
		Walk(v, n.Func)
		for _, a := range n.Args {
			Walk(v, a)
		}
	case *Argument:
		Walk(v, n.Name)
		Walk(v, n.Value)
	case *Selector:
		Walk(v, n.X)
		Walk(v, n.Selector)
	case *IndexExpr:
		Walk(v, n.X)
		Walk(v, n.Index)
	case *PathExpr:
		Walk(v, n.X)
		for _, p := range n.Path {
			Walk(v, p)
		}
	case *Ident:
	}
	v.Visit(nil)
}
