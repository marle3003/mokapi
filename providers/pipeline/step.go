package pipeline

type StepContext interface {
	Get(t Type) interface{}
}

type Step interface {
	Start() StepExecution
}

type StepExecution interface {
	Run(ctx StepContext) (interface{}, error)
}
