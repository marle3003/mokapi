package runtime

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang"
	"mokapi/providers/pipeline/lang/types"
	"reflect"
)

type assignVisitor struct {
	scope  *Scope
	stack  *stack
	assign *lang.Assignment
	outer  visitor

	setValue    func(types.Object)
	val         types.Object
	lhsExecuted bool
}

func (v *assignVisitor) Visit(node lang.Node) lang.Visitor {
	if node != nil && !v.outer.hasErrors() {
		if v.lhsExecuted {
			return v.outer.Visit(node)
		} else {
			switch n := node.(type) {
			case *lang.SymbolRef:
				if _, ok := v.scope.Symbol(n.Name); !ok && v.assign.Tok != lang.DEFINE {
					v.outer.addError(errors.Errorf("unresovled reference '%v'", n.Name))
					return nil
				} else if ok {
					v.val, _ = v.scope.Symbol(n.Name)
				}

				v.setValue = func(obj types.Object) {
					v.scope.SetSymbol(n.Name, obj)
				}
				v.lhsExecuted = true
				return nil
			default:
				v.outer.addError(errors.Errorf("unsupported operand on lhs %v", reflect.TypeOf(n)))
			}
			return v.outer.Visit(node)
		}
	}

	val := v.stack.Pop()

	var err error
	switch v.assign.Tok {
	case lang.ADD_ASSIGN:
		val, err = v.val.InvokeOp(lang.ADD, val)
	case lang.SUB_ASSIGN:
		val, err = v.val.InvokeOp(lang.SUB, val)
	case lang.MUL_ASSIGN:
		val, err = v.val.InvokeOp(lang.MUL, val)
	case lang.QUO_ASSIGN:
		val, err = v.val.InvokeOp(lang.QUO, val)
	case lang.REM_ASSIGN:
		val, err = v.val.InvokeOp(lang.REM, val)
	}
	if err != nil {
		v.outer.addError(err)
		return nil
	}

	if v.setValue == nil {
		v.outer.addError(errors.Errorf("unable to assign value '%v'", val))
	} else {
		v.setValue(val)
	}

	return nil
}
