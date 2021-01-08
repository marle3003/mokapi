package types

import (
	"github.com/pkg/errors"
)

type Predicate func(Object) (bool, error)

func invokePredicate(obj Object, args []Object) (bool, error) {
	if len(args) != 1 {
		return false, errors.Errorf("invalid number of arguments")
	}
	closure, ok := args[0].(*Closure)
	if !ok {
		return false, errors.Errorf("invalid type of argument: expected: Closure")
	}
	return newPredicate(closure)(obj)
}

func newPredicate(c *Closure) Predicate {
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
