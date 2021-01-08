package basics

import (
	"fmt"
	"mokapi/providers/pipeline/lang/types"
)

type EnvStep struct {
	types.AbstractStep
}

type EnvExecution struct {
	Message string `step:"message,position=0,required"`
}

func (e *EnvStep) Start() types.StepExecution {
	return &EnvExecution{}
}

func (e *EnvExecution) Run(_ types.StepContext) (interface{}, error) {
	fmt.Println(e.Message)
	return nil, nil
}
