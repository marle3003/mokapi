package lua

import (
	"fmt"
	"math"
	"reflect"
	"strings"
)

func Dump(i interface{}) string {
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
			sb.WriteString(fmt.Sprintf("%v: %v", t.Field(i).Name, Dump(f.Interface())))
		}
		sb.WriteString("}")
	case reflect.Map:
		for _, k := range v.MapKeys() {
			sb.WriteString(fmt.Sprintf("%v: %v", k, Dump(v.MapIndex(k).Interface())))
		}
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
