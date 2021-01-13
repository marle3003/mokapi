package ast

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
		if n.When != nil {
			Walk(v, n.When)
		}
		Walk(v, n.Steps)
	case *StepBlock:
		for _, s := range n.Statments {
			Walk(v, s)
		}
	case *Block:
		for _, s := range n.Stmts {
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
		Walk(v, n.X)
	case *Call:
		Walk(v, n.Func)
		for _, a := range n.Args {
			Walk(v, a)
		}
	case *Argument:
		Walk(v, n.Name)
		Walk(v, n.Value)
	case *IndexExpr:
		Walk(v, n.X)
		Walk(v, n.Index)
	case *PathExpr:
		Walk(v, n.X)
		Walk(v, n.Path)
		for _, a := range n.Args {
			Walk(v, a)
		}
	case *Closure:
		for _, p := range n.Params {
			Walk(v, p)
		}
		Walk(v, n.Block)
	case *DeclStmt:
		Walk(v, n.Name)
		Walk(v, n.Type)
	case *Ident:
	}
	v.Visit(nil)
}
