package types

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/token"
	"reflect"
	"strings"
)

type Array struct {
	value []Object
}

func NewArray() *Array {
	return &Array{value: []Object{}}
}

func (a *Array) Elem() interface{} {
	var r []interface{}
	for _, i := range a.value {
		r = append(r, i.Elem())
	}
	return r
}

func (a *Array) Index(index int) (Object, error) {
	if index < len(a.value) {
		return a.value[index], nil
	}

	return nil, fmt.Errorf("syntax error: index '%v' out of range", index)
}

func (a *Array) String() string {
	sb := strings.Builder{}
	sb.WriteString("[")
	for i, o := range a.value {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(o.String())
	}
	sb.WriteString("]")
	return sb.String()
}

func (a *Array) Add(obj Object) {
	a.value = append(a.value, obj)
}

func (a *Array) AddRange(o *Array) {
	for _, i := range o.value {
		a.value = append(a.value, i)
	}
}

func (a *Array) Contains(obj Object) (*Bool, error) {
	for _, i := range a.value {
		r, err := i.InvokeOp(token.EQL, obj)
		if b, ok := r.(*Bool); err == nil && ok && b.value {
			return b, nil
		}
	}
	return NewBool(false), nil
}

func (a *Array) Set(o Object) error {
	if v, isArray := o.(*Array); isArray {
		a.value = v.value
		return nil
	} else {
		return errors.Errorf("type '%v' can not be set to array", o.GetType())
	}
}

func (a *Array) Find(match Predicate) (Object, error) {
	for _, item := range a.value {
		if matches, err := match(item); err == nil && matches {
			return item, nil
		} else if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (a *Array) FindAll(match Predicate) (*Array, error) {
	result := NewArray()
	for _, item := range a.value {
		if matches, err := match(item); err == nil && matches {
			result.Add(item)
		} else if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (a *Array) GetType() reflect.Type {
	return reflect.TypeOf(a.value)
}

func (a *Array) Children() *Array {
	return a
}

func (a *Array) InvokeOp(op token.Token, _ Object) (Object, error) {
	return nil, errors.Errorf("type array does not support operator %v", op)
}

func (a *Array) GetField(name string) (Object, error) {
	return getField(a, name)
}

func (a *Array) HasField(name string) bool {
	return hasField(a, name)
}

func (a *Array) InvokeFunc(name string, args map[string]Object) (Object, error) {
	return invokeFunc(a, name, args)
}

func (a *Array) SetField(field string, value Object) error {
	return setField(a, field, value)
}
