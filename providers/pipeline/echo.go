package pipeline

import (
	"fmt"
)

type EchoStep struct {
}

type EchoExecution struct {
	Message string `step:"message,position=0,required"`
}

func (e *EchoStep) Start() StepExecution {
	return &EchoExecution{}
}

func (e *EchoExecution) Run(_ StepContext) (interface{}, error) {
	fmt.Println(e.Message)
	return nil, nil
}
