package js

import (
	"fmt"
	"github.com/dop251/goja"
	engine "mokapi/engine/common"
	"mokapi/js/common"
	"mokapi/js/modules"
	"mokapi/js/modules/http"
	"mokapi/js/modules/kafka"
	"strings"
)

type factory func(engine.Host, *goja.Runtime) interface{}

var moduleTypes = map[string]factory{
	"mokapi":    modules.NewMokapi,
	"generator": modules.NewGenerator,
	"http":      http.New,
	"kafka":     kafka.New,
}

type require struct {
	exports map[string]goja.Value
	runtime *goja.Runtime
	host    engine.Host
}

func enableRequire(runtime *goja.Runtime, host engine.Host) {
	r := &require{
		runtime: runtime,
		host:    host,
		exports: make(map[string]goja.Value),
	}
	runtime.Set("require", r.require)
}

func (r *require) require(call goja.FunctionCall) goja.Value {
	file := call.Argument(0).String()
	if len(file) == 0 {
		panic(r.runtime.ToValue("missing argument"))
	}

	if e, ok := r.exports[file]; ok {
		return e
	}

	if f, ok := moduleTypes[file]; ok {
		m := f(r.host, r.runtime)
		e := common.Map(r.runtime, m)
		r.exports[file] = e
		return e
	} else {
		if !strings.HasSuffix(file, ".js") {
			file = file + ".js"
		}
		s, err := r.host.OpenScript(file)
		if err != nil {
			panic(err)
		}

		js, ok := s.(*Script)
		if !ok {
			panic(fmt.Sprintf("not supporting %v", file))
		}

		r.exports[file] = js.exports
		return js.exports
	}
}
