package basics

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"mokapi/providers/pipeline/lang/runtime"
	"mokapi/providers/pipeline/lang/types"
	"path/filepath"
)

type YamlStep struct {
	types.AbstractStep
}

type YamlExecution struct {
	File       string `step:"file,required"`
	AsTemplate bool
}

func (e *YamlStep) Start() types.StepExecution {
	return &YamlExecution{}
}

func (e *YamlExecution) Run(ctx types.StepContext) (interface{}, error) {
	env, ok := ctx.Get(runtime.EnvVarsType).(runtime.EnvVars)
	if !ok {
		return nil, errors.Errorf("env not defined")
	}

	dir, _ := env["WORKING_DIRECTORY"]

	file := filepath.Join(dir, fmt.Sprintf("%v", e.File))

	data, err := readFile(file, e.AsTemplate)
	if err != nil {
		return nil, err
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing YAML file %s", e.File)
	}

	return convertObject(m), nil
}
