package types

import (
	"github.com/pkg/errors"
	"reflect"
)

type Closure struct {
	value ClosureFunc
}

func NewClosure(f ClosureFunc) *Closure {
	return &Closure{value: f}
}

func (c *Closure) Invoke(path *Path, args []Object) (Object, error) {
	if len(path.Head()) > 0 {
		return nil, errors.Errorf("member '%v' in path '%v' is not defined on type closure", path.Head(), path)
	}
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

func (c *Closure) toPredicate() Predicate {
	return func(o Object) (bool, error) {
		r, err := c.value([]Object{o})
		if err != nil {
			return false, err
		}
		if b, ok := r.(*Bool); ok {
			return b.value, nil
		}

		return false, errors.Errorf("unexpected return type: expected bool")
	}
}
