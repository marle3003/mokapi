package pipeline

import (
	"io/ioutil"
	"mokapi/providers/pipeline/basics"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/parser"
	"mokapi/providers/pipeline/lang/runtime"
	"mokapi/providers/pipeline/lang/types"
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
		"readJson":   &basics.JsonStep{},
		"random":     &basics.RandomStep{},
	}
)

type PipelineOptions func(scope *ast.Scope) error

func WithGlobalVars(vars map[types.Type]interface{}) PipelineOptions {
	return func(ctx *ast.Scope) error {
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
	return func(ctx *ast.Scope) error {
		for name, step := range steps {
			ctx.SetSymbol(name, step.(types.Object))
		}
		return nil
	}
}

func WithParams(params map[string]interface{}) PipelineOptions {
	return func(ctx *ast.Scope) error {
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
			expando.SetField(name, obj)
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

	scope := ast.NewScope(builtInFunctions)
	err = WithGlobalVars(map[types.Type]interface{}{
		runtime.EnvVarsType: runtime.NewEnvVars(
			runtime.FromOS(),
			runtime.With(map[string]string{
				"WORKING_DIRECTORY": filepath.Dir(file),
			})),
	})(scope)
	for _, o := range options {
		err = o(scope)
		if err != nil {
			return err
		}
	}

	var f *ast.File
	f, err = parser.ParseFile(src, scope)
	if err != nil {
		return
	}

	return runtime.RunPipeline(f, name)
}
