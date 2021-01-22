package types

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/token"
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

func (b *Bool) Set(o Object) error {
	if v, isBool := o.(*Bool); isBool {
		b.value = v.value
		return nil
	} else {
		return errors.Errorf("type '%v' can not be set to bool", o.GetType())
	}
}

func (b *Bool) Elem() interface{} {
	return b.value
}

func (b *Bool) InvokeOp(op token.Token, obj Object) (Object, error) {
	if other, ok := obj.(*Bool); ok {
		switch op {
		case token.LAND:
			return NewBool(b.value && other.value), nil
		case token.LOR:
			return NewBool(b.value || other.value), nil
		}
	}
	return nil, fmt.Errorf("unsupported operation '%v' on type bool", op)
}

func (b *Bool) GetType() reflect.Type {
	return reflect.TypeOf(b.value)
}

func (b *Bool) GetField(name string) (Object, error) {
	return getField(b, name)
}

func (b *Bool) HasField(name string) bool {
	return hasField(b, name)
}

func (b *Bool) InvokeFunc(name string, args map[string]Object) (Object, error) {
	return invokeFunc(b, name, args)
}

func (b *Bool) SetField(field string, value Object) error {
	return setField(b, field, value)
}
