package lua

import "time"

type EventLog struct {
	Workflows []*WorkflowLog
}

type WorkflowLog struct {
	Name     string
	Log      []string
	Duration time.Duration
}
