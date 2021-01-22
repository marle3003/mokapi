package types

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/token"
	"reflect"
)

type ClosureFunc func(parameters []Object) (Object, error)

type Closure struct {
	value ClosureFunc
}

func NewClosure(f ClosureFunc) *Closure {
	return &Closure{value: f}
}

func (c *Closure) String() string {
	return c.GetType().String()
}

func (c Closure) Elem() interface{} {
	return c.value
}

func (c *Closure) GetType() reflect.Type {
	return reflect.TypeOf(c.value)
}

func (c *Closure) InvokeOp(op token.Token, _ Object) (Object, error) {
	return nil, errors.Errorf("type array does not support operator %v", op)
}

func (c *Closure) GetField(name string) (Object, error) {
	return getField(c, name)
}

func (c *Closure) HasField(name string) bool {
	return hasField(c, name)
}

func (c *Closure) InvokeFunc(name string, args map[string]Object) (Object, error) {
	return invokeFunc(c, name, args)
}

func (c *Closure) SetField(field string, value Object) error {
	return setField(c, field, value)
}

func (c *Closure) Set(o Object) error {
	if v, isClosure := o.(*Closure); isClosure {
		c.value = v.value
		return nil
	} else {
		return errors.Errorf("type '%v' can not be set to array", o.GetType())
	}
}
