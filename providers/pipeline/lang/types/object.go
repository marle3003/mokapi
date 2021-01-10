package types

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang"
	"reflect"
)

type Object interface {
	String() string
	GetType() reflect.Type
	GetField(string) (Object, error)
	HasField(string) bool
	InvokeFunc(string, map[string]Object) (Object, error)
	InvokeOp(op lang.Token, obj Object) (Object, error)
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

func (o *ObjectImpl) InvokeOp(op lang.Token, _ Object) (Object, error) {
	return nil, errors.Errorf("type %v does not support operator %v", o.GetType(), op)
}

func (o *ObjectImpl) Elem() interface{} {
	return nil
}
