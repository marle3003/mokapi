package pipeline

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/types"
	"path/filepath"
)

type Type string

var (
	builtInFunctions = map[string]Step{
		"echo":       &EchoStep{},
		"readYaml":   &YamlStep{},
		"getContext": &GetContextStep{},
		"readFile":   &ReadFileStep{},
		"mustache":   &MustacheStep{},
		"delay":      &DelayStep{},
		"fileExists": &FileExistsStep{},
		"xmlPath":    &XmlPathStep{},
		//"toJson": commands.ToJson,
	}

	EnvVarsType Type = "env"
)

type PipelineOptions func(*context) error

func WithGlobalVars(vars map[Type]interface{}) PipelineOptions {
	return func(ctx *context) error {
		for t, i := range vars {
			o, err := types.Convert(i)
			if err != nil {
				return err
			}
			ctx.vars[string(t)] = o
		}
		return nil
	}
}

func WithSteps(steps map[string]Step) PipelineOptions {
	return func(ctx *context) error {
		for name, step := range steps {
			ctx.steps[name] = step
		}
		return nil
	}
}

func WithParams(params map[string]interface{}) PipelineOptions {
	return func(ctx *context) error {
		var expando *types.Expando
		if e, ok := ctx.vars["params"]; !ok {
			expando = types.NewExpando()
			ctx.vars["params"] = expando
		} else {
			expando = e.(*types.Expando)
		}

		for name, v := range params {
			obj, err := types.Convert(v)
			if err != nil {
				return err
			}
			err = expando.Set(name, obj)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func Run(file string, name string, options ...PipelineOptions) error {

	p, err := getPipeline(file, name)
	if err != nil {
		return errors.Wrapf(err, "unable to read pipeline '%v' from file '%v'", name, file)
	}

	ctx := newContext()
	err = WithGlobalVars(map[Type]interface{}{
		EnvVarsType: NewEnvVars(
			fromOS(),
			With(map[string]string{
				"WORKING_DIRECTORY": filepath.Dir(file),
			})),
	})(ctx)
	if err != nil {
		return err
	}

	err = WithSteps(builtInFunctions)(ctx)
	if err != nil {
		return err
	}

	for _, o := range options {
		err = o(ctx)
		if err != nil {
			return err
		}
	}

	_, err = p.eval(ctx)
	return err
}
