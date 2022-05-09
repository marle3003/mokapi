package convert

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"math"
	"reflect"
)

func toBool(lv lua.LValue, to interface{}) error {
	b := reflect.ValueOf(lv == lua.LTrue)
	reflect.ValueOf(to).Elem().Set(b)
	return nil
}

func toString(lv lua.LValue, to interface{}) error {
	s := reflect.ValueOf(lv.String())
	reflect.ValueOf(to).Elem().Set(s)
	return nil
}

func toNumber(lv lua.LNumber, to interface{}) error {
	v := reflect.ValueOf(to).Elem()
	f := float64(lv)
	switch to.(type) {
	case *int, *int8, *int16, *int32, *int64:
		i := math.Trunc(f)
		v.SetInt(int64(i))
	case *float32, *float64:
		v.SetFloat(f)
	case *interface{}:
		v.Set(reflect.ValueOf(f))
	default:
		return fmt.Errorf("unable to convert number to %s", v.Type().Name())
	}

	return nil
}
