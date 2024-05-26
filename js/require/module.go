package require

import (
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"net/url"
	"path/filepath"
)

var (
	supportedExtensions = []string{".js", ".ts", ".json", ".yaml", ".yml"}
	ModuleFileNotFound  = errors.New("module not found")
)

type module struct {
	registry *Registry
	host     SourceLoader

	vm      *goja.Runtime
	modules map[string]*goja.Object

	currentSource *dynamic.Config
}

func (m *module) require(call goja.FunctionCall) (module goja.Value) {
	mod := m.requireModule(call.Arguments[0].String())
	return mod.Get("exports")
}

func (m *module) requireModule(modPath string) *goja.Object {
	if len(modPath) == 0 {
		panic(m.vm.ToValue("missing argument"))
	}
	cmp := m.getCurrentModulePath()
	key := fmt.Sprintf("%v:%v", filepath.Dir(cmp), modPath)

	if v, ok := m.modules[key]; ok {
		return v
	}
	if loader, ok := m.registry.native[modPath]; ok {
		mod := m.createModuleObject()
		loader(m.vm, mod)
		m.modules[key] = mod
		return mod
	}
	if u, err := url.Parse(modPath); err == nil && len(u.Scheme) > 0 {
		src, err := m.host.OpenFile(modPath, "")
		if err == nil {
			if mod, err := m.loadModule(modPath, src); err == nil && mod != nil {
				m.modules[key] = mod
				return mod
			}
		}
	}

	dir := filepath.Dir(cmp)
	mod, err := m.loadFileModule(filepath.Join(dir, modPath))
	if err == nil && mod != nil {
		return mod
	} else if !errors.Is(err, ModuleFileNotFound) {
		panic(m.vm.ToValue(fmt.Sprintf("loaded module %v contains error: %v", modPath, err)))
	}

	mod, err = m.loadNodeModule(modPath, dir)
	if err == nil && mod != nil {
		return mod
	} else if !errors.Is(err, ModuleFileNotFound) {
		panic(m.vm.ToValue(fmt.Sprintf("loaded module %v contains error: %v", modPath, err)))
	}

	panic(m.vm.ToValue(fmt.Sprintf("module %v not found in %v", modPath, cmp)))
}

func (m *module) loadFileModule(modPath string) (*goja.Object, error) {
	if len(filepath.Ext(modPath)) > 0 {
		src, err := m.host.OpenFile(modPath, "")
		if err != nil {
			return nil, ModuleFileNotFound
		}

		if filepath.Ext(modPath) == ".yaml" {
			return m.loadYaml(string(src.Raw))
		}

		return m.loadModule(modPath, src)
	}

	for _, ext := range supportedExtensions {
		p := modPath + ext
		if mod, err := m.loadFileModule(p); err == nil {
			return mod, nil
		} else if !errors.Is(err, ModuleFileNotFound) {
			return nil, err
		}
	}

	return m.loadDirectoryModule(modPath)
}

func (m *module) loadDirectoryModule(modPath string) (*goja.Object, error) {
	if len(modPath) == 0 {
		return nil, ModuleFileNotFound
	}
	if mod, err := m.loadFromPackageFile(modPath); err == nil {
		return mod, err
	}
	if mod, err := m.loadFileModule(filepath.Join(modPath, "index.js")); err == nil {
		return mod, nil
	}
	if mod, err := m.loadFileModule(filepath.Join(modPath, "index.ts")); err == nil {
		return mod, nil
	}
	if mod, err := m.loadFileModule(filepath.Join(modPath, "index.json")); err == nil {
		return mod, nil
	}

	return nil, ModuleFileNotFound
}

func (m *module) loadFromPackageFile(modPath string) (*goja.Object, error) {
	src, err := m.host.OpenFile(filepath.Join(modPath, "package.json"), "")
	if err != nil {
		return nil, err
	}

	pkg := struct {
		Main string
	}{}
	err = json.Unmarshal(src.Raw, &pkg)
	if err != nil {
		return nil, fmt.Errorf("unable to parse package.json")
	}

	modPath = filepath.Join(modPath, pkg.Main)
	return m.loadFileModule(modPath)
}

func (m *module) loadNodeModule(modPath, dir string) (*goja.Object, error) {
	for len(dir) > 0 {
		p := filepath.Join(dir, "node_modules", modPath)
		if mod, err := m.loadDirectoryModule(p); err == nil {
			return mod, nil
		}
		if p == string(filepath.Separator) {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return nil, ModuleFileNotFound
}

func (m *module) loadModule(modPath string, source *dynamic.Config) (*goja.Object, error) {
	prg, err := m.registry.getModuleProgram(modPath, source)
	if err != nil {
		return nil, err
	}

	parent := m.currentSource
	dynamic.AddRef(parent, source)
	m.currentSource = source
	defer func() {
		m.currentSource = parent
	}()

	f, err := m.vm.RunProgram(prg)
	if err != nil {
		return nil, err
	}
	if call, ok := goja.AssertFunction(f); ok {
		mod := m.createModuleObject()
		exports := mod.Get("exports")
		require := m.vm.Get("require")
		_, err = call(exports, exports, mod, require)
		return mod, err
	} else {
		return nil, fmt.Errorf("invalid module")
	}
}

func (m *module) getCurrentModulePath() string {
	var buf [2]goja.StackFrame
	frames := m.vm.CaptureCallStack(2, buf[:0])
	if len(frames) < 2 {
		return "."
	}
	return frames[1].SrcName()
}

func (m *module) createModuleObject() *goja.Object {
	mod := m.vm.NewObject()
	mod.Set("exports", m.vm.NewObject())
	return mod
}

func (m *module) loadYaml(source string) (*goja.Object, error) {
	mod := m.createModuleObject()
	result := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(source), result)
	if err != nil {
		return nil, err
	}
	mod.Set("exports", m.vm.ToValue(result))
	return mod, nil
}
