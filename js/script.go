package js

import (
	"fmt"
	"github.com/dop251/goja"
	engine "mokapi/engine/common"
	"mokapi/js/compiler"
)

type Script struct {
	runtime  *goja.Runtime
	prg      *goja.Program
	exports  goja.Value
	compiler *compiler.Compiler
	host     engine.Host
}

func New(filename, src string, host engine.Host) (*Script, error) {
	s := &Script{
		runtime: goja.New(),
		host:    host,
	}

	var err error
	if s.compiler, err = compiler.New(); err != nil {
		return nil, err
	}

	if s.prg, err = s.compiler.Compile(filename, src); err != nil {
		return nil, err
	}

	s.runtime.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	s.exports = s.runtime.NewObject()
	s.runtime.Set("exports", s.exports)

	enableRequire(s.runtime, host)
	enableConsole(s.runtime, host)
	enableOpen(s.runtime, host)

	_, err = s.runtime.RunProgram(s.prg)
	if err != nil {
		return nil, err
	}

	return s, err
}

func (s *Script) Run() error {
	o := s.exports.ToObject(s.runtime)
	if f, ok := goja.AssertFunction(o.Get("default")); ok {
		_, err := f(goja.Undefined())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Script) Close() {
	s.runtime.Interrupt(fmt.Errorf("closing"))
}

func getExports(runtime *goja.Runtime, exports map[string]goja.Value) error {
	v := runtime.Get("exports")
	if v == nil || goja.IsNull(v) || goja.IsUndefined(v) {
		return fmt.Errorf("export must be an object")
	}

	o := v.ToObject(runtime)
	for _, k := range o.Keys() {
		v := o.Get(k)
		exports[k] = v
	}

	if len(exports) == 0 {
		return fmt.Errorf("no exported functions in script")
	}

	return nil
}
