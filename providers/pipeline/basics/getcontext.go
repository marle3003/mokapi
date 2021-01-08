package basics

import (
	"mokapi/providers/pipeline/lang/types"
)

type GetContextStep struct {
	types.AbstractStep
}

type GetContextStepExecution struct {
	Type string `step:"type,position=0,required"`
}

func (e *GetContextStep) Start() types.StepExecution {
	return &GetContextStepExecution{}
}

func (e *GetContextStepExecution) Run(ctx types.StepContext) (interface{}, error) {
	o := ctx.Get(types.Type(e.Type))
	return o, nil
}
