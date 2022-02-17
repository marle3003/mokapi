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
	return FromValue(tbl)
}

func FromValue(lv lua.LValue) interface{} {
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
		n := v.MaxN()
		if n == 0 { // table
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

				key := fmt.Sprintf("%v", FromValue(k))
				ret.Set(key, FromValue(val))
			}
			return ret
		} else { // array
			ret := make([]interface{}, 0, n)
			for i := 1; i <= n; i++ {
				ret = append(ret, FromValue(v.RawGetInt(i)))
			}
			return ret
		}
	case *lua.LUserData:
		from := reflect.ValueOf(v.Value)
		t := from.Elem().Type()
		fields := make([]reflect.StructField, 0, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fields = append(fields, reflect.StructField{
				Name: f.Name,
				Type: reflect.TypeOf((*interface{})(nil)).Elem(),
			})
		}

		p := reflect.New(reflect.StructOf(fields))
		e := p.Elem()
		for _, fd := range fields {
			f := e.FieldByName(fd.Name)
			fv := from.Elem().FieldByName(fd.Name).Interface()
			val := FromValue(fv.(lua.LValue))
			if val != nil {
				f.Set(reflect.ValueOf(val))
			} else {
				f.Set(reflect.Zero(fd.Type))
			}
		}

		return p.Interface()
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
		if val.IsNil() {
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

	switch converted := v.(type) {
	case *lua.LUserData:
		srcVal := reflect.ValueOf(converted.Value).Elem()
		t := srcVal.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fv := srcVal.Field(i).Interface()
			if fv != nil {
				dstF := dstVal.FieldByName(f.Name)
				switch dstF.Kind() {
				case reflect.Ptr:
					if err := Map(dstF.Interface(), fv.(lua.LValue)); err != nil {
						return err
					}
				case reflect.Struct:
					if err := Map(dstF.Addr().Interface(), fv.(lua.LValue)); err != nil {
						return err
					}
				default:
					v := FromValue(fv.(lua.LValue))
					if v != nil {
						dstF.Set(convert(reflect.ValueOf(v), dstF.Type()))
					} else {
						dstF.Set(reflect.Zero(dstF.Type()))
					}
				}
			}
		}
	case *lua.LTable:
		n := converted.MaxN()
		if n > 0 {
			return fmt.Errorf("map does not support array")
		}

		converted.ForEach(func(key, value lua.LValue) {
			k := fmt.Sprintf("%v", FromValue(key))
			val := reflect.ValueOf(FromValue(value))
			f := dstVal.FieldByName(strings.Title(k))
			val = convert(val, f.Type())

			dstVal.FieldByName(strings.Title(k)).Set(val)
		})
	default:
		return fmt.Errorf("map does not support %t", v)
	}

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
