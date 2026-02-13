package engine

import (
	"cmp"
	"mokapi/engine/common"
	"slices"
	"time"

	log "github.com/sirupsen/logrus"
)

func (e *Engine) Run(event string, args ...interface{}) []*common.Action {
	var ehs []*eventHandler
	for _, h := range e.scripts {
		ehs = append(ehs, h.events[event]...)
	}
	slices.SortStableFunc(ehs, func(a, b *eventHandler) int { return -1 * cmp.Compare(a.priority, b.priority) })

	var result []*common.Action

	for _, eh := range ehs {
		a := runEventHandler(eh, args...)
		if a != nil {
			result = append(result, a)
		}
	}

	return result
}

func runEventHandler(eh *eventHandler, args ...interface{}) *common.Action {
	action := &common.Action{
		Tags: eh.tags,
	}
	start := time.Now()
	logs := len(action.Logs)

	ctx := &common.EventContext{
		EventLogger: action.AppendLog,
		Args:        args,
	}

	if b, err := eh.handler(ctx); err != nil {
		log.Errorf("unable to execute event handler: %v", err)
		action.Error = &common.Error{Message: err.Error()}
	} else if !b && logs == len(action.Logs) {
		return nil
	}
	log.WithField("handler", action).Debug("processed event handler")

	action.Parameters = getDeepCopy(args)
	action.Duration = time.Now().Sub(start).Milliseconds()
	return action
}
