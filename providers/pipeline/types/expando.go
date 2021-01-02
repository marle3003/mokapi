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

func (e *Expando) Invoke(name string, args []Object) (Object, error) {
	if result, ok := e.value[name]; ok {
		if closure, ok := result.(*Closure); ok {
			return closure.Invoke(args)
		}
		return result, nil
	}
	return nil, errors.Errorf("does not contain member %v", name)
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

func (e *Expando) Operator(_ ArithmeticOperator, _ Object) (Object, error) {
	return nil, errors.Errorf("not implemented")
}
