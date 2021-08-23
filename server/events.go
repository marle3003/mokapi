package server

type EventLog struct {
	Workflows []*WorkflowLog
}

type WorkflowLog struct {
	Name string
	Log  []string
}
