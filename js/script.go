package js

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/pkg/errors"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	engine "mokapi/engine/common"
	"mokapi/js/compiler"
	"mokapi/js/console"
	"mokapi/js/eventloop"
	"mokapi/js/faker"
	"mokapi/js/file"
	"mokapi/js/http"
	"mokapi/js/kafka"
	"mokapi/js/ldap"
	"mokapi/js/mail"
	"mokapi/js/mokapi"
	"mokapi/js/mustache"
	"mokapi/js/process"
	"mokapi/js/require"
	"mokapi/js/yaml"
	"net/url"
	"reflect"
	"strings"
	"sync"
)

var (
	NoDefaultFunction = errors.New("js: no default function found")

	singletonRegistry *require.Registry
	registryOne       sync.Once
	registryErr       error
)

type Script struct {
	runtime  *goja.Runtime
	compiler *compiler.Compiler
	host     engine.Host
	file     *dynamic.Config
	config   static.JsConfig
	runner   *eventloop.EventLoop
	registry *require.Registry
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

func NewScript(opts ...Option) (*Script, error) {
	s := &Script{}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
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

	return nil
}

func (s *Script) RunDefault() (goja.Value, error) {
	if err := s.ensureRuntime(); err != nil {
		return nil, err
	}

	s.runner.StartLoop()

	result, err := s.runner.RunAsync(func(vm *goja.Runtime) (goja.Value, error) {
		v := vm.Get("exports")
		if v == goja.Null() {
			return nil, NoDefaultFunction
		}
		exports := v.ToObject(vm)
		if f, ok := goja.AssertFunction(exports.Get("default")); ok {
			return f(goja.Undefined())
		} else {
			data := exports.Get("mokapi")
			if data != nil && !goja.IsUndefined(data) && !goja.IsNull(data) {
				return data, nil
			}
		}
		return nil, NoDefaultFunction
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
}

func (s *Script) CanClose() bool {
	return !s.runner.HasJobs()
}

func (s *Script) ensureRuntime() error {
	if s.runtime != nil {
		return nil
	}
	s.runtime = goja.New()
	s.runner = eventloop.New(s.runtime)
	path := getScriptPath(s.file.Info.Kernel().Url)

	s.runtime.SetFieldNameMapper(&customFieldNameMapper{})
	registry, err := s.getRegistry()
	if err != nil {
		return err
	}
	registry.Enable(s.runtime)
	s.enableInternal()
	console.Enable(s.runtime)
	file.Enable(s.runtime, s.host)
	process.Enable(s.runtime)

	prg, err := registry.GetProgram(path, string(s.file.Raw))
	if err != nil {
		return err
	}

	s.runner.Run(func(vm *goja.Runtime) {
		_, err = vm.RunProgram(prg)
	})
	return err
}

func (s *Script) enableInternal() {
	o := s.runtime.NewObject()
	o.Set("host", s.host)
	o.Set("loop", s.runner)
	s.runtime.Set("mokapi/internal", o)

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

func (s *Script) getRegistry() (*require.Registry, error) {
	if s.registry != nil {
		return s.registry, nil
	}

	registryOne.Do(func() {
		singletonRegistry, registryErr = require.NewRegistry(s.host.OpenFile)
		if registryErr != nil {
			return
		}

		RegisterNativeModules(singletonRegistry)
	})
	return singletonRegistry, registryErr
}

func RegisterNativeModules(registry *require.Registry) {
	registry.RegisterNativeModule("mokapi", mokapi.Require)
	registry.RegisterNativeModule("mokapi/faker", faker.Require)
	registry.RegisterNativeModule("mokapi/kafka", kafka.Require)
	registry.RegisterNativeModule("mokapi/http", http.Require)
	registry.RegisterNativeModule("mokapi/mustache", mustache.Require)
	registry.RegisterNativeModule("mokapi/yaml", yaml.Require)
	registry.RegisterNativeModule("mokapi/mail", mail.Require)
	registry.RegisterNativeModule("mokapi/ldap", ldap.Require)
}
