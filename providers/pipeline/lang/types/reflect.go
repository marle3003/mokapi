package types

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

func hasField(i interface{}, name string) bool {
	if len(name) == 0 {
		return false
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

	if _, err := strconv.Atoi(name); err == nil {
		if _, ok := i.(*Array); ok {
			return true
		}
		t := reflect.TypeOf(i)
		return t.Kind() == reflect.Slice
	}
	fieldName := strings.Title(name)
	f := v.FieldByName(fieldName)
	if !f.IsValid() {
		// check for field on pointer
		f = reflect.Indirect(ptr).FieldByName(fieldName)
		return f.IsValid()
	} else {
		return true
	}
}

func setField(i interface{}, name string, value Object) error {
	if len(name) == 0 {
		return errors.Errorf("invalid empty field name")
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

	if index, err := strconv.Atoi(name); err == nil {
		if array, ok := i.(*Array); ok {
			array.value[index] = value
		}
		if v.Kind() == reflect.Slice {
			return setValue(v.Index(index), value)
		} else {
			return errors.Errorf("index operator with '%v' used on %v", index, reflect.TypeOf(i))
		}
	} else {
		fieldName := strings.Title(name)

		f := v.FieldByName(fieldName)
		if !f.IsValid() {
			// check for field on pointer
			f = reflect.Indirect(ptr).FieldByName(fieldName)
		}
		if f.IsValid() {
			return setValue(f, value)
		} else {
			return errors.Errorf("field '%v' is not defined on type %v", name, reflect.TypeOf(i))
		}
	}
}

func setValue(field reflect.Value, value Object) error {
	o := reflect.TypeOf((*Object)(nil)).Elem()
	if field.Type().Implements(o) {
		field.Set(reflect.ValueOf(value))
	} else {
		v, err := ConvertFrom(value, field.Type())
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(v))
	}
	return nil
}

func getField(i interface{}, name string) (Object, error) {
	if len(name) == 0 {
		if o, ok := i.(Object); ok {
			return o, nil
		} else {
			return Convert(i)
		}
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
		if v.Kind() == reflect.Slice {
			fieldValue = v.Index(index).Interface()
		} else {
			return nil, errors.Errorf("index operator with '%v' used on %v", index, reflect.TypeOf(i))
		}
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

func invokeFunc(i interface{}, name string, args map[string]Object) (Object, error) {
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

	funcName := strings.Title(name)

	//check for method on value
	m := v.MethodByName(funcName)
	if !m.IsValid() {
		m = ptr.MethodByName(funcName)
	}
	if m.IsValid() {
		obj, err := invoke(m, args)
		if err != nil {
			return nil, errors.Wrapf(err, "func %v", funcName)
		}
		return obj, nil
	}

	return nil, errors.Errorf("func '%v' is not defined on type %v", name, reflect.TypeOf(i))
}

func invoke(f reflect.Value, args map[string]Object) (Object, error) {
	fInfo := f.Type()

	callArgs, err := createArgs(fInfo, args)
	if err != nil {
		return nil, err
	}

	result := f.Call(callArgs)

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

func createArgs(f reflect.Type, args map[string]Object) ([]reflect.Value, error) {
	var callArgs []reflect.Value
	for i := 0; i < f.NumIn(); i++ {
		t := f.In(i)
		v := reflect.New(t).Elem()

		var value reflect.Value
		var err error

		// todo: add feature named parameter
		k := fmt.Sprintf("%v", i)
		if arg, ok := args[k]; ok {
			o := reflect.TypeOf((*Object)(nil)).Elem()
			if t.Implements(o) {
				if p, ok := arg.(*PathValue); ok {
					value = reflect.ValueOf(p.value)
				} else {
					value = reflect.ValueOf(arg)
				}
			} else {
				value, err = ConvertFrom(arg, t)
			}
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
