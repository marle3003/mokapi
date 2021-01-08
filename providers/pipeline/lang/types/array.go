package types

import (
	"fmt"
	"reflect"
	"strings"
)

type Array struct {
	value []Object
}

func NewArray() *Array {
	return &Array{value: []Object{}}
}

func (a *Array) GetField(name string) (Object, error) {
	return getField(a, name)
}

func (a *Array) Index(index int) (Object, error) {
	if index < len(a.value) {
		return a.value[index], nil
	}

	return nil, fmt.Errorf("syntax error: index '%v' out of range", index)
}

func (a *Array) Set(v []Object) {
	a.value = v
}

func (a *Array) String() string {
	sb := strings.Builder{}
	sb.WriteString("[")
	for i, o := range a.value {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(o.String())
	}
	sb.WriteString("]")
	return sb.String()
}

func (a *Array) Operator(op Operator, obj Object) (Object, error) {
	return nil, fmt.Errorf("unsupported operation '%v' on type array", op)
}

func (a *Array) Add(obj Object) {
	a.value = append(a.value, obj)
}

func (a *Array) Find(match Predicate) (Object, error) {
	for _, item := range a.value {
		if matches, err := match(item); err == nil {
			if matches {
				return item, nil
			}
		} else {
			return nil, err
		}
	}
	return nil, nil
}

func (a *Array) FindAll(match Predicate) ([]Object, error) {
	result := make([]Object, 0)
	for _, item := range a.value {
		if matches, err := match(item); err == nil {
			if matches {
				result = append(result, item)
			}
		} else {
			return nil, err
		}
	}
	return result, nil
}

func (a *Array) GetType() reflect.Type {
	return reflect.TypeOf(a.value)
}

func (a *Array) Children() *Array {
	return a
}

//func (a *Array) depthFirst() Iterator {
//	ch := make(chan Object)
//	go func() {
//		defer close(ch)
//
//		for _, v := range a.value {
//			if list, ok := v.(Collection); ok {
//				for o := range list.depthFirst() {
//					ch <- o
//				}
//			}
//			ch <- v
//		}
//	}()
//	return ch
//}
