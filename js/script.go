package js

import (
	"fmt"
	"github.com/dop251/goja"
	engine "mokapi/engine/common"
	"mokapi/js/compiler"
)

type Script struct {
	runtime  *goja.Runtime
	exports  goja.Value
	compiler *compiler.Compiler
	host     engine.Host
	require  *require
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

	s.runtime.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	s.require = enableRequire(s, host)
	enableConsole(s.runtime, host)
	enableOpen(s.runtime, host)

	s.exports, err = s.openScript(filename, src)
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

func (s *Script) openScript(filename, src string) (goja.Value, error) {
	exports := s.runtime.NewObject()
	s.runtime.Set("exports", exports)
	prg, err := s.compiler.Compile(filename, src)
	if err != nil {
		return nil, err
	}
	_, err = s.runtime.RunProgram(prg)
	if err != nil {
		return nil, err
	}
	return exports, nil
}

func (s *Script) Close() {
	s.runtime.Interrupt(fmt.Errorf("closing"))
	s.runtime = nil
	s.compiler = nil
	s.exports = nil
	s.require.close()
}
