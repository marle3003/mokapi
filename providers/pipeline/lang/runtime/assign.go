package runtime

import (
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/token"
	"mokapi/providers/pipeline/lang/types"
)

type assignVisitor struct {
	assign *ast.Assignment
	outer  visitor
}

func newAssignVisitor(assign *ast.Assignment, outer visitor) *assignVisitor {
	return &assignVisitor{assign: assign, outer: outer}
}

func (v *assignVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil && !v.outer.HasErrors() {
		return v.outer.Visit(node)
	}
	var err error

	val := v.outer.Stack().Pop()

	if v.assign.Tok == token.DEFINE {
		v.outer.Stack().Pop() // should be nil
		if ident, isIdent := v.assign.Lhs.(*ast.Ident); isIdent {
			v.outer.Scope().SetSymbol(ident.Name, val)
			return nil
		} else {
			v.outer.AddError(v.assign.TokPos, "expected identifier on left side of :=")
			return nil
		}
	}

	if p, isPath := v.assign.Lhs.(*ast.PathExpr); isPath {
		fieldName := v.outer.Stack().Pop().String()
		obj := v.outer.Stack().Pop()
		if obj == nil {
			v.outer.AddError(p.X.Pos(), "nil reference")
			return nil
		}
		var fieldVal types.Object
		if v.assign.Tok != token.ASSIGN {
			fieldVal, err = obj.GetField(fieldName)
			if err != nil {
				v.outer.AddError(p.X.Pos(), "nil reference")
				return nil
			}
			val, err = v.getValue(fieldVal, val)
			if err != nil {
				v.outer.AddError(p.Path.Pos(), "nil reference")
				return nil
			}
		} else {
			obj.SetField(fieldName, val)
		}
	} else {
		obj := v.outer.Stack().Pop()

		if v.assign.Tok != token.ASSIGN {
			val, err = v.getValue(obj, val)
			if err != nil {
				v.outer.AddError(p.Path.Pos(), "nil reference")
				return nil
			}
		} else {
			if obj == nil {
				v.outer.AddErrorf(v.assign.TokPos, "unable to assign value '%v'", val)
				return nil
			} else {
				obj.Set(val)
			}
		}
	}

	return nil
}

func (v *assignVisitor) getValue(val types.Object, newVal types.Object) (types.Object, error) {
	switch v.assign.Tok {
	case token.ADD_ASSIGN:
		return val.InvokeOp(token.ADD, newVal)
	case token.SUB_ASSIGN:
		return val.InvokeOp(token.SUB, newVal)
	case token.MUL_ASSIGN:
		return val.InvokeOp(token.MUL, newVal)
	case token.QUO_ASSIGN:
		return val.InvokeOp(token.QUO, newVal)
	case token.REM_ASSIGN:
		return val.InvokeOp(token.REM, newVal)
	}
	return newVal, nil
}
