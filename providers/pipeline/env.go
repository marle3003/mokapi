package pipeline

import (
	"fmt"
)

type EnvStep struct {
}

type EnvExecution struct {
	Message string `step:"message,position=0,required"`
}

func (e *EnvStep) Start() StepExecution {
	return &EnvExecution{}
}

func (e *EnvExecution) Run(_ StepContext) (interface{}, error) {
	fmt.Println(e.Message)
	return nil, nil
}
