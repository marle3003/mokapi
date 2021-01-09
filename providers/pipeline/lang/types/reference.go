package types

import (
	"fmt"
	"reflect"
)

type Reference struct {
	ObjectImpl
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

func (r *Reference) Val() interface{} {
	return r.value
}

func (r *Reference) InvokeFunc(name string, args map[string]Object) (Object, error) {
	return invokeFunc(r.value, name, args)
}
