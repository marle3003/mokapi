package types

import (
	"fmt"
	"reflect"
)

type String struct {
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

func (s *String) Operator(op Operator, obj Object) (Object, error) {
	switch op {
	case Addition:
		return NewString(s.value + obj.String()), nil
	default:
		return nil, fmt.Errorf("unsupported operation '%v' on type string", op)
	}
}

func (s *String) GetType() reflect.Type {
	return reflect.TypeOf(s.value)
}
