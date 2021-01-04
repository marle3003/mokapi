package types

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

type Array struct {
	value []Object
}

func NewArray() *Array {
	return &Array{value: []Object{}}
}

func (a *Array) Value() interface{} {
	result := make([]interface{}, len(a.value))
	for i, value := range a.value {
		if v, ok := value.(ValueType); ok {
			result[i] = v.Value()
		}
	}
	return result
}

func (a *Array) GetIndex(index int) (Object, error) {
	if index < len(a.value) {
		return a.value[index], nil
	}

	return nil, fmt.Errorf("syntax error: index '%v' out of range", index)
}

func (a *Array) SetValue(i interface{}) error {
	v, ok := i.([]Object)
	if !ok {
		return fmt.Errorf("syntax error: unable to cast object of type %v to array", reflect.TypeOf(i))
	}
	a.value = v
	return nil
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

func (a *Array) Equals(obj Object) bool {
	if other, ok := obj.(*Array); ok {
		if len(a.value) != len(other.value) {
			return false
		}

		for i, v := range a.value {
			if !v.(Object).Equals(other.value[i].(Object)) {
				return false
			}
		}
		return true
	}
	return false
}

func (a *Array) GetType() reflect.Type {
	return reflect.TypeOf(a.value)
}

func (a *Array) Invoke(path *Path, args []Object) (Object, error) {
	switch path.Head() {
	case "":
		return a, nil
	case "*":
		if !path.MoveNext() {
			return a, nil
		}
		if path.Head() == "find" {
			invokeFind(a, args)
		}
		result := NewArray()
		for _, v := range a.value {
			obj, err := v.Invoke(path, args)
			if err != nil {
				return nil, err
			}
			result.Add(obj)
		}
		return result, nil
	case "**":
		path.MoveNext()
		result := NewArray()
		for o := range a.Iterator() {
			if len(path.Head()) == 0 {
				result.Add(o)
			} else {
				if path.Head() == "find" {
					if list, ok := o.(Collection); ok {
						found, _ := invokeFind(list, args)
						if found != nil {
							return found, nil
						}
					} else {
						m, _ := invokePredicate(o, args)
						if m {
							return o, nil
						}
					}
				} else if path.Head() == "findAll" {
					if list, ok := o.(Collection); ok {
						findResult, _ := invokeFindAll(list, args)
						for _, r := range findResult {
							result.Add(r)
						}
					} else {
						m, _ := invokePredicate(o, args)
						if m {
							result.Add(o)
						}
					}
				} else {
					v, err := o.Invoke(path.Copy(), args)
					if err != nil {
						continue
					}

					result.Add(v)
				}
			}
		}
		return result, nil
	}

	index, err := strconv.Atoi(path.Head())
	if err != nil {
		return nil, errors.Errorf("member '%v' in path '%v' is not defined on type array", path.Head(), path)
	}

	if index >= len(a.value) {
		return nil, errors.Errorf("index out of bounds: index: %v, Size: %v", index, len(a.value))
	}
	if path.MoveNext() {
		return a.value[index].Invoke(path, args)
	}
	return a.value[index], nil
}

func (a *Array) Iterator() chan Object {
	ch := make(chan Object)
	go func() {
		defer close(ch)

		for _, v := range a.value {
			if i, ok := v.(Iterator); ok {
				for o := range i.Iterator() {
					ch <- o
				}
			}
			ch <- v
		}
	}()
	return ch
}
