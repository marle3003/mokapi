package basics

import (
	"fmt"
	"mokapi/providers/pipeline/lang/types"
)

type EchoStep struct {
	types.AbstractStep
}

type EchoExecution struct {
	Message string `step:"message,position=0,required"`
}

func (e *EchoStep) Start() types.StepExecution {
	return &EchoExecution{}
}

func (e *EchoExecution) Run(_ types.StepContext) (interface{}, error) {
	fmt.Println(e.Message)
	return nil, nil
}
