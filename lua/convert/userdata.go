package convert

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"mokapi/sortedmap"
	"reflect"
	"strings"
)

func fromUserData(userdata *lua.LUserData, to interface{}) error {
	v := reflect.ValueOf(to).Elem()
	switch v.Kind() {
	case reflect.Ptr, reflect.Struct:
		if v.Type() == reflect.TypeOf((*sortedmap.LinkedHashMap)(nil)) {
			return toSortedMapFromUserData(userdata, to)
		}
		return toStructFromUserData(userdata, to)
	case reflect.Interface:
		return toSortedMapFromUserData(userdata, to)
	default:
		return fmt.Errorf("unable to convert userdata to %v", v.Kind())
	}
}

func toSortedMapFromUserData(userdata *lua.LUserData, to interface{}) error {
	vm := reflect.ValueOf(to).Elem()
	m, _ := vm.Interface().(*sortedmap.LinkedHashMap)
	if m == nil {
		m = sortedmap.NewLinkedHashMap()
		vm.Set(reflect.ValueOf(m))
	}

	from := reflect.ValueOf(userdata.Value)
	if from.Kind() == reflect.Ptr {
		from = from.Elem()
	}
	t := from.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := from.FieldByName(f.Name).Interface()

		var value interface{}
		err := FromLua(fv.(lua.LValue), &value)
		if err != nil {
			return err
		}

		m.Set(f.Name, value)
	}

	return nil
}

func toStructFromUserData(userdata *lua.LUserData, to interface{}) error {
	v := reflect.ValueOf(to).Elem()
	if v.Kind() == reflect.Ptr {
		if v.IsValid() {
			v = reflect.New(v.Type().Elem())
			reflect.ValueOf(to).Elem().Set(v)
		}
		v = v.Elem()
	}

	from := reflect.ValueOf(userdata.Value)
	if from.Kind() == reflect.Ptr {
		from = from.Elem()
	}
	t := from.Type()
	for i := 0; i < t.NumField(); i++ {
		val := from.Field(i).Interface()

		field, found := v.Type().FieldByName(strings.Title(t.Field(i).Name))
		if !found {
			continue
		}

		value := create(field.Type)

		err := FromLua(val.(lua.LValue), value)
		if err != nil {
			return err
		}

		vv := reflect.ValueOf(value)
		if field.Type.Kind() != reflect.Ptr {
			vv = vv.Elem()
		}

		v.FieldByIndex(field.Index).Set(vv)
	}

	return nil
}
