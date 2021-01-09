package types

type Type string

type StepContext interface {
	Get(t Type) interface{}
}

type Step interface {
	Start() StepExecution
}

type StepExecution interface {
	Run(ctx StepContext) (interface{}, error)
}

type AbstractStep struct {
	ObjectImpl
}
