package types

import (
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

type Object interface {
	String() string
	GetType() reflect.Type
	GetField(string) (Object, error)
	//invokeMethod(string, []Object) (Object, error)
}

type ObjectImpl struct{}

func (o *ObjectImpl) String() string {
	return o.GetType().String()
}

func (o *ObjectImpl) GetType() reflect.Type {
	return reflect.TypeOf(o)
}

func (o *ObjectImpl) GetField(name string) (Object, error) {
	return getField(o, name)
}

func getField(i interface{}, name string) (Object, error) {
	if len(name) == 0 {
		if o, ok := i.(Object); ok {
			return o, nil
		} else {
			return Convert(i)
		}
	}
	if obj, ok := i.(Object); ok {
		return obj.GetField(name)
	}

	v := reflect.ValueOf(i)
	var ptr reflect.Value
	if v.Type().Kind() == reflect.Ptr {
		ptr = v
		v = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(i))
		temp := ptr.Elem()
		temp.Set(v)
	}

	var fieldValue interface{} = nil
	if index, err := strconv.Atoi(name); err == nil {
		if array, ok := i.(*Array); ok {
			return array.Index(index)
		}
		fieldValue = v.Index(index).Interface()
	} else {
		fieldName := strings.Title(name)

		f := v.FieldByName(fieldName)
		if !f.IsValid() {
			// check for field on pointer
			f = reflect.Indirect(ptr).FieldByName(fieldName)
		}
		if f.IsValid() {
			fieldValue = f.Interface()
		} else {
			return nil, errors.Errorf("field '%v' is not defined on type %v", name, reflect.TypeOf(i))
		}
	}

	if o, ok := fieldValue.(Object); ok {
		return o, nil
	} else {
		return Convert(fieldValue)
	}
}
