package utils

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"math"
)

func MapTable(tbl *lua.LTable) interface{} {
	if tbl == nil {
		return nil
	}
	return toValue(tbl)
}

func toValue(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
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
			ret := make(map[string]interface{})
			v.ForEach(func(key, value lua.LValue) {
				k := fmt.Sprintf("%v", toValue(key))
				ret[k] = toValue(value)
			})
			return ret
		} else { // array
			ret := make([]interface{}, 0, n)
			for i := 1; i <= n; i++ {
				ret = append(ret, toValue(v.RawGetInt(i)))
			}
			return ret
		}
	default:
		return v
	}
}
