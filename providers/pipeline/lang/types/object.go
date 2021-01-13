package types

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/token"
	"reflect"
)

type Object interface {
	String() string
	GetType() reflect.Type
	GetField(string) (Object, error)
	HasField(string) bool
	InvokeFunc(string, map[string]Object) (Object, error)
	InvokeOp(op token.Token, obj Object) (Object, error)
	SetField(string, Object) error
	Set(Object) error
	Elem() interface{}
}

type ObjectImpl struct{}

func (o *ObjectImpl) String() string {
	return o.GetType().String()
}

func (o *ObjectImpl) GetType() reflect.Type {
	return reflect.TypeOf(o)
}

func (o *ObjectImpl) GetField(name string) (Object, error) {
	return getField(o, name)
}

func (o *ObjectImpl) HasField(name string) bool {
	return hasField(o, name)
}

func (o *ObjectImpl) InvokeFunc(name string, args map[string]Object) (Object, error) {
	return invokeFunc(o, name, args)
}

func (o *ObjectImpl) InvokeOp(op token.Token, _ Object) (Object, error) {
	return nil, errors.Errorf("type %v does not support operator %v", o.GetType(), op)
}

func (o *ObjectImpl) Elem() interface{} {
	return nil
}

func (o *ObjectImpl) SetField(name string, v Object) error {
	return errors.Errorf("set field is not supported")
}

func (o *ObjectImpl) Set(v Object) error {
	return errors.Errorf("set value is not supported")
}
