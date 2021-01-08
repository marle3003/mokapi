package types

import (
	"fmt"
	"reflect"
)

type Bool struct {
	value bool
}

func NewBool(b bool) *Bool {
	return &Bool{value: b}
}

func (b *Bool) String() string {
	return fmt.Sprintf("%v", b.value)
}

func (b *Bool) Val() bool {
	return b.value
}

func (b *Bool) GetField(name string) (Object, error) {
	return getField(b, name)
}

func (b *Bool) Operator(op Operator, obj Object) (Object, error) {
	if other, ok := obj.(*Bool); ok {
		switch op {
		case And:
			return NewBool(b.value && other.value), nil
		case Or:
			return NewBool(b.value || other.value), nil
		}
	}
	return nil, fmt.Errorf("unsupported operation '%v' on type bool", op)
}

func (b *Bool) GetType() reflect.Type {
	return reflect.TypeOf(b.value)
}
