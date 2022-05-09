package js

import (
	"fmt"
	"github.com/dop251/goja"
	engine "mokapi/engine/common"
	"mokapi/js/common"
	"mokapi/js/modules"
)

type factory func(engine.Host, *goja.Runtime) interface{}

var moduleTypes = map[string]factory{
	"mokapi":    modules.NewMokapi,
	"generator": modules.NewGenerator,
}

type require struct {
	modules map[string]interface{}
	runtime *goja.Runtime
	host    engine.Host
}

func enableRequire(runtime *goja.Runtime, host engine.Host) {
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
