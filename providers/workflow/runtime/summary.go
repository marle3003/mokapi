package runtime

import (
	"fmt"
	"mokapi/config/dynamic/mokapi"
	"time"
)

type Summary struct {
	Workflows []*WorkflowSummary
}

type WorkflowStatus int

const (
	Successful WorkflowStatus = iota
	Skip
	Error
)

type WorkflowSummary struct {
	Name     string
	Steps    []*StepSummary
	Duration time.Duration
	Status   WorkflowStatus
}

type Log []string

type StepSummary struct {
	Id       string
	Name     string
	Log      Log
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

	return &StepSummary{Name: name, Id: getStepId(step), Log: make([]string, 0)}
}

func (l *Log) Append(s string) {
	*l = append(*l, s)
}

func (l *Log) AppendRange(s []string) {
	*l = append(*l, s...)
}

func (l *Log) AppendGroup(name string, log Log) {
	l.Append("##[group]" + name)
	*l = append(*l, log...)
	l.Append("##[endgroup]")
}

func newLog(format string, a ...interface{}) Log {
	return Log{fmt.Sprintf(format, a...)}
}
