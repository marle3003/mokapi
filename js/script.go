package js

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/pkg/errors"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	engine "mokapi/engine/common"
	"mokapi/js/console"
	"mokapi/js/encoding"
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
	host     engine.Host
	file     *dynamic.Config
	config   static.JsConfig
	loop     *eventloop.EventLoop
	registry *require.Registry
}

func New(file *dynamic.Config, host engine.Host, config static.JsConfig) (*Script, error) {
	s := &Script{
		host:   host,
		file:   file,
		config: config,
	}

	var err error

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

	s.loop.StartLoop()

	result, err := s.loop.RunAsync(func(vm *goja.Runtime) (goja.Value, error) {
		v := vm.Get("module")
		if v == goja.Null() {
			return nil, NoDefaultFunction
		}
		mod := v.ToObject(vm)
		exports := mod.Get("exports").ToObject(vm)
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

	s.loop.Run(fn)
	return nil
}

func (s *Script) Close() {
	if s.loop != nil {
		s.loop.Stop()
	}
	if s.runtime != nil {
		s.runtime.Interrupt(fmt.Errorf("closing"))
		s.runtime = nil
	}
}

func (s *Script) CanClose() bool {
	return !s.loop.HasJobs()
}

func (s *Script) ensureRuntime() error {
	if s.runtime != nil {
		return nil
	}
	s.runtime = goja.New()
	s.loop = eventloop.New(s.runtime)

	s.runtime.SetFieldNameMapper(&customFieldNameMapper{})
	registry, err := s.getRegistry()
	if err != nil {
		return err
	}

	EnableInternal(s.runtime, s.host, s.loop, s.file)

	registry.Enable(s.runtime)
	console.Enable(s.runtime)
	file.Enable(s.runtime, s.host)
	process.Enable(s.runtime)

	prg, err := registry.GetProgram(s.file)
	if err != nil {
		return err
	}

	s.loop.Run(func(vm *goja.Runtime) {
		mod := vm.NewObject()
		mod.Set("exports", vm.NewObject())
		vm.Set("module", mod)

		_, err = vm.RunProgram(prg)
	})
	return err
}

func EnableInternal(vm *goja.Runtime, host engine.Host, loop *eventloop.EventLoop, file *dynamic.Config) {
	o := vm.NewObject()
	o.Set("host", host)
	o.Set("loop", loop)
	o.Set("file", file)
	vm.Set("mokapi/internal", o)
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

func (s *Script) getRegistry() (*require.Registry, error) {
	if s.registry != nil {
		return s.registry, nil
	}

	registryOne.Do(func() {
		singletonRegistry, registryErr = require.NewRegistry()
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
	registry.RegisterNativeModule("mokapi/smtp", mail.Require)
	registry.RegisterNativeModule("mokapi/ldap", ldap.Require)
	registry.RegisterNativeModule("mokapi/encoding", encoding.Require)
}
