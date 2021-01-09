package types

import (
	"github.com/pkg/errors"
)

type Predicate func(Object) (bool, error)

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
