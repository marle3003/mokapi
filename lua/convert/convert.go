package convert

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"mokapi/sortedmap"
	"reflect"
)

func FromLua(lv lua.LValue, to interface{}) error {
	switch lt := lv.(type) {
	case *lua.LNilType:
		v := reflect.ValueOf(to).Elem()
		v.Set(reflect.Zero(v.Type()))
		return nil
	case lua.LBool:
		return toBool(lt, to)
	case lua.LString:
		return toString(lt, to)
	case lua.LNumber:
		return toNumber(lt, to)
	case *lua.LTable:
		return fromTable(lt, to)
	case *lua.LUserData:
		return fromUserData(lt, to)
	default:
		return fmt.Errorf("type %v not supported", lt)
	}
}

func ToLua(l *lua.LState, from interface{}) (lua.LValue, error) {
	switch i := from.(type) {
	case bool:
		return lua.LBool(i), nil
	case string:
		return lua.LString(i), nil
	case int:
		return lua.LNumber(i), nil
	case int8:
		return lua.LNumber(i), nil
	case int16:
		return lua.LNumber(i), nil
	case int32:
		return lua.LNumber(i), nil
	case int64:
		return lua.LNumber(i), nil
	case float32:
		return lua.LNumber(i), nil
	case float64:
		return lua.LNumber(i), nil
	default:
		v := reflect.ValueOf(from)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		switch v.Kind() {
		case reflect.Invalid:
			return lua.LNil, nil
		case reflect.Slice:
			if v.IsNil() || v.Len() == 0 {
				return lua.LNil, nil
			}
			tbl := l.NewTable()
			for i := 0; i < v.Len(); i++ {
				item, err := ToLua(l, v.Index(i).Interface())
				if err != nil {
					return nil, err
				}
				tbl.Append(item)
			}
			return tbl, nil
		case reflect.Map:
			if v.IsNil() {
				return lua.LNil, nil
			}
			tbl := l.NewTable()
			for _, k := range v.MapKeys() {
				val, err := ToLua(l, v.MapIndex(k).Interface())
				if err != nil {
					return nil, err
				}
				l.SetField(tbl, fmt.Sprintf("%v", k), val)
			}
			return tbl, nil
		case reflect.Struct:
			if sm, ok := i.(sortedmap.LinkedHashMap); ok {
				tbl := l.NewTable()
				for it := sm.Iter(); it.Next(); {
					val, err := ToLua(l, it.Value())
					if err != nil {
						return nil, err
					}
					l.SetField(tbl, it.Key().(string), val)
				}
				return tbl, nil
			}

			fields := make([]reflect.StructField, 0, v.Type().NumField())
			for i := 0; i < v.Type().NumField(); i++ {
				f := v.Type().Field(i)
				if !f.IsExported() {
					continue
				}
				fields = append(fields, reflect.StructField{
					Name: f.Name,
					Type: reflect.TypeOf((*lua.LValue)(nil)).Elem(),
				})
			}

			p := reflect.New(reflect.StructOf(fields))
			to := p.Elem()
			for _, fd := range fields {
				fv := v.FieldByName(fd.Name).Interface()
				lv, err := ToLua(l, fv)
				if err != nil {
					return nil, err
				}
				to.FieldByName(fd.Name).Set(reflect.ValueOf(lv))
			}

			ud := luar.New(l, i).(*lua.LUserData)
			ud.Value = p.Interface()
			return ud, nil
		}
		return nil, fmt.Errorf("type %t not supported", i)
	}
}
