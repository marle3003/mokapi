package runtime

import (
	"mokapi/config/dynamic/mokapi"
	"time"
)

type Summary struct {
	Workflows []*WorkflowSummary
}

type WorkflowStatus int

const (
	Successful WorkflowStatus = iota
	Error
)

type WorkflowSummary struct {
	Name     string
	Steps    []*StepSummary
	Duration time.Duration
	Status   WorkflowStatus
}

type StepSummary struct {
	Id       string
	Name     string
	Log      string
	Duration time.Duration
	Status   WorkflowStatus
}

func NewStepSummary(step mokapi.Step) *StepSummary {
	name := step.Name
	if len(name) == 0 {
		if len(step.Run) > 0 {
			name = step.Run
		} else {
			name = step.Uses
		}
	}

	return &StepSummary{Name: name, Id: getStepId(step)}
}
