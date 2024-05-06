package compiler

import (
	_ "embed"
	"github.com/dop251/goja"
	"path/filepath"
	"sync"
)

//go:embed babel.min.js
var babelSource string

var (
	babelPrg *goja.Program
	babelErr error
	babelOne sync.Once
)

type babel struct {
	runtime   *goja.Runtime
	this      goja.Value
	transform goja.Callable
	mutex     sync.Mutex
}

func newBabel() (*babel, error) {
	babelOne.Do(func() {
		babelPrg, babelErr = goja.Compile("<mokapi/babel.min.js>", babelSource, false)
	})
	if babelErr != nil {
		return nil, babelErr
	}

	vm := goja.New()

	logFunc := func(goja.FunctionCall) goja.Value { return nil }
	vm.Set("console", map[string]func(goja.FunctionCall) goja.Value{
		"log":   logFunc,
		"error": logFunc,
		"warn":  logFunc,
	})

	_, err := vm.RunProgram(babelPrg)
	if err != nil {
		return nil, err
	}

	jObj := vm.Get("Babel")

	babel := &babel{
		runtime: vm,
		this:    jObj,
		mutex:   sync.Mutex{},
	}

	if err = vm.ExportTo(jObj.ToObject(vm).Get("transform"), &babel.transform); err != nil {
		return nil, err
	}

	return babel, nil
}

func (b *babel) Transform(filename, src string) (code string, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	opts := map[string]interface{}{
		"presets": []string{"env"},
		"plugins": []interface{}{
			"transform-exponentiation-operator",
		},
		"ast":           false,
		"sourceMaps":    false,
		"babelrc":       false,
		"compact":       false,
		"retainLines":   true,
		"highlightCode": false,
		"filename":      filename}

	if filepath.Ext(filename) == ".ts" {
		opts["presets"] = []string{"env", "typescript"}
		opts["plugins"] = []interface{}{
			"transform-exponentiation-operator",
			"transform-async-to-generator",
		}
	}

	v, err := b.transform(b.this, b.runtime.ToValue(src), b.runtime.ToValue(opts))
	if err != nil {
		return
	}

	o := v.ToObject(b.runtime)
	code = o.Get("code").String()

	return
}
