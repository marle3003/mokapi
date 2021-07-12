package functions

import (
	"fmt"
	"reflect"
)

func Any(args ...interface{}) (interface{}, error) {
	source := args[0]
	if source == nil {
		return nil, fmt.Errorf("any: source parameter is null")
	}
	p := newPredicate(args[1].(Function))
	switch reflect.TypeOf(source).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(source)

		for i := 0; i < s.Len(); i++ {
			v := s.Index(i).Interface()
			if found, _ := p(v); found {
				return true, nil
			}
		}
		return false, nil
	default:
		return nil, fmt.Errorf("any: unexpected source type %t", source)
	}
}
