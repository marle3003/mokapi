package convert

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"mokapi/sortedmap"
	"reflect"
	"strings"
)

func fromTable(tbl *lua.LTable, to interface{}) error {
	v := reflect.ValueOf(to).Elem()
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		return toArray(tbl, to)
	case reflect.Map:
		return toMap(tbl, to)
	case reflect.Ptr, reflect.Struct:
		if v.Type() == reflect.TypeOf((*sortedmap.LinkedHashMap[string, interface{}])(nil)) {
			return toSortedMapFromTable(tbl, to)
		}
		return toStructFromTable(tbl, to)
	case reflect.Interface:
		if tbl.MaxN() == 0 {
			return toSortedMapFromTable(tbl, to)
		}
		return toInterfaceArray(tbl, to)
	default:
		return fmt.Errorf("unable to convert table to %v", v.Kind())
	}
}

func toArray(tbl *lua.LTable, to interface{}) error {
	v := reflect.ValueOf(to).Elem()
	for i := 1; i <= tbl.MaxN(); i++ {
		item := reflect.New(v.Type().Elem()).Interface()
		err := FromLua(tbl.RawGetInt(i), item)
		if err != nil {
			return err
		}
		vi := reflect.ValueOf(item)
		if v.Type().Elem().Kind() != reflect.Ptr {
			vi = vi.Elem()
		}
		v = reflect.Append(v, vi)
	}
	reflect.ValueOf(to).Elem().Set(v)
	return nil
}

func toInterfaceArray(tbl *lua.LTable, to interface{}) error {
	a := make([]interface{}, 0, tbl.Len())
	for i := 1; i <= tbl.MaxN(); i++ {
		var value interface{}
		err := FromLua(tbl.RawGetInt(i), &value)
		if err != nil {
			return err
		}
		a = append(a, value)
	}
	reflect.ValueOf(to).Elem().Set(reflect.ValueOf(a))
	return nil
}

func toMap(tbl *lua.LTable, to interface{}) error {
	v := reflect.ValueOf(to).Elem()
	t := v.Type()
	if v.IsZero() {
		v = reflect.MakeMap(t)
		reflect.ValueOf(to).Elem().Set(v)
	}
	var err interface{}

	func() {
		defer func() {
			err = recover()
		}()

		tbl.ForEach(func(lk lua.LValue, lv lua.LValue) {
			key := reflect.New(v.Type().Key()).Interface()
			err := FromLua(lk, key)
			if err != nil {
				panic(err)
			}
			value := reflect.New(v.Type().Elem()).Interface()
			err = FromLua(lv, value)
			if err != nil {
				panic(err)
			}

			vk := reflect.ValueOf(key)
			if t.Key().Kind() != reflect.Ptr {
				vk = vk.Elem()
			}
			vv := reflect.ValueOf(value)
			if t.Elem().Kind() != reflect.Ptr {
				vv = vv.Elem()
			}

			v.SetMapIndex(vk, vv)
		})
	}()

	if err != nil {
		return fmt.Errorf("unable to convert to map: %v", err)
	}
	return nil
}

func toSortedMapFromTable(tbl *lua.LTable, to interface{}) error {
	vm := reflect.ValueOf(to).Elem()
	m, _ := vm.Interface().(*sortedmap.LinkedHashMap[string, interface{}])
	if m == nil {
		m = sortedmap.NewLinkedHashMap()
		vm.Set(reflect.ValueOf(m))
	}

	// using v.Next instead of v.ForEach to ensure the order of the table items
	// v.ForEach loops over a map, v.Next loops over the keys which is an array
	k := lua.LNil
	var v lua.LValue
	for {
		k, v = tbl.Next(k)
		if k == lua.LNil {
			break
		}

		var key interface{}
		err := FromLua(k, &key)
		if err != nil {
			return err
		}

		var value interface{}
		err = FromLua(v, &value)
		if err != nil {
			return err
		}

		m.Set(key.(string), value)
	}
	return nil
}

func toStructFromTable(tbl *lua.LTable, to interface{}) error {
	v := reflect.ValueOf(to).Elem()
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v = reflect.New(v.Type().Elem())
			reflect.ValueOf(to).Elem().Set(v)
		}
		v = v.Elem()
	}
	t := v.Type()
	k := lua.LNil
	var val lua.LValue
	for {
		k, val = tbl.Next(k)
		if k == lua.LNil {
			break
		}

		fieldName := k.String()
		field, found := t.FieldByName(strings.Title(fieldName))
		if !found {
			continue
		}

		value := create(field.Type)

		err := FromLua(val, value)
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

func create(t reflect.Type) interface{} {
	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem()).Interface()
	} else {
		return reflect.New(t).Interface()
	}
}
