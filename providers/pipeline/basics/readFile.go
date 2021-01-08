package basics

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"mokapi/providers/pipeline/lang/types"
	"mokapi/providers/pipeline/runtime"
	"path/filepath"
)

type ReadFileStep struct {
	types.AbstractStep
}

type ReadFileStepExecution struct {
	File     string `step:"file,position=0,required"`
	AsString bool
}

func (e *ReadFileStep) Start() types.StepExecution {
	return &ReadFileStepExecution{AsString: true}
}

func (e *ReadFileStepExecution) Run(ctx types.StepContext) (interface{}, error) {
	env, ok := ctx.Get(runtime.EnvVarsType).(runtime.EnvVars)
	if !ok {
		return nil, errors.Errorf("env not defined")
	}

	dir, _ := env["WORKING_DIRECTORY"]

	file := filepath.Join(dir, fmt.Sprintf("%v", e.File))

	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if e.AsString {
		return string(bytes), nil
	}

	return bytes, nil
}
