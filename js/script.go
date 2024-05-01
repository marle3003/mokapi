package js

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/pkg/errors"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	engine "mokapi/engine/common"
	"mokapi/js/compiler"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
)

var NoDefaultFunction = errors.New("js: no default function found")

type Script struct {
	runtime  *goja.Runtime
	compiler *compiler.Compiler
	host     engine.Host
	require  *requireModule
	file     *dynamic.Config
	config   static.JsConfig
	runner   *runner
}

func New(file *dynamic.Config, host engine.Host, config static.JsConfig) (*Script, error) {
	s := &Script{
		host:   host,
		file:   file,
		config: config,
	}

	var err error
	if s.compiler, err = compiler.New(); err != nil {
		return nil, err
	}

	return s, err
}

func (s *Script) Run() error {
	if err := s.ensureRuntime(); err != nil {
		return err
	}

	_, err := s.RunDefault()
	if err != nil {
		if errors.Is(err, NoDefaultFunction) {
			return nil
		}
		return err
	}
	s.runner.StartLoop()

	return nil
}

func (s *Script) RunDefault() (goja.Value, error) {
	if err := s.ensureRuntime(); err != nil {
		return nil, err
	}

	var result goja.Value
	var err error
	s.runner.Run(func(vm *goja.Runtime) {
		v := vm.Get("exports")
		if v == goja.Null() {
			return
		}
		exports := v.ToObject(vm)
		if f, ok := goja.AssertFunction(exports.Get("default")); ok {
			result, err = f(goja.Undefined())
		} else {
			data := exports.Get("mokapi")
			if data != nil && !goja.IsUndefined(data) && !goja.IsNull(data) {
				s.processObject(data)
			} else {
				err = NoDefaultFunction
			}
		}
	})

	if err != nil {
		return nil, err
	}

	s.processObject(result)
	return result, nil
}

func (s *Script) RunFunc(fn func(vm *goja.Runtime)) error {
	if err := s.ensureRuntime(); err != nil {
		return err
	}

	s.runner.Run(fn)
	return nil
}

func (s *Script) Close() {
	s.runner.Stop()
	if s.runtime != nil {
		s.runtime.Interrupt(fmt.Errorf("closing"))
		s.runtime = nil
	}
	s.compiler = nil
	if s.require != nil {
		s.require.Close()
	}
}

func (s *Script) CanClose() bool {
	return !s.runner.HasJobs()
}

func (s *Script) ensureRuntime() error {
	if s.runtime != nil {
		return nil
	}
	s.runtime = goja.New()
	s.runner = newRunner(s.runtime)
	path := getScriptPath(s.file.Info.Kernel().Url)
	workingDir := filepath.Dir(path)

	s.runtime.SetFieldNameMapper(&customFieldNameMapper{})
	s.require = newRequire(
		s.host.OpenFile,
		s.compiler,
		s.file.Info.Url.String(),
		workingDir,
		map[string]ModuleLoader{
			"mokapi":          s.loadNativeModule(newMokapi),
			"mokapi/faker":    s.loadNativeModule(newFaker),
			"faker":           s.loadDeprecatedNativeModule(newFaker, "deprecated module faker: Please use mokapi/faker instead"),
			"mokapi/http":     s.loadNativeModule(newHttp),
			"http":            s.loadDeprecatedNativeModule(newHttp, "deprecated module http: Please use mokapi/http instead"),
			"mokapi/kafka":    s.loadNativeModule(newKafka),
			"kafka":           s.loadDeprecatedNativeModule(newKafka, "deprecated module kafka: Please use mokapi/kafka instead"),
			"mokapi/mustache": s.loadNativeModule(newMustache),
			"mustache":        s.loadDeprecatedNativeModule(newMustache, "deprecated module mustache: Please use mokapi/mustache instead"),
			"mokapi/yaml":     s.loadNativeModule(newYaml),
			"yaml":            s.loadDeprecatedNativeModule(newYaml, "deprecated module yaml: Please use mokapi/yaml instead"),
			"mokapi/mail": s.loadNativeModule(func(host engine.Host, runtime *goja.Runtime) interface{} {
				return newMail(host, runtime, filepath.Dir(workingDir))
			}),
			"mokapi/ldap": func() goja.Value {
				return NewLdapModule(s.runtime)
			},
		})
	s.require.Enable(s.runtime)
	enableConsole(s.runtime, s.host)
	enableOpen(s.runtime, s.host)
	enableProcess(s.runtime)

	prg, err := s.compiler.Compile(path, string(s.file.Raw))
	if err != nil {
		return err
	}

	s.runner.Run(func(vm *goja.Runtime) {
		_, err = vm.RunProgram(prg)
	})
	return err
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

func (s *Script) loadNativeModule(f func(engine.Host, *goja.Runtime) interface{}) ModuleLoader {
	return func() goja.Value {
		m := f(s.host, s.runtime)
		return mapToJSValue(s.runtime, m)
	}
}

func (s *Script) loadDeprecatedNativeModule(f func(engine.Host, *goja.Runtime) interface{}, msg string) ModuleLoader {
	filename := getScriptPath(s.file.Info.Url)
	return func() goja.Value {
		s.host.Warn(fmt.Sprintf("%v: %v", msg, filename))
		m := f(s.host, s.runtime)
		return mapToJSValue(s.runtime, m)
	}
}

// customFieldNameMapper default implementation filters out
// "invalid" identifiers but also prevents accessing by
// index operator such as object['prop']
type customFieldNameMapper struct {
}

func (cfm customFieldNameMapper) FieldName(_ reflect.Type, f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if len(tag) == 0 {
		return uncapitalize(f.Name)
	}
	if idx := strings.IndexByte(tag, ','); idx != -1 {
		tag = tag[:idx]
	}

	return tag
}

func (cfm customFieldNameMapper) MethodName(_ reflect.Type, m reflect.Method) string {
	return uncapitalize(m.Name)
}

func uncapitalize(s string) string {
	return strings.ToLower(s[0:1]) + s[1:]
}

func getScriptPath(u *url.URL) string {
	if len(u.Path) > 0 {
		return u.Path
	}
	return u.Opaque
}
