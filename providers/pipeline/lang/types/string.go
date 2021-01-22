package types

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/token"
	"reflect"
	"strings"
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

func (s *String) Elem() interface{} {
	return s.value
}

func (s *String) Set(o Object) error {
	if v, isString := o.(*String); isString {
		s.value = v.value
		return nil
	} else {
		return errors.Errorf("type '%v' can not be set to string", o.GetType())
	}
}

func (s *String) InvokeOp(op token.Token, obj Object) (Object, error) {
	switch op {
	case token.ADD:
		return NewString(s.value + obj.String()), nil
	case token.EQL:
		return NewBool(s.value == obj.String()), nil
	case token.NEQ:
		return NewBool(s.value != obj.String()), nil
	default:
		return nil, fmt.Errorf("unsupported operation '%v' on type string", op)
	}
}

func (s *String) GetType() reflect.Type {
	return reflect.TypeOf(s.value)
}

func (s *String) Contains(substr string) bool {
	return strings.Contains(s.value, substr)
}

func (s *String) HasPrefix(substr string) bool {
	return strings.HasPrefix(s.value, substr)
}

func (s *String) HasSuffix(substr string) bool {
	return strings.HasSuffix(s.value, substr)
}

func (s *String) GetField(name string) (Object, error) {
	return getField(s, name)
}

func (s *String) HasField(name string) bool {
	return hasField(s, name)
}

func (s *String) InvokeFunc(name string, args map[string]Object) (Object, error) {
	return invokeFunc(s, name, args)
}

func (s *String) SetField(field string, value Object) error {
	return setField(s, field, value)
}
