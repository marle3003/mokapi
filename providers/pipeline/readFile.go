package pipeline

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
)

type ReadFileStep struct {
}

type ReadFileStepExecution struct {
	File     string `step:"file,position=0,required"`
	AsString bool
}

func (e *ReadFileStep) Start() StepExecution {
	return &ReadFileStepExecution{AsString: true}
}

func (e *ReadFileStepExecution) Run(ctx StepContext) (interface{}, error) {
	env, ok := ctx.Get(EnvVarsType).(EnvVars)
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
