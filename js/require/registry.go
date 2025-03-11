package require

import (
	"bytes"
	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/js/compiler"
	"net/url"
	"path/filepath"
	"sync"
	"text/template"
)

type ModuleLoader func(vm *goja.Runtime, module *goja.Object)

type SourceLoader interface {
	OpenFile(file, hint string) (*dynamic.Config, error)
}

type entry struct {
	program *goja.Program
	hash    []byte
	m       sync.Mutex
}

type Registry struct {
	native  map[string]ModuleLoader
	modules map[string]*entry
	scripts map[string]*entry

	compiler *compiler.Compiler

	m sync.Mutex
}

func NewRegistry() (*Registry, error) {
	reg := &Registry{
		native:  map[string]ModuleLoader{},
		modules: map[string]*entry{},
		scripts: map[string]*entry{},
	}
	var err error
	reg.compiler, err = compiler.New()
	return reg, err
}

func (r *Registry) Enable(vm *goja.Runtime) {
	o := vm.Get("mokapi/internal").(*goja.Object)
	host := o.Get("host").Export().(common.Host)
	file := o.Get("file").Export().(*dynamic.Config)

	m := &module{
		registry:      r,
		host:          host,
		vm:            vm,
		modules:       map[string]*goja.Object{},
		currentSource: file,
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

func (r *Registry) getModuleProgram(modPath string, file *dynamic.Config) (*goja.Program, error) {
	r.m.Lock()
	e, found := r.modules[modPath]
	if !found {
		e = &entry{hash: file.Info.Checksum}
		r.modules[modPath] = e
	}
	r.m.Unlock()

	if e.program != nil && bytes.Equal(e.hash, file.Info.Checksum) {
		return e.program, nil
	}

	e.m.Lock()
	defer e.m.Unlock()

	if e.program == nil || !bytes.Equal(e.hash, file.Info.Checksum) {
		source := string(file.Raw)
		if filepath.Ext(modPath) == ".json" {
			source = "module.exports = JSON.parse('" + template.JSEscapeString(source) + "')"
		}

		prg, err := r.compiler.CompileModule(file.Info.Kernel().Path(), source)
		if err != nil {
			return nil, err
		}
		e.program = prg
		e.hash = file.Info.Checksum
	}
	return e.program, nil
}

func (r *Registry) GetProgram(file *dynamic.Config) (*goja.Program, error) {
	path := getScriptPath(file.Info.Kernel().Url)

	r.m.Lock()
	e, found := r.scripts[path]
	if !found {
		e = &entry{hash: file.Info.Checksum}
		r.scripts[path] = e
	}
	r.m.Unlock()

	if e.program != nil && bytes.Equal(e.hash, file.Info.Checksum) {
		return e.program, nil
	}

	e.m.Lock()
	defer e.m.Unlock()

	if e.program == nil || !bytes.Equal(e.hash, file.Info.Checksum) {
		prg, err := r.compiler.Compile(path, string(file.Raw))
		if err != nil {
			return nil, err
		}
		e.program = prg
		e.hash = file.Info.Checksum
	}
	return e.program, nil
}

func getScriptPath(u *url.URL) string {
	if len(u.Path) > 0 {
		return u.Path
	}
	return u.Opaque
}
