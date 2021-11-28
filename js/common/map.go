package common

import (
	"github.com/dop251/goja"
	"reflect"
	"strings"
)

func Map(rt *goja.Runtime, module interface{}) goja.Value {
	exports := make(map[string]interface{})

	val := reflect.ValueOf(module)
	typ := val.Type()
	for i := 0; i < typ.NumMethod(); i++ {
		meth := typ.Method(i)

		name := strings.ToLower(meth.Name[0:1]) + meth.Name[1:]
		exports[name] = val.Method(i).Interface()
	}

	return rt.ToValue(exports)
}
