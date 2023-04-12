package enginetest

import "mokapi/engine/common"

type engine struct {
	emit func(event string, args ...interface{}) []*common.Action
}

func NewEngine() common.EventEmitter {
	return &engine{}
}

func NewEngineWithHandler(handler func(event string, args ...interface{}) []*common.Action) common.EventEmitter {
	return &engine{emit: handler}
}

func (e *engine) Emit(event string, args ...interface{}) []*common.Action {
	if e.emit != nil {
		return e.emit(event, args...)
	}
	return nil
}
