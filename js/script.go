package js

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/pkg/errors"
	engine "mokapi/engine/common"
	"mokapi/js/compiler"
)

var NoDefaultFunction = errors.New("js: no default function found")

type Script struct {
	runtime  *goja.Runtime
	exports  goja.Value
	compiler *compiler.Compiler
	host     engine.Host
	require  *require
	filename string
	source   string
}

func New(filename, src string, host engine.Host) (*Script, error) {
	s := &Script{
		host:     host,
		filename: filename,
		source:   src,
	}

	var err error
	if s.compiler, err = compiler.New(); err != nil {
		return nil, err
	}

	return s, err
}

func (s *Script) Run() error {
	_, err := s.RunDefault()
	if err == NoDefaultFunction {
		s.runtime = nil
		return nil
	}
	return err
}

func (s *Script) RunDefault() (goja.Value, error) {
	if err := s.ensureRuntime(); err != nil {
		return nil, err
	}
	o := s.exports.ToObject(s.runtime)
	if f, ok := goja.AssertFunction(o.Get("default")); ok {
		i, err := f(goja.Undefined())
		if err != nil {
			return nil, err
		}
		s.processObject(i)
		return i, nil
	} else {
		data := o.Get("mokapi")
		if data != nil && !goja.IsUndefined(data) && !goja.IsNull(data) {
			s.processObject(data)
			return data, nil
		}
	}
	return nil, NoDefaultFunction
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

func (s *Script) requireFile(filename, src string) (goja.Value, error) {
	module := s.runtime.NewObject()
	exports := s.runtime.NewObject()
	module.Set("exports", exports)

	prg, err := s.compiler.CompileModule(filename, src)
	if err != nil {
		return nil, err
	}
	f, err := s.runtime.RunProgram(prg)
	if err != nil {
		return nil, err
	}
	if call, ok := goja.AssertFunction(f); ok {
		_, err = call(exports, exports, module)
		if err != nil {
			return nil, err
		}
	}
	return exports, nil
}

func (s *Script) Close() {
	if s.runtime != nil {
		s.runtime.Interrupt(fmt.Errorf("closing"))
		s.runtime = nil
	}
	s.compiler = nil
	s.exports = nil
	if s.require != nil {
		s.require.close()
	}
}

func (s *Script) ensureRuntime() (err error) {
	if s.runtime != nil {
		return nil
	}
	s.runtime = goja.New()

	s.runtime.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	s.require = enableRequire(s, s.host)
	enableConsole(s.runtime, s.host)
	enableOpen(s.runtime, s.host)

	s.exports, err = s.openScript(s.filename, s.source)
	return
}

func (s *Script) processObject(v goja.Value) {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return
	}
	m, ok := v.Export().(map[string]interface{})
	if !ok {
		return
	}
	if http, ok := m["http"]; ok {
		s.addHttpEvent(http)
	}
}

func (s *Script) addHttpEvent(i interface{}) {
	f := func(args ...interface{}) (bool, error) {
		if len(args) != 2 {
			return false, fmt.Errorf("expected args: request, response")
		}
		req := args[0].(*engine.EventRequest)
		res := args[1].(*engine.EventResponse)
		return engine.EventHandler(req, res, i)
	}

	s.host.On("http", f, nil)
}
