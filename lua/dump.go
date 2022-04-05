package lua

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"math"
	"mokapi/lua/utils"
	"mokapi/sortedmap"
	"reflect"
	"strings"
)

func Dump(i interface{}) string {
	if lv, ok := i.(lua.LValue); ok {
		i = utils.FromValue(lv, nil)
	}

	if m, ok := i.(*sortedmap.LinkedHashMap); ok {
		return m.String()
	}

	var sb strings.Builder
	v := reflect.ValueOf(i)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	k := v.Kind()
	_ = k
	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		sb.WriteString(fmt.Sprintf("%v{", t.Name()))
		for i := 0; i < v.NumField(); i++ {
			if i > 0 {
				sb.WriteString(", ")
			}
			f := v.Field(i)
			if !f.CanInterface() {
				continue
			}
			name := t.Field(i).Name
			sb.WriteString(fmt.Sprintf("%v: %v", name, Dump(f.Interface())))
		}
		sb.WriteString("}")
	case reflect.Map:
		sb.WriteString("{")
		for i, k := range v.MapKeys() {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%v: %v", k, Dump(v.MapIndex(k).Interface())))
		}
		sb.WriteString("}")
	case reflect.Float64, reflect.Float32:
		f := v.Float()
		if i := math.Trunc(f); i == f {
			sb.WriteString(fmt.Sprintf("%v", int64(i)))
		} else {
			sb.WriteString(fmt.Sprintf("%f", f))
		}
	default:
		sb.WriteString(fmt.Sprintf("%v", i))
	}

	return sb.String()
}
