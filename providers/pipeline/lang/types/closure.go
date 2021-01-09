package types

import (
	"reflect"
)

type ClosureFunc func(parameters []Object) (Object, error)

type Closure struct {
	ObjectImpl
	value ClosureFunc
}

func NewClosure(f ClosureFunc) *Closure {
	return &Closure{value: f}
}

func (c *Closure) String() string {
	return c.GetType().String()
}

func (c *Closure) GetType() reflect.Type {
	return reflect.TypeOf(c.value)
}
