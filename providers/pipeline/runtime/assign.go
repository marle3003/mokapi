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
	lhsExecuted bool
}

func (v *assignVisitor) Visit(node lang.Node) lang.Visitor {
	if node != nil && !v.outer.hasErrors() {
		if v.lhsExecuted {
			return v.outer.Visit(node)
		} else {
			switch n := node.(type) {
			case *lang.Ident:
				if _, ok := v.scope.Symbol(n.Name); !ok && v.assign.Tok == lang.ASSIGN {
					v.outer.addError(errors.Errorf("undefined identifier '%v'", n.Name))
					return nil
				}
				v.setValue = func(obj types.Object) {
					v.scope.SetSymbol(n.Name, obj)
				}
				v.lhsExecuted = true
			default:
				v.outer.addError(errors.Errorf("unsupported operand on lhs %v", reflect.TypeOf(n)))
			}
			return v
		}
	}

	val := v.stack.Pop()
	if v.setValue == nil {
		v.outer.addError(errors.Errorf("unable to assign value '%v'", val))
	} else {
		v.setValue(val)
	}

	return nil
}
