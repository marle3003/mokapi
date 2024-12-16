package encoding

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
)

func Require(vm *goja.Runtime, module *goja.Object) {
	o := vm.Get("mokapi/internal").(*goja.Object)
	host := o.Get("host").Export().(common.Host)
	obj := module.Get("exports").(*goja.Object)

	b64 := &base64{
		rt:   vm,
		host: host,
	}
	b := vm.NewObject()
	b.Set("encode", b64.Encode)
	b.Set("decode", b64.Decode)
	obj.Set("base64", b)
}
