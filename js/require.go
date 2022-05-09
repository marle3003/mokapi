package js

import (
	"github.com/dop251/goja"
	engine "mokapi/engine/common"
	"mokapi/js/common"
	"mokapi/js/compiler"
	"mokapi/js/modules"
)

type factory func(engine.Host, *goja.Runtime) interface{}

var moduleTypes = map[string]factory{
	"mokapi":    modules.NewMokapi,
	"generator": modules.NewGenerator,
}

type require struct {
	exports  map[string]goja.Value
	runtime  *goja.Runtime
	host     engine.Host
	compiler *compiler.Compiler
}

func enableRequire(runtime *goja.Runtime, host engine.Host) {
	r := &require{
		runtime: runtime,
		host:    host,
		exports: make(map[string]goja.Value),
	}
	var err error
	if r.compiler, err = compiler.New(); err != nil {
		panic(err)
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
		src, err := r.host.OpenFile(file)
		if err != nil {
			src, err = r.host.OpenFile(file + ".js")
			if err != nil {
				panic(err)
			}
		}

		prg, err := r.compiler.Compile(file, src)
		if err != nil {
			panic(err)
		}

		export := r.runtime.NewObject()
		_, err = r.runtime.RunProgram(prg)
		if err != nil {
			panic(err)
		}

		r.exports[file] = export
		return export

		//panic(r.runtime.ToValue(fmt.Sprintf("unknown module %v", file)))
	}
}
