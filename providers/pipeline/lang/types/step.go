package types

import (
	"reflect"
)

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
}

func (s *AbstractStep) String() string {
	return s.GetType().String()
}

func (s *AbstractStep) GetField(name string) (Object, error) {
	return getField(s, name)
}

func (s *AbstractStep) GetType() reflect.Type {
	return reflect.TypeOf(s)
}
