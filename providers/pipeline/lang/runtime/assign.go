package runtime

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/token"
	"mokapi/providers/pipeline/lang/types"
)

type assignVisitor struct {
	scope  *ast.Scope
	stack  *stack
	assign *ast.Assignment
	outer  visitor
}

func (v *assignVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil && !v.outer.hasErrors() {
		return v.outer.Visit(node)
	}
	var err error

	val := v.stack.Pop()

	if v.assign.Tok == token.DEFINE {
		v.stack.Pop() // should be nil
		if ident, isIdent := v.assign.Lhs.(*ast.Ident); isIdent {
			v.scope.SetSymbol(ident.Name, val)
			return nil
		} else {
			v.outer.addError(errors.Errorf("expected identifier on left side of :="))
			return nil
		}
	}

	if _, isPath := v.assign.Lhs.(*ast.PathExpr); isPath {
		fieldName := v.stack.Pop().String()
		obj := v.stack.Pop()
		if obj == nil {
			v.outer.addError(errors.Errorf("null reference"))
			return nil
		}
		var fieldVal types.Object
		if v.assign.Tok != token.ASSIGN {
			fieldVal, err = obj.GetField(fieldName)
			if err != nil {
				v.outer.addError(errors.Errorf("null reference"))
				return nil
			}
			val, err = v.getValue(fieldVal, val)
			if err != nil {
				v.outer.addError(err)
				return nil
			}
		} else {
			obj.SetField(fieldName, val)
		}
	} else {
		obj := v.stack.Pop()

		if v.assign.Tok != token.ASSIGN {
			val, err = v.getValue(obj, val)
			if err != nil {
				v.outer.addError(err)
				return nil
			}
		} else {
			if obj == nil {
				v.outer.addError(errors.Errorf("unable to assign value '%v'", val))
				return nil
			} else {
				obj.Set(val)
			}
		}
	}

	//reflect.Indirect(reflect.ValueOf(obj)).Set(reflect.ValueOf(val))
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
