package types

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/token"
	"reflect"
)

type Reference struct {
	value interface{}
}

func NewReference(i interface{}) *Reference {
	return &Reference{value: i}
}

func (r *Reference) String() string {
	return fmt.Sprintf("%v", r.value)
}

func (r *Reference) GetType() reflect.Type {
	return reflect.TypeOf(r)
}

func (r *Reference) GetField(name string) (Object, error) {
	return getField(r.value, name)
}

func (r *Reference) SetField(name string, value Object) error {
	return setField(r.value, name, value)
}

func (r *Reference) HasField(name string) bool {
	return hasField(r.value, name)
}

func (r *Reference) Elem() interface{} {
	return r.value
}

func (r *Reference) InvokeFunc(name string, args map[string]Object) (Object, error) {
	return invokeFunc(r.value, name, args)
}

func (r *Reference) InvokeOp(op token.Token, _ Object) (Object, error) {
	return nil, errors.Errorf("type reference does not support operator %v", op)
}

func (r *Reference) Set(o Object) error {
	r.value = o.Elem()
	return nil
}
