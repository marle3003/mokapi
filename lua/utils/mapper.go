package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"math"
	"mokapi/sortedmap"
	"reflect"
	"strings"
)

func MapTable(tbl *lua.LTable) interface{} {
	if tbl == nil {
		return nil
	}
	return FromValue(tbl, nil)
}

func FromValue(lv lua.LValue, hint reflect.Type) interface{} {
	if hint == nil {
		hint = reflect.TypeOf((*interface{})(nil)).Elem()
	}

	isPtr := false

	switch v := lv.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return v == lua.LTrue
	case lua.LString:
		return string(v)
	case lua.LNumber:
		f := float64(v)
		if i := math.Trunc(f); i == f {
			return int64(i)
		}
		return f
	case *lua.LTable:
		switch {
		case hint.Kind() == reflect.Array:
			ret := reflect.New(hint).Elem()
			for i := 1; i <= v.MaxN(); i++ {
				item := FromValue(v.RawGetInt(i), hint.Elem())
				reflect.Append(ret, reflect.ValueOf(item))
			}
			return ret
		case hint.Kind() == reflect.Slice:
			length := v.Len()
			ret := reflect.MakeSlice(hint, 0, length)
			for i := 1; i <= v.MaxN(); i++ {
				item := FromValue(v.RawGetInt(i), hint.Elem())
				ret = reflect.Append(ret, reflect.ValueOf(item))
			}
			return ret.Interface()
		case hint.Kind() == reflect.Ptr:
			hint = hint.Elem()
			isPtr = true
			fallthrough
		case hint.Kind() == reflect.Struct:
			ret := reflect.New(hint)
			t := ret.Elem()
			k := lua.LNil
			var val lua.LValue
			for {
				k, val = v.Next(k)
				if k == lua.LNil {
					break
				}

				fieldName := k.String()
				field, found := hint.FieldByName(strings.Title(fieldName))
				if !found {
					continue
				}
				value := FromValue(val, field.Type)
				converted := convert(reflect.ValueOf(value), field.Type)
				t.FieldByIndex(field.Index).Set(converted)
			}
			if isPtr {
				return ret.Interface()
			}
			return t.Interface()
		case hint.Kind() == reflect.Map:
			ret := reflect.MakeMap(hint)

			v.ForEach(func(vKey lua.LValue, vVal lua.LValue) {
				key := FromValue(vKey, hint.Key())
				val := FromValue(vVal, hint.Elem())
				ret.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(val))
			})
			return ret.Interface()
		default:
			if v.MaxN() == 0 {
				ret := sortedmap.NewLinkedHashMap()

				// using v.Next instead of v.ForEach to ensure the order of the table items
				// v.ForEach loops over a map, v.Next loops over the keys which is an array
				k := lua.LNil
				var val lua.LValue
				for {
					k, val = v.Next(k)
					if k == lua.LNil {
						break
					}

					key := fmt.Sprintf("%v", FromValue(k, reflect.TypeOf("")))
					ret.Set(key, FromValue(val, hint))
				}
				return ret
			}
			length := v.Len()
			ret := make([]interface{}, 0, length)
			for i := 1; i <= length; i++ {
				ret = append(ret, FromValue(v.RawGetInt(i), hint))
			}
			return ret
		}
	case *lua.LUserData:
		switch {
		case hint.Kind() == reflect.Ptr:
			hint = hint.Elem()
			isPtr = true
			fallthrough
		case hint.Kind() == reflect.Struct:
			dest := reflect.New(hint)
			from := reflect.ValueOf(v.Value)
			ft := from.Elem().Type()
			dt := dest.Elem()
			for i := 0; i < ft.NumField(); i++ {
				f := ft.Field(i)
				dstF := dt.FieldByName(f.Name)
				fv := from.Elem().FieldByName(f.Name).Interface()
				val := FromValue(fv.(lua.LValue), dstF.Type())
				if val != nil {
					con := convert(reflect.ValueOf(val), dstF.Type())
					dstF.Set(con)
				} else {
					dstF.Set(reflect.Zero(f.Type))
				}
			}
			if isPtr {
				return dest.Interface()
			}
			return dt.Interface()
		case hint.Kind() == reflect.Interface:
			m := sortedmap.NewLinkedHashMap()
			from := reflect.ValueOf(v.Value)
			t := from.Elem().Type()
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				fv := from.Elem().FieldByName(f.Name).Interface()
				val := FromValue(fv.(lua.LValue), nil)
				if val != nil {
					m.Set(f.Name, val)
				} else {
					m.Set(f.Name, nil)
				}
			}
			return m
		}
		return v
	default:
		return v
	}
}

