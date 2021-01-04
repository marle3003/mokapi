package types

import (
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

func SetField(_ Object, _ []string, _ Object) error {
	return errors.Errorf("not implemented")
}

func InvokeMember(obj Object, segments []string, args []Object) (Object, error) {
	current := obj
	for _, name := range segments[:len(segments)-1] {
		m, err := invokeMember(current, name, args)
		if err != nil {
			return nil, err
		}
		current = m
	}

	return invokeMember(current, segments[len(segments)-1], args)
}

func getMember(obj Object, name string) (Object, error) {
	v := reflect.ValueOf(obj)
	f := v.FieldByName(name)
	if f.IsValid() {
		i := f.Interface()
		if o, ok := i.(Object); ok {
			return o, nil
		}
		return Convert(i)
	}

	return nil, errors.Errorf("type %v does not container member %v", obj.GetType(), name)
}

func invokeMember(i interface{}, name string, args []Object) (Object, error) {
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

	memberName := strings.Title(name)

	// check for field on value
	f := v.FieldByName(memberName)
	if !f.IsValid() {
		// check for field on pointer
		f = reflect.Indirect(ptr).FieldByName(memberName)
	}
	if f.IsValid() {
		i := f.Interface()
		if o, ok := i.(Object); ok {
			return o, nil
		}
		return Convert(i)
	}

	//check for method on value
	m := v.MethodByName(memberName)
	if !m.IsValid() {
		m = ptr.MethodByName(memberName)
	}
	if m.IsValid() {
		obj, err := invokeMethod(m, args)
		if err != nil {
			return nil, errors.Wrapf(err, "method %v", name)
		}
		return obj, nil
	}

	if class, ok := i.(Class); ok {
		return class.Invoke(name, args)
	}

	return nil, errors.Errorf("type %v does not contains member %v", reflect.TypeOf(i), name)
}

func invokeMethod(method reflect.Value, args []Object) (Object, error) {
	methodInfo := method.Type()

	callArgs, err := createArgs(methodInfo, args)
	if err != nil {
		return nil, err
	}

	result := method.Call(callArgs)

	if len(result) == 0 {
		return nil, nil
	}

	if len(result) >= 2 {
		if err, ok := result[1].Interface().(error); ok {
			return nil, err
		}
	}

	return Convert(result[0].Interface())
}

func createArgs(method reflect.Type, args []Object) ([]reflect.Value, error) {
	var callArgs []reflect.Value
	for i := 0; i < method.NumIn(); i++ {
		t := method.In(i)
		v := reflect.New(t).Elem()

		var value reflect.Value
		var err error
		if i < len(args) {
			value, err = ConvertFrom(args[i], t)
			if err != nil {
				return nil, err
			}
		} else {
			value = reflect.Zero(t)
		}

		v.Set(value)
		callArgs = append(callArgs, v)
	}

	return callArgs, nil
}
