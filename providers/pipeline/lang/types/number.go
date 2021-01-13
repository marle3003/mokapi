package types

import (
	"fmt"
	"github.com/pkg/errors"
	"math"
	"mokapi/providers/pipeline/lang/token"
	"reflect"
)

type Number struct {
	ObjectImpl
	value float64
}

func NewNumber(f float64) *Number {
	return &Number{value: f}
}

func (n *Number) Elem() interface{} {
	return n.value
}

func (n *Number) GetField(name string) (Object, error) {
	return getField(n, name)
}

func (b *Number) Set(o Object) error {
	if v, isNum := o.(*Number); isNum {
		b.value = v.value
		return nil
	} else {
		return errors.Errorf("type '%v' can not be set to number", o.GetType())
	}
}

func (n *Number) String() string {
	return fmt.Sprintf("%v", n.value)
}

func (n *Number) InvokeOp(op token.Token, obj Object) (Object, error) {
	if other, ok := obj.(*Number); ok {
		switch op {
		case token.ADD:
			return NewNumber(n.value + other.value), nil
		case token.SUB:
			return NewNumber(n.value - other.value), nil
		case token.MUL:
			return NewNumber(n.value * other.value), nil
		case token.QUO:
			if other.value == 0 {
				return nil, errors.New("divide by zero")
			}
			return NewNumber(n.value / other.value), nil
		case token.REM:
			if n.value != math.Trunc(n.value) || other.value != math.Trunc(other.value) {
				return nil, errors.New("unable to use operator '%' on floating number")

			}
			v := float64(int64(n.value) % int64(other.value))
			return NewNumber(v), nil
		case token.EQL:
			return NewBool(n.value == other.value), nil

		default:
			return nil, fmt.Errorf("unsupported operation '%v' on type number", op)
		}
	}
	return nil, fmt.Errorf("operator '%v' is not defined on %v", op, reflect.TypeOf(obj))

}
