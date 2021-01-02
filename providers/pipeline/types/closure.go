package types

import "reflect"

type Closure struct {
	value ClosureFunc
}

func NewClosure(f ClosureFunc) *Closure {
	return &Closure{value: f}
}

func (c *Closure) Invoke(args []Object) (Object, error) {
	return c.value(args)
}

func (c *Closure) String() string {
	return c.GetType().String()
}

func (c *Closure) GetType() reflect.Type {
	return reflect.TypeOf(c.value)
}

func (c *Closure) Equals(obj Object) bool {
	return false
}
