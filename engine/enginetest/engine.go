package enginetest

import (
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/engine/common"
	"mokapi/js"
	"mokapi/js/jstest"
	"path"
)

func NewEngine(opts ...engine.Options) *engine.Engine {
	loader := engine.NewDefaultScriptLoader(&static.Config{})

	opts = append([]engine.Options{
		engine.WithScriptLoader(engine.ScriptLoaderFunc(func(file *dynamic.Config, host common.Host) (common.Script, error) {
			if path.Ext(file.Info.Kernel().Path()) == ".js" {
				return jstest.New(js.WithFile(file), js.WithHost(host))
			}
			return loader.Load(file, host)
		})),
		engine.WithLogger(&Logger{}),
	}, opts...)
	return engine.NewEngine(opts...)
}
