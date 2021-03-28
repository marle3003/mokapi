package basics

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/runtime"
	"mokapi/providers/pipeline/lang/types"
	"os"
	"path/filepath"
)

type FileExistsStep struct {
	types.AbstractStep
}

type FileExistsStepExecution struct {
	File string `step:"file,position=0,required"`
}

func (e *FileExistsStep) Start() types.StepExecution {
	return &FileExistsStepExecution{}
}

func (e *FileExistsStepExecution) Run(ctx types.StepContext) (interface{}, error) {
	env, ok := ctx.Get(runtime.EnvVarsType).(runtime.EnvVars)
	if !ok {
		return nil, errors.Errorf("env not defined")
	}

	dir, _ := env["WORKING_DIRECTORY"]

	file := filepath.Join(dir, fmt.Sprintf("%v", e.File))

	if _, err := os.Stat(file); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}
