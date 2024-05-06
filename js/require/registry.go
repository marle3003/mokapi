package require

import (
	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/js/compiler"
	"path/filepath"
	"sync"
	"text/template"
)

type ModuleLoader func(vm *goja.Runtime, module *goja.Object)

type SourceLoader interface {
	OpenFile(file, hint string) (*dynamic.Config, error)
}

type Registry struct {
	native  map[string]ModuleLoader
	modules map[string]*goja.Program
	scripts map[string]*goja.Program

	compiler *compiler.Compiler

	m sync.Mutex
}

func NewRegistry() (*Registry, error) {
	reg := &Registry{
		native:  map[string]ModuleLoader{},
		modules: map[string]*goja.Program{},
		scripts: map[string]*goja.Program{},
	}
	var err error
	reg.compiler, err = compiler.New()
	return reg, err
}

func (r *Registry) Enable(vm *goja.Runtime) {
	o := vm.Get("mokapi/internal").(*goja.Object)
	host := o.Get("host").Export().(common.Host)

	m := &module{
		registry: r,
		host:     host,
		vm:       vm,
		modules:  map[string]*goja.Object{},
	}
	if err := vm.Set("require", m.require); err != nil {
		log.Errorf("enabling require module: %v", err)
	}
}

func (r *Registry) RegisterNativeModule(name string, loader ModuleLoader) {
	r.m.Lock()
	defer r.m.Unlock()

	r.native[name] = loader
}

func (r *Registry) getModuleProgram(modPath, source string) (*goja.Program, error) {
	r.m.Lock()
	defer r.m.Unlock()

	prg := r.modules[modPath]
	if prg == nil {
		if filepath.Ext(modPath) == ".json" {
			source = "module.exports = JSON.parse('" + template.JSEscapeString(source) + "')"
		}

		var err error
		prg, err = r.compiler.CompileModule(modPath, source)
		if err != nil {
			return nil, err
		}
		r.modules[modPath] = prg
	}
	return prg, nil
}

func (r *Registry) GetProgram(path, source string) (*goja.Program, error) {
	r.m.Lock()
	defer r.m.Unlock()

	prg := r.scripts[path]
	if prg == nil {
		var err error
		prg, err = r.compiler.Compile(path, source)
		if err != nil {
			return nil, err
		}
		r.scripts[path] = prg
	}
	return prg, nil
}
