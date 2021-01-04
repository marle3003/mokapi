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
	return closure.toPredicate()(obj)
}
