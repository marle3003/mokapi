package enginetest

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/engine/common"
	"mokapi/js"
	"mokapi/js/jstest"
	"mokapi/runtime"
	"path"
)

func NewEngine(opts ...engine.Options) *engine.Engine {
	loader := engine.NewDefaultScriptLoader(&static.Config{})

	opts = append([]engine.Options{
		engine.WithScriptLoader(engine.ScriptLoaderFunc(func(file *dynamic.Config, host common.Host) (common.Script, error) {
			ext := path.Ext(file.Info.Kernel().Path())
			if ext == ".js" || ext == ".ts" {
				// use this loader to ensure not to reuse the JavaScript Registry which is a singleton
				return jstest.New(js.WithFile(file), js.WithHost(host))
			}
			return loader.Load(file, host)
		})),
		engine.WithLogger(&Logger{}),
		engine.WithApp(runtime.New(&static.Config{}, &dynamictest.Reader{})),
	}, opts...)
	return engine.NewEngine(opts...)
}
