package types

import (
	"fmt"
	"reflect"
	"strings"
)

type Expando struct {
	ObjectImpl
	value map[string]Object
}

func NewExpando() *Expando {
	return &Expando{value: map[string]Object{}}
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

func (e *Expando) Set(name string, value Object) {
	e.value[name] = value
}

func (e *Expando) GetType() reflect.Type {
	return reflect.TypeOf(e)
}

func (e *Expando) GetField(name string) (Object, error) {
	if v, ok := e.value[name]; ok {
		return v, nil
	}
	return getField(e, name)
}

func (e *Expando) HasField(name string) bool {
	if _, ok := e.value[name]; ok {
		return true
	}
	return hasField(e, name)
}
