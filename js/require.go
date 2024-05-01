package js

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/js/compiler"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	ModuleFileNotFound = errors.New("module not found")
)

const jsonParseFunc = "export default JSON.parse('%v')"

type Option func(m *requireModule)

type ModuleLoader func() goja.Value

type SourceLoader func(file, hint string) (*dynamic.Config, error)

type requireModule struct {
	native        map[string]ModuleLoader
	sourceLoader  SourceLoader
	compiler      *compiler.Compiler
	workingDir    string
	file          string
	runtime       *goja.Runtime
	exports       map[string]goja.Value
	globalFolders []string
}

func newRequire(loader SourceLoader, c *compiler.Compiler, file string, workingDir string, native map[string]ModuleLoader) *requireModule {
	m := &requireModule{
		native:       native,
		exports:      map[string]goja.Value{},
		sourceLoader: loader,
		compiler:     c,
		file:         file,
		workingDir:   workingDir,
	}
	return m
}

func (m *requireModule) Enable(rt *goja.Runtime) {
	m.runtime = rt
	if err := rt.Set("require", m.require); err != nil {
		log.Errorf("enabling require module: %v", err)
	}
}

func (m *requireModule) require(call goja.FunctionCall) (module goja.Value) {
	modPath := call.Argument(0).String()
	if len(modPath) == 0 {
		panic(m.runtime.ToValue("missing argument"))
	}

	key := fmt.Sprintf("%v:%v", m.workingDir, modPath)
	if e, ok := m.exports[key]; ok {
		return e
	}
	if loader, ok := m.native[modPath]; ok {
		v := loader()
		m.exports[modPath] = v
		return v
	}
	var err error
	var u *url.URL
	if u, err = url.Parse(modPath); err == nil && len(u.Scheme) > 0 {
		f, err := m.sourceLoader(modPath, "")
		if err == nil {
			if module, err = m.loadModule(modPath, string(f.Raw)); err != nil && module != nil {
				m.exports[key] = module
			}
		}
	} else if strings.HasPrefix(modPath, "./") || strings.HasPrefix(modPath, "../") || strings.HasPrefix(modPath, "/") {
		if module, err = m.loadFileModule(modPath); err == nil && module != nil {
			m.exports[key] = module
		}
	} else {
		if module, err = m.loadFileModule(modPath); err == nil && module != nil {
			m.exports[key] = module
		} else if module, err = m.loadNodeModule(modPath); err == nil && module != nil {
			m.exports[key] = module
		}
	}

	if module == nil {
		panic(m.runtime.ToValue(fmt.Sprintf("module %v not found in %v: %v", modPath, m.file, err)))
	}

	return
}

func (m *requireModule) loadFileModule(modPath string) (goja.Value, error) {
	if len(filepath.Ext(modPath)) > 0 {
		f, err := m.sourceLoader(modPath, m.workingDir)
		if err != nil {
			return nil, err
		}
		fileUrl := f.Info.Kernel().Url
		if filepath.Ext(modPath) == ".json" {
			return m.loadModule(fileUrl.String(), fmt.Sprintf(jsonParseFunc, template.JSEscapeString(string(f.Raw))))
		} else if filepath.Ext(modPath) == ".yaml" {
			result := make(map[string]interface{})
			err := yaml.Unmarshal(f.Raw, result)
			if err != nil {
				return nil, err
			}
			return m.runtime.ToValue(result), nil
		}
		path := fileUrl.Path
		if len(fileUrl.Opaque) > 0 {
			path = fileUrl.Opaque
		}
		return m.loadModule(path, string(f.Raw))
	} else {
		files := []string{
			modPath + ".js",
			modPath + ".ts",
			filepath.Join(modPath, "index.js"),
			filepath.Join(modPath, "index.ts"),
			modPath + ".json",
			modPath + ".yaml",
		}
		return m.tryLoadFileModule(files...)
	}
}

func (m *requireModule) loadNodeModule(mod string) (goja.Value, error) {
	oldPath := m.workingDir
	defer func() { m.workingDir = oldPath }()

	for _, dir := range m.globalFolders {
		m.workingDir = dir
		if mod, err := m.loadNodeModule(mod); mod != nil && err == nil {
			return mod, nil
		}
	}
	m.workingDir = oldPath

	for len(m.workingDir) > 0 {
		current := m.workingDir
		m.workingDir = filepath.Join(m.workingDir, "node_modules", mod)
		f, err := m.sourceLoader("package.json", m.workingDir)
		if err != nil {
			if v, err := m.loadFileModule("index.js"); err == nil {
				return v, nil
			}
		} else {
			if v, err := m.loadFromPackage(m.workingDir, string(f.Raw)); err == nil {
				return v, nil
			}

			if v, err := m.loadFileModule("index.js"); err == nil {
				return v, nil
			}
		}

		if m.workingDir == string(filepath.Separator) {
			break
		}

		m.workingDir = filepath.Dir(current)
		if m.workingDir == current {
			break
		}
	}
	return nil, fmt.Errorf("node module does not exist")
}

func (m *requireModule) loadFromPackage(modPath, src string) (goja.Value, error) {
	pkg := struct {
		Main string
	}{}
	err := json.Unmarshal([]byte(src), &pkg)
	if err != nil {
		return nil, fmt.Errorf("unable to parse package.json")
	}

	return m.loadFileModule(pkg.Main)
}

func (m *requireModule) loadModule(modPath, source string) (goja.Value, error) {
	oldPath := m.workingDir
	m.workingDir = filepath.Dir(modPath)
	defer func() { m.workingDir = oldPath }()

	module := m.runtime.NewObject()
	_ = module.Set("workingDir", m.workingDir)
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

func (m *requireModule) tryLoadFileModule(files ...string) (goja.Value, error) {
	for _, file := range files {
		if v, err := m.loadFileModule(file); err == nil {
			return v, nil
		}
	}
	return nil, ModuleFileNotFound
}

func (m *requireModule) Close() {
	m.native = nil
	m.exports = nil
}
