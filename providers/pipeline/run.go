package pipeline

import (
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

func Run(file string, name string, options ...PipelineOptions) error {

	p, err := getPipeline(file, name)
	if err != nil {
		return err
	}

	ctx := newContext()
	WithGlobalVars(map[Type]interface{}{
		EnvVarsType: NewEnvVars(
			fromOS(),
			With(map[string]string{
				"WORKING_DIRECTORY": filepath.Dir(file),
			})),
	})(ctx)
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
