package runtime

import (
	"fmt"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/types"
)

type pathVisitor struct {
	outer visitor
	path  *ast.PathExpr
}

func newPathVisitor(path *ast.PathExpr, outer visitor) *pathVisitor {
	return &pathVisitor{path: path, outer: outer}
}

func (v *pathVisitor) Visit(node ast.Node) ast.Visitor {
	if v.outer.HasErrors() {
		return nil
	}
	if node != nil {
		switch n := node.(type) {
		case *ast.Ident:
			v.outer.Stack().Push(types.NewString(n.Name))
			return nil
		case *ast.IndexExpr:
			return v
		}
		return v.outer.Visit(node)
	}

	n := len(v.path.Args)
	args := make(map[string]types.Object, n)
	for i := n - 1; i >= 0; i-- {
		kv := v.outer.Stack().Pop().(*types.KeyValuePair)
		argName := kv.Key
		if len(argName) == 0 {
			argName = fmt.Sprintf("%v", i)
		}
		args[argName] = kv.Value
	}

	path := v.outer.Stack().Pop()

	obj := v.outer.Stack().Pop()
	if obj == nil {
		v.outer.AddErrorf(v.path.Path.Pos(), "can not access '%v' on nil value", path)
		return nil
	}
	if ident, isIdent := v.path.X.(*ast.Ident); isIdent {
		if o, exists := v.outer.Scope().Symbol(ident.Name); exists {
			obj = o
		} else {
			v.outer.AddErrorf(v.path.Pos(), "unable to resolve identifier %v", obj.String())
		}
	}

	p, ok := obj.(*types.Path)
	if !ok {
		p = types.NewPath(obj)
	}

	val, err := p.Resolve(path.String(), args)
	if err != nil {
		if v.path.Lhs {
			v.outer.Stack().Push(obj)
			v.outer.Stack().Push(path)
		} else {
			v.outer.AddError(v.path.Path.Pos(), err.Error())
		}
	} else {
		v.outer.Stack().Push(val)
	}

	return nil
}
