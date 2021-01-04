package types

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
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

func (s *String) Operator(op Operator, obj Object) (Object, error) {
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

func (s *String) Invoke(path *Path, _ []Object) (Object, error) {
	if path.Head() == "" {
		return s, nil
	}

	index, err := strconv.Atoi(path.Head())
	if err != nil {
		return nil, errors.Errorf("member '%v' in path '%v' is not defined on type string", path.Head(), path)
	}

	if index >= len(s.value) {
		return nil, errors.Errorf("index out of bounds: index: %v, Size: %v", index, len(s.value))
	}

	return NewString(string(s.value[index])), nil
}
