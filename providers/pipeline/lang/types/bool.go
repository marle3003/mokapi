package types

import (
	"fmt"
	"mokapi/providers/pipeline/lang"
	"reflect"
)

type Bool struct {
	ObjectImpl
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

func (b *Bool) InvokeOp(op lang.Token, obj Object) (Object, error) {
	if other, ok := obj.(*Bool); ok {
		switch op {
		case lang.LAND:
			return NewBool(b.value && other.value), nil
		case lang.LOR:
			return NewBool(b.value || other.value), nil
		}
	}
	return nil, fmt.Errorf("unsupported operation '%v' on type bool", op)
}

func (b *Bool) GetType() reflect.Type {
	return reflect.TypeOf(b.value)
}
