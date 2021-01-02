package pipeline

type GetContextStep struct {
}

type GetContextStepExecution struct {
	Type string `step:"type,position=0,required"`
}

func (e *GetContextStep) Start() StepExecution {
	return &GetContextStepExecution{}
}

func (e *GetContextStepExecution) Run(ctx StepContext) (interface{}, error) {
	o := ctx.Get(Type(e.Type))
	return o, nil
}
