package require

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
	"mokapi/js/compiler"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	ModuleFileNotFound = errors.New("module not found")
)

const jsonParseFunc = "export default JSON.parse('%v')"

type Option func(m *Module)

type ModuleLoader func() goja.Value

type SourceLoader func(file, hint string) (string, string, error)

type Module struct {
	native       map[string]ModuleLoader
	sourceLoader SourceLoader
	compiler     *compiler.Compiler
	workingDir   string
	runtime      *goja.Runtime
	exports      map[string]goja.Value
}

func New(opts ...Option) *Module {
	m := &Module{native: map[string]ModuleLoader{}, exports: map[string]goja.Value{}}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func WithNativeModule(name string, loader ModuleLoader) Option {
	return func(m *Module) {
		m.native[name] = loader
	}
}

func WithCompiler(compiler *compiler.Compiler) Option {
	return func(m *Module) {
		m.compiler = compiler
	}
}

func WithWorkingDir(dir string) Option {
	return func(m *Module) {
		m.workingDir = dir
	}
}

func WithSourceLoader(loader SourceLoader) Option {
	return func(m *Module) {
		m.sourceLoader = loader
	}
}

func (m *Module) Enable(rt *goja.Runtime) {
	m.runtime = rt
	if err := rt.Set("require", m.require); err != nil {
		log.Errorf("enabling require module: %v", err)
	}
}

func (m *Module) require(call goja.FunctionCall) (module goja.Value) {
	modPath := call.Argument(0).String()
	if len(modPath) == 0 {
		panic(m.runtime.ToValue("missing argument"))
	}

	if e, ok := m.exports[modPath]; ok {
		return e
	}
	if loader, ok := m.native[modPath]; ok {
		v := loader()
		m.exports[modPath] = v
		return v
	}

	var err error
	if strings.HasPrefix(modPath, "./") || strings.HasPrefix(modPath, "../") || strings.HasPrefix(modPath, "/") {
		if module, err = m.loadFileModule(modPath); err == nil && module != nil {
			m.exports[modPath] = module
		}
	} else {
		if module, err = m.loadFileModule(modPath); err == nil && module != nil {
			m.exports[modPath] = module
		} else if module, err = m.loadNodeModule(modPath); err == nil && module != nil {
			m.exports[modPath] = module
		}
	}

	if module == nil {
		panic(m.runtime.ToValue(fmt.Sprintf("module %v not found", modPath)))
	}

	return
}

func (m *Module) loadFileModule(modPath string) (goja.Value, error) {
	if len(filepath.Ext(modPath)) > 0 {
		p, src, err := m.sourceLoader(modPath, m.workingDir)
		if err == nil {
			return m.loadModule(p, src)
		}
		if filepath.Ext(modPath) == ".json" {
			return m.loadModule(p, fmt.Sprintf(jsonParseFunc, template.JSEscapeString(src)))
		}
	} else {
		if v, err := m.loadFileModule(modPath + ".js"); err == nil {
			return v, nil
		}

		if v, err := m.loadFileModule(modPath + ".json"); err == nil {
			return v, nil
		}
	}

	return nil, ModuleFileNotFound
}

func (m *Module) loadNodeModule(mod string) (goja.Value, error) {
	dir := m.workingDir
	for len(dir) > 0 {
		path := filepath.Join(dir, "node_modules", mod)
		_, src, err := m.sourceLoader(filepath.Join(path, "package.json"), m.workingDir)
		if err != nil {
			if v, err := m.loadFileModule(filepath.Join(path, "index.js")); err == nil {
				return v, nil
			}
		} else {
			if v, err := m.loadFromPackage(path, src); err == nil {
				return v, nil
			}

			if v, err := m.loadFileModule(filepath.Join(path, "index.js")); err == nil {
				return v, nil
			}
		}

		if dir == string(filepath.Separator) {
			break
		}

		current := dir
		dir = filepath.Dir(dir)
		if dir == current {
			break
		}
	}
	return nil, fmt.Errorf("node module does not exist")
}

func (m *Module) loadFromPackage(modPath, src string) (goja.Value, error) {
	pkg := struct {
		Main string
	}{}
	err := json.Unmarshal([]byte(src), &pkg)
	if err != nil {
		return nil, fmt.Errorf("unable to parse package.json")
	}

	return m.loadFileModule(filepath.Join(modPath, pkg.Main))
}

func (m *Module) loadModule(modPath, source string) (goja.Value, error) {
	oldPath := m.workingDir
	m.workingDir = filepath.Dir(modPath)
	defer func() { m.workingDir = oldPath }()

	module := m.runtime.NewObject()
	exports := m.runtime.NewObject()
	if err := module.Set("exports", exports); err != nil {
		panic(fmt.Sprintf("unable to import module %v: %v", modPath, err))
	}

	prg, err := m.compiler.CompileModule(modPath, source)
	if err != nil {
		panic(err)
	}
	f, err := m.runtime.RunProgram(prg)
	if err != nil {
		panic(err)
	}
	if call, ok := goja.AssertFunction(f); ok {
		_, err = call(exports, exports, module)
		if err != nil {
			panic(err)
		}
	}

	return exports, nil
}

func (m *Module) Close() {
	m.native = nil
	m.exports = nil
}
