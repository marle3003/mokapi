package pipeline

import (
	"io/ioutil"
	"mokapi/providers/pipeline/basics"
	"mokapi/providers/pipeline/lang"
	"mokapi/providers/pipeline/lang/types"
	"mokapi/providers/pipeline/runtime"
	"path/filepath"
)

var (
	builtInFunctions = map[string]types.Object{
		"echo":       &basics.EchoStep{},
		"readYaml":   &basics.YamlStep{},
		"getContext": &basics.GetContextStep{},
		"readFile":   &basics.ReadFileStep{},
		"mustache":   &basics.MustacheStep{},
		"delay":      &basics.DelayStep{},
		"fileExists": &basics.FileExistsStep{},
		"xmlPath":    &basics.XmlPathStep{},
	}
)

type PipelineOptions func(scope *runtime.Scope) error

func WithGlobalVars(vars map[types.Type]interface{}) PipelineOptions {
	return func(ctx *runtime.Scope) error {
		for t, i := range vars {
			o, err := types.Convert(i)
			if err != nil {
				return err
			}
			ctx.SetSymbol(string(t), o)
		}
		return nil
	}
}

func WithSteps(steps map[string]types.Step) PipelineOptions {
	return func(ctx *runtime.Scope) error {
		for name, step := range steps {
			ctx.SetSymbol(name, step.(types.Object))
		}
		return nil
	}
}

func WithParams(params map[string]interface{}) PipelineOptions {
	return func(ctx *runtime.Scope) error {
		var expando *types.Expando
		if v, ok := ctx.Symbol("params"); !ok {
			expando = types.NewExpando()
			ctx.SetSymbol("params", expando)
		} else {
			expando = v.(*types.Expando)
		}

		for name, v := range params {
			obj, err := types.Convert(v)
			if err != nil {
				return err
			}
			expando.Set(name, obj)
		}

		return nil
	}
}

func Run(file, name string, options ...PipelineOptions) (err error) {
	var src []byte
	src, err = ioutil.ReadFile(file)
	if err != nil {
		return
	}

	var f *lang.File
	f, err = lang.ParseFile(src)
	if err != nil {
		return
	}

	context := runtime.NewScope(builtInFunctions)
	err = WithGlobalVars(map[types.Type]interface{}{
		runtime.EnvVarsType: runtime.NewEnvVars(
			runtime.FromOS(),
			runtime.With(map[string]string{
				"WORKING_DIRECTORY": filepath.Dir(file),
			})),
	})(context)
	for _, o := range options {
		err = o(context)
		if err != nil {
			return err
		}
	}

	return runtime.RunPipeline(f, name, context)
}
