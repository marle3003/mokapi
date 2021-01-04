package pipeline

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

type FileExistsStep struct {
}

type FileExistsStepExecution struct {
	File     string `step:"file,position=0,required"`
	AsString bool
}

func (e *FileExistsStep) Start() StepExecution {
	return &FileExistsStepExecution{AsString: true}
}

func (e *FileExistsStepExecution) Run(ctx StepContext) (interface{}, error) {
	env, ok := ctx.Get(EnvVarsType).(EnvVars)
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
