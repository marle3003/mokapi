package basics

import (
	"math/rand"
	"mokapi/providers/pipeline/lang/types"
)

type RandomStep struct {
	types.AbstractStep
}

type RandomExecution struct {
	Type string `step:"type"`
	Max  int    `step:"max"`
}

func (e *RandomStep) Start() types.StepExecution {
	return &RandomExecution{}
}

func (e *RandomExecution) Run(_ types.StepContext) (interface{}, error) {
	switch e.Type {
	case "float":
		return rand.Float64(), nil
	default:
		if e.Max <= 0 {
			return rand.Int(), nil
		}
		return rand.Intn(e.Max), nil

	}
	return nil, nil
}
