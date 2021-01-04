package types

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
)

type Bool struct {
	value bool
}

func NewBool(b bool) *Bool {
	return &Bool{value: b}
}

func (b *Bool) Value() interface{} {
	return b.value
}

func (b *Bool) SetValue(i interface{}) error {
	v, ok := i.(bool)
	if !ok {
		return fmt.Errorf("syntax error: unable to cast object of type %v to bool", reflect.TypeOf(i))
	}
	b.value = v
	return nil
}

func (b *Bool) String() string {
	return fmt.Sprintf("%v", b.value)
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

func (b *Bool) Equals(obj Object) bool {
	if other, ok := obj.(*Bool); ok {
		return b.value == other.value
	}
	return false
}

func (b *Bool) CompareTo(obj Object) (int, error) {
	if other, ok := obj.(*Bool); ok {
		if b.value == other.value {
			return 0, nil
		}
		if !b.value {
			return -1, nil
		}
		return 1, nil
	}
	return 0, fmt.Errorf("unable to comapre to %v", reflect.TypeOf(obj))
}

func (b *Bool) GetType() reflect.Type {
	return reflect.TypeOf(b.value)
}

func (b *Bool) Invoke(path *Path, _ []Object) (Object, error) {
	if path.Head() == "" {
		return b, nil
	}
	return nil, errors.Errorf("member '%v' in path '%v' is not defined on type bool", path.Head(), path)
}
