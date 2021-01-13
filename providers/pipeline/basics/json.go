package basics

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"mokapi/providers/pipeline/lang/runtime"
	"mokapi/providers/pipeline/lang/types"
	"path/filepath"
)

type JsonStep struct {
	types.AbstractStep
}

type JsonExecution struct {
	File string `step:"file,required"`
}

func (e *JsonStep) Start() types.StepExecution {
	return &JsonExecution{}
}

func (e *JsonExecution) Run(ctx types.StepContext) (interface{}, error) {
	env, ok := ctx.Get(runtime.EnvVarsType).(runtime.EnvVars)
	if !ok {
		return nil, errors.Errorf("env not defined")
	}

	dir, _ := env["WORKING_DIRECTORY"]

	file := filepath.Join(dir, fmt.Sprintf("%v", e.File))

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var m interface{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing JSON file %s", e.File)
	}

	return convertObject(m), nil
}
