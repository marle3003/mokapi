package js

import (
	"fmt"
	"github.com/dop251/goja"
	"mokapi/js/common"
	"mokapi/js/compiler"
)

type Script struct {
	runtime  *goja.Runtime
	prg      *goja.Program
	exports  map[string]goja.Callable
	compiler *compiler.Compiler
	host     common.Host
}

func New(filename, src string, host common.Host) (*Script, error) {
	s := &Script{
		runtime: goja.New(),
		exports: make(map[string]goja.Callable),
		host:    host,
	}

	var err error
	if s.compiler, err = compiler.New(); err != nil {
		return nil, err
	}

	if s.prg, err = s.compiler.Compile(filename, src); err != nil {
		return nil, err
	}

	s.runtime.SetFieldNameMapper(goja.TagFieldNameMapper("js", true))

	exports := s.runtime.NewObject()
	s.runtime.Set("exports", exports)

	enableRequire(s.runtime, host)
	enableConsole(s.runtime, host)

	_, err = s.runtime.RunProgram(s.prg)
	if err != nil {
		return nil, err
	}

	err = s.getExports()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Script) Run() error {
	_, err := s.exports["default"](goja.Undefined())
	return err
}

func (s *Script) Close() {
	s.runtime.Interrupt(fmt.Errorf("closing"))
}

func (s *Script) getExports() error {
	v := s.runtime.Get("exports")
	if v == nil || goja.IsNull(v) || goja.IsUndefined(v) {
		return fmt.Errorf("export must be an object")
	}

	o := v.ToObject(s.runtime)
	for _, k := range o.Keys() {
		v := o.Get(k)
		if f, ok := goja.AssertFunction(v); ok {
			s.exports[k] = f
		}
	}

	if len(s.exports) == 0 {
		return fmt.Errorf("no exported functions in script")
	}

	return nil
}
