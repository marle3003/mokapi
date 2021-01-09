package types

import (
	"fmt"
	"github.com/pkg/errors"
	"math"
	"mokapi/providers/pipeline/lang"
	"reflect"
)

type Number struct {
	ObjectImpl
	value float64
}

func NewNumber(f float64) *Number {
	return &Number{value: f}
}

func (n *Number) GetField(name string) (Object, error) {
	return getField(n, name)
}

func (n *Number) String() string {
	return fmt.Sprintf("%v", n.value)
}

func (n *Number) InvokeOp(op lang.Token, obj Object) (Object, error) {
	if other, ok := obj.(*Number); ok {
		switch op {
		case lang.ADD:
			return NewNumber(n.value + other.value), nil
		case lang.SUB:
			return NewNumber(n.value - other.value), nil
		case lang.MUL:
			return NewNumber(n.value * other.value), nil
		case lang.QUO:
			if other.value == 0 {
				return nil, errors.New("divide by zero")
			}
			return NewNumber(n.value / other.value), nil
		case lang.REM:
			if n.value != math.Trunc(n.value) || other.value != math.Trunc(other.value) {
				return nil, errors.New("unable to use operator '%' on floating number")

			}
			v := float64(int64(n.value) % int64(other.value))
			return NewNumber(v), nil
		case lang.EQL:
			return NewBool(n.value == other.value), nil

		default:
			return nil, fmt.Errorf("unsupported operation '%v' on type number", op)
		}
	}
	return nil, fmt.Errorf("operator '%v' is not defined on %v", op, reflect.TypeOf(obj))

}
