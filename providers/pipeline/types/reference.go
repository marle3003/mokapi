package types

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
)

type Reference struct {
	value interface{}
}

func NewReference(i interface{}) *Reference {
	return &Reference{value: i}
}

func (r *Reference) Equals(obj Object) bool {
	if other, ok := obj.(*Reference); ok {
		return r.value == other.value
	}
	return false
}

func (r *Reference) String() string {
	return fmt.Sprintf("%v", r.value)
}

func (r *Reference) GetType() reflect.Type {
	return reflect.TypeOf(r)
}

func (r *Reference) Invoke(path *Path, args []Object) (Object, error) {
	return invokeMember(r.value, path, args)
}

func (r *Reference) Set(name string, value Object) error {
	return errors.Errorf("not implemented")
}

func (r *Reference) Value() interface{} {
	return r.value
}
func (r *Reference) SetValue(i interface{}) error {
	r.value = i
	return nil
}

func (r *Reference) Operator(_ Operator, _ Object) (Object, error) {
	return nil, errors.Errorf("not implemented")
}

func (r *Reference) Iterator() chan Object {
	return iterator(r.value)
}
