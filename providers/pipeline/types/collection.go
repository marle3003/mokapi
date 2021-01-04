package types

import (
	"github.com/pkg/errors"
)

type Collection interface {
	Add(obj Object)
	Find(match Predicate) (Object, error)
	FindAll(match Predicate) ([]Object, error)
}

func invokeFind(list Collection, args []Object) (Object, error) {
	if len(args) != 1 {
		return nil, errors.Errorf("invalid number of arguments")
	}
	closure, ok := args[0].(*Closure)
	if !ok {
		return nil, errors.Errorf("invalid type of argument: expected: Closure")
	}
	return list.Find(closure.toPredicate())
}

func invokeFindAll(list Collection, args []Object) ([]Object, error) {
	if len(args) != 1 {
		return nil, errors.Errorf("invalid number of arguments")
	}
	closure, ok := args[0].(*Closure)
	if !ok {
		return nil, errors.Errorf("invalid type of argument: expected: Closure")
	}
	return list.FindAll(closure.toPredicate())
}
