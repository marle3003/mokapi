package util

import (
	"mokapi/config/dynamic"
	"mokapi/lib"

	"github.com/dop251/goja"
)

func JsType(v interface{}) string {
	return lib.TypeFrom(v)
}

func GetScriptFile(vm *goja.Runtime) *dynamic.Config {
	return vm.Get("mokapi/internal").(*goja.Object).Get("file").Export().(*dynamic.Config)
}
