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

func (s *String) Value() interface{} {
	return s.value
}

func (s *String) SetValue(obj interface{}) error {
	s.value = fmt.Sprintf("%v", obj)
	return nil
}

func (s *String) String() string {
	return s.value
}

func (s *String) Operator(op ArithmeticOperator, obj Object) (Object, error) {
	switch op {
	case Addition:
		return NewString(s.value + obj.String()), nil
	default:
		return nil, fmt.Errorf("unsupported operation '%v' on type string", op)
	}
}

func (s *String) Equals(obj Object) bool {
	return s.value == obj.String()
}

func (s *String) GetType() reflect.Type {
	return reflect.TypeOf(s.value)
}
