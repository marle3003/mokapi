package functions

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
)

type Function func(args ...interface{}) (interface{}, error)

type Predicate func(interface{}) (bool, error)

func newPredicate(f Function) Predicate {
	return func(i interface{}) (bool, error) {
		r, _ := f(i)

		if b, ok := r.(bool); ok {
			return b, nil
		}
		return false, errors.Errorf("unexpected return type: expected bool")
	}
}

func Find(args ...interface{}) (interface{}, error) {
	source := args[0]
	if source == nil {
		return nil, fmt.Errorf("find: source parameter is null")
	}
	p := newPredicate(args[1].(Function))
	switch reflect.TypeOf(source).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(source)

		for i := 0; i < s.Len(); i++ {
			v := s.Index(i).Interface()
			if found, _ := p(v); found {
				return v, nil
			}
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("find: unexpected source type %t", source)
	}
}
