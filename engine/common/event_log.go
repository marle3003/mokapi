package common

import "sync"

// EventLog holds the fields of an event log that are written after the event has been
// pushed to the event store: the duration of the request and the actions of the event
// handlers. Readers such as the dashboard API may already access the log at that time,
// so all access goes through the mutex.
type EventLog struct {
	m        sync.Mutex
	duration int64
	actions  []*Action
}

func (l *EventLog) SetActions(actions []*Action) {
	l.m.Lock()
	defer l.m.Unlock()
	l.actions = actions
}

func (l *EventLog) GetActions() []*Action {
	l.m.Lock()
	defer l.m.Unlock()
	return l.actions
}

func (l *EventLog) SetDuration(duration int64) {
	l.m.Lock()
	defer l.m.Unlock()
	l.duration = duration
}

func (l *EventLog) GetDuration() int64 {
	l.m.Lock()
	defer l.m.Unlock()
	return l.duration
}
