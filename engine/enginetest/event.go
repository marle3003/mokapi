package enginetest

import "mokapi/engine/common"

type eventEmitter struct {
	emit func(event string, args ...interface{}) []*common.Action
}

func NewEventEmitter() common.EventEmitter {
	return &eventEmitter{}
}

func NewEngineWithHandler(handler func(event string, args ...interface{}) []*common.Action) common.EventEmitter {
	return &eventEmitter{emit: handler}
}

func (e *eventEmitter) Emit(event string, args ...interface{}) []*common.Action {
	if e.emit != nil {
		return e.emit(event, args...)
	}
	return nil
}
