package types

import (
	"fmt"
	"mokapi/providers/pipeline/lang"
	"reflect"
)

type String struct {
	ObjectImpl
	value string
}

func NewString(s string) *String {
	return &String{value: s}
}

func (s *String) String() string {
	return s.value
}

func (s *String) GetField(name string) (Object, error) {
	return getField(s, name)
}

func (s *String) InvokeOp(op lang.Token, obj Object) (Object, error) {
	switch op {
	case lang.ADD:
		return NewString(s.value + obj.String()), nil
	case lang.EQL:
		return NewBool(s.value == obj.String()), nil
	case lang.NEQ:
		return NewBool(s.value != obj.String()), nil
	default:
		return nil, fmt.Errorf("unsupported operation '%v' on type string", op)
	}
}

func (s *String) GetType() reflect.Type {
	return reflect.TypeOf(s.value)
}
