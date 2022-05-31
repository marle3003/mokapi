package compiler

import (
	_ "embed"
	"github.com/dop251/goja"
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
		babelPrg, babelErr = goja.Compile("mokapi/babel.min.js", babelSource, false)
	})
	if babelErr != nil {
		return nil, babelErr
	}

	rt := goja.New()

	logFunc := func(goja.FunctionCall) goja.Value { return nil }
	rt.Set("console", map[string]func(goja.FunctionCall) goja.Value{
		"log":   logFunc,
		"error": logFunc,
		"warn":  logFunc,
	})

	_, err := rt.RunProgram(babelPrg)
	if err != nil {
		return nil, err
	}

	jObj := rt.Get("Babel")

	babel := &babel{
		runtime: rt,
		this:    jObj,
		mutex:   sync.Mutex{},
	}

	if err = rt.ExportTo(jObj.ToObject(rt).Get("transform"), &babel.transform); err != nil {
		return nil, err
	}

	return babel, nil
}

func (b *babel) Transform(src string) (code string, err error) {
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
		"filename":      "babel.js"}
	v, err := b.transform(b.this, b.runtime.ToValue(src), b.runtime.ToValue(opts))
	if err != nil {
		return
	}

	o := v.ToObject(b.runtime)
	code = o.Get("code").String()

	return
}
