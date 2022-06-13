package js

import (
	"fmt"
	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
	engine "mokapi/engine/common"
	"mokapi/js/common"
	"mokapi/js/modules"
	"mokapi/js/modules/faker"
	"mokapi/js/modules/http"
	"mokapi/js/modules/kafka"
	"mokapi/js/modules/mustache"
	"mokapi/js/modules/yaml"
	"path/filepath"
	"text/template"
)

type factory func(engine.Host, *goja.Runtime) interface{}

var moduleTypes = map[string]factory{
	"mokapi":   modules.NewMokapi,
	"faker":    faker.New,
	"http":     http.New,
	"kafka":    kafka.New,
	"yaml":     yaml.New,
	"mustache": mustache.New,
}

type require struct {
	exports map[string]goja.Value
	runtime *goja.Runtime
	host    engine.Host
	open    func(filename, src string) (goja.Value, error)
}

func enableRequire(script *Script, host engine.Host) *require {
	r := &require{
		runtime: script.runtime,
		host:    host,
		exports: make(map[string]goja.Value),
		open:    script.requireFile,
	}
	script.runtime.Set("require", r.require)

	return r
}

func (r *require) require(call goja.FunctionCall) goja.Value {
	file := call.Argument(0).String()
	if len(file) == 0 {
		panic(r.runtime.ToValue("missing argument"))
	}

	if e, ok := r.exports[file]; ok {
		return e
	} else if f, ok := moduleTypes[file]; ok {
		m := f(r.host, r.runtime)
		e := common.Map(r.runtime, m)
		r.exports[file] = e
		return e
	} else {
		src, err := r.loadModule(file)
		if err != nil {
			log.Errorf("unable to load module %v: %v", file, err)
			return goja.Null()
		}

		export, err := r.open(file, src)
		if err != nil {
			panic(err)
		}
		r.exports[file] = export
		return export
	}
}

func (r *require) close() {
	r.exports = nil
	r.runtime = nil
}

const json = "export default JSON.parse('%v')"

func (r *require) loadModule(file string) (string, error) {
	path := file

	if len(filepath.Ext(path)) > 0 {
		src, err := r.host.OpenScript(path)
		if err == nil && filepath.Ext(path) == ".json" {
			return fmt.Sprintf(json, template.JSEscapeString(src)), nil
		}
		return src, err
	}

	path = file + ".js"
	src, err := r.host.OpenScript(path)
	if err == nil {
		return src, nil
	}

	path = file + ".json"
	src, err = r.host.OpenScript(path)
	if err == nil {
		return fmt.Sprintf(json, template.JSEscapeString(src)), nil
	}

	return "", err
}
