package types

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

type Expando struct {
	value map[string]Object
}

func NewExpando() *Expando {
	return &Expando{value: map[string]Object{}}
}

func (e *Expando) Invoke(path *Path, args []Object) (Object, error) {
	switch path.Head() {
	case "":
		return e, nil
	case "*":
		if !path.MoveNext() {
			return e, nil
		}
		result := NewArray()
		for _, v := range e.value {
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
		for o := range e.Iterator() {
			if len(path.Head()) == 0 {
				result.Add(o)
			} else {
				v, err := o.Invoke(path.Copy(), args)
				if err != nil {
					continue
				}
				result.Add(v)
			}
		}
		return result, nil
	}

	if v, ok := e.value[path.Head()]; ok {
		if path.MoveNext() {
			return v.Invoke(path, args)
		} else {
			return v, nil
		}
	}

	return nil, errors.Errorf("name '%v' in path '%v' not found", path.Head(), path)
}

func (e *Expando) Set(name string, obj Object) error {
	e.value[name] = obj
	return nil
}

func (e *Expando) GetEnumerator() []Object {
	result := make([]Object, len(e.value))
	index := 0
	for key, value := range e.value {
		result[index] = NewKeyValuePair(key, value)
		index++
	}
	return result
}

func (e *Expando) Value() interface{} {
	result := map[string]interface{}{}
	for key, value := range e.value {
		if v, ok := value.(ValueType); ok {
			result[key] = v.Value()
		}
	}
	return result
}

func (e *Expando) String() string {
	sb := strings.Builder{}
	sb.WriteString("{")
	counter := 0
	for k, v := range e.value {
		if counter > 0 {
			sb.WriteString(", ")
		}
		obj := v.(Object)
		sb.WriteString(fmt.Sprintf("%v: %v", k, obj.String()))
		counter++
	}
	sb.WriteString("}")
	return sb.String()
}

func (e *Expando) Equals(obj Object) bool {
	if other, ok := obj.(*Expando); ok {
		if len(e.value) != len(other.value) {
			return false
		}

		for k, v := range e.value {
			if ov, ok := other.value[k]; ok {
				if !v.(Object).Equals(ov.(Object)) {
					return false
				}
			} else {
				return false
			}
		}
		return true
	}
	return false
}

func (e *Expando) SetValue(i interface{}) error {
	if m, ok := i.(map[string]interface{}); ok {
		e.value = map[string]Object{}
		for k, v := range m {
			if obj, ok := v.(Object); ok {
				e.value[k] = obj
			} else {
				obj, err := Convert(i)
				if err != nil {
					return err
				}
				e.value[k] = obj
			}
		}
	}
	return nil
}

func (e *Expando) GetType() reflect.Type {
	return reflect.TypeOf(e.value)
}

func (e *Expando) Operator(_ Operator, _ Object) (Object, error) {
	return nil, errors.Errorf("not implemented")
}

func (e *Expando) Iterator() chan Object {
	ch := make(chan Object)
	go func() {
		defer close(ch)

		for _, v := range e.value {
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
