package js

import (
	"fmt"
	"github.com/dop251/goja"
	"mokapi/js/common"
	"mokapi/js/modules"
)

type factory func(common.Host, *goja.Runtime) interface{}

var moduleTypes = map[string]factory{
	"mokapi": modules.NewMokapi,
}

type require struct {
	modules map[string]interface{}
	runtime *goja.Runtime
	host    common.Host
}

func enableRequire(runtime *goja.Runtime, host common.Host) {
	r := &require{
		runtime: runtime,
		host:    host,
	}
	runtime.Set("require", r.require)
}

func (r *require) require(call goja.FunctionCall) goja.Value {
	file := call.Argument(0).String()
	if len(file) == 0 {
		panic(r.runtime.ToValue("missing argument"))
	}
	if f, ok := moduleTypes[file]; ok {
		m := f(r.host, r.runtime)
		return common.Map(r.runtime, m)
	} else {
		panic(r.runtime.ToValue(fmt.Sprintf("unknown module %v", file)))
	}
}
