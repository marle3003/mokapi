package enginetest

import "mokapi/engine/common"

type engine struct {
	emit func(event string, args ...interface{})
}

func NewEngine() common.EventEmitter {
	return &engine{}
}

func NewEngineWithHandler(handler func(event string, args ...interface{})) common.EventEmitter {
	return &engine{emit: handler}
}

func (e *engine) Emit(event string, args ...interface{}) {
	if e.emit != nil {
		e.emit(event, args...)
	}
}
