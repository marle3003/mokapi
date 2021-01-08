package types

import (
	"fmt"
	"github.com/pkg/errors"
	"math"
	"reflect"
)

type Number struct {
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

func (n *Number) Operator(op Operator, obj Object) (Object, error) {
	if other, ok := obj.(*Number); ok {
		switch op {
		case Addition:
			return NewNumber(n.value + other.value), nil
		case Subtraction:
			return NewNumber(n.value - other.value), nil
		case Multiplication:
			return NewNumber(n.value * other.value), nil
		case Division:
			if other.value == 0 {
				return nil, errors.New("divide by zero")
			}
			return NewNumber(n.value / other.value), nil
		case Remainder:
			if n.value != math.Trunc(n.value) || other.value != math.Trunc(other.value) {
				return nil, errors.New("unable to use operator '%' on floating number")

			}
			v := float64(int64(n.value) % int64(other.value))
			return NewNumber(v), nil

		default:
			return nil, fmt.Errorf("unsupported operation '%v' on type number", op)
		}
	}
	return nil, fmt.Errorf("operator '%v' is not defined on %v", op, reflect.TypeOf(obj))

}

func (n *Number) GetType() reflect.Type {
	return reflect.TypeOf(n.value)
}
