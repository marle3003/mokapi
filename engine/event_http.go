package engine

import (
	"mokapi/engine/common"
)

type HttpEventHandler = eventHandler[common.HttpFilter]

type HttpEventDispatcher struct {
	dispatcher[common.HttpFilter]
}

func (sh *scriptHost) OnHttp(filter common.HttpFilter, do common.EventHandlerFunc, args common.EventArgs) {
	sh.engine.HttpEventDispatcher.Register(sh.name, newEventHandler[common.HttpFilter](sh, "http", filter, do, args))
}

func (e *HttpEventDispatcher) EmitHttp(req *common.HttpEventRequest, res *common.HttpEventResponse) []*common.Action {
	return e.emit(emitOptions[common.HttpFilter]{
		shouldEmit: func(f common.HttpFilter) bool {
			return matchesIfPresent(f.Method, req.Method) &&
				matchesIfPresent(f.OperationId, req.OperationId) &&
				matchesIfPresent(f.Path, req.Key) &&
				matchesIfPresent(f.Api, req.Api)
		},
		stop: func() bool { return res.StopPropagation },
	}, req, res)
}