func ToValue(l *lua.LState, i interface{}) lua.LValue {
	switch val := reflect.ValueOf(i); val.Kind() {
	case reflect.Bool:
		return lua.LBool(val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(float64(val.Uint()))
	case reflect.Float32, reflect.Float64:
		return lua.LNumber(val.Float())
	case reflect.Ptr:
		if sm, ok := i.(*sortedmap.LinkedHashMap); ok {
			tbl := l.NewTable()
			for it := sm.Iter(); it.Next(); {
				l.SetField(tbl, it.Key().(string), ToValue(l, it.Value()))
			}
			return tbl
		}

		from := reflect.ValueOf(i)
		t := from.Elem().Type()
		fields := make([]reflect.StructField, 0, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			fields = append(fields, reflect.StructField{
				Name: f.Name,
				Type: reflect.TypeOf((*lua.LValue)(nil)).Elem(),
			})
		}

		v := reflect.New(reflect.StructOf(fields))
		e := v.Elem()
		for _, fd := range fields {
			f := e.FieldByName(fd.Name)
			fv := from.Elem().FieldByName(fd.Name).Interface()
			f.Set(reflect.ValueOf(ToValue(l, fv)))
		}

		ud := luar.New(l, i).(*lua.LUserData)
		ud.Value = v.Interface()
		return ud
	case reflect.Map:
		if val.IsNil() {
			return lua.LNil
		}
		if val.IsNil() {
			return lua.LNil
		}
		tbl := l.NewTable()
		for _, e := range val.MapKeys() {
			l.SetField(tbl, fmt.Sprintf("%v", e), ToValue(l, val.MapIndex(e).Interface()))
		}
		return tbl
	case reflect.Slice:
		if val.IsNil() || val.Len() == 0 {
			return lua.LNil
		}
		tbl := l.NewTable()
		for i := 0; i < val.Len(); i++ {
			tbl.Append(ToValue(l, val.Index(i).Interface()))
		}
		return tbl
	case reflect.String:
		return lua.LString(val.String())
	case reflect.Struct:
		from := reflect.ValueOf(i)
		t := from.Type()
		fields := make([]reflect.StructField, 0, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fields = append(fields, reflect.StructField{
				Name: f.Name,
				Type: reflect.TypeOf((*lua.LValue)(nil)).Elem(),
			})
		}

		v := reflect.New(reflect.StructOf(fields))
		e := v.Elem()
		for _, fd := range fields {
			f := e.FieldByName(fd.Name)
			fv := from.FieldByName(fd.Name)
			f.Set(reflect.ValueOf(ToValue(l, fv.Interface())))
		}

		ud := luar.New(l, i).(*lua.LUserData)
		ud.Value = v.Interface()
		return ud
	case reflect.Invalid:
		return lua.LNil
	default:
		panic(fmt.Sprintf("not supported kind %v", val.Kind()))
	}
}

func Map(i interface{}, v lua.LValue) error {
	dstVal := reflect.ValueOf(i)

	if dstVal.Kind() == reflect.Ptr {
		dstVal = dstVal.Elem()
	}

	r := FromValue(v, reflect.TypeOf(i))
	rv := reflect.ValueOf(r)
	rv = convert(rv, dstVal.Type())

	dstVal.Set(rv.Elem())

	return nil
}

func convert(val reflect.Value, t reflect.Type) reflect.Value {
	if val.Type().AssignableTo(t) {
		return val
	}

	if val.Type().ConvertibleTo(t) {
		return val.Convert(t)
	}

	switch t.Kind() {
	case reflect.Map:
		i := val.Interface()
		if sm, ok := i.(*sortedmap.LinkedHashMap); ok {
			m := reflect.MakeMap(t)
			tKey := t.Key()
			tVal := t.Elem()
			for it := sm.Iter(); it.Next(); {
				k := convert(reflect.ValueOf(it.Key()), tKey)
				v := convert(reflect.ValueOf(it.Value()), tVal)
				m.SetMapIndex(k, v)
			}
			return m
		}
		return val
	case reflect.Struct:
	}

	log.Debugf("unable to convert value to %v from package %v", t.Name(), t.PkgPath())

	return val
}
