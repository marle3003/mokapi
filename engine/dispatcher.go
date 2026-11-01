package engine

import (
	"cmp"
	"mokapi/engine/common"
	"slices"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type eventHandler[F any] struct {
	Filter  F
	Execute common.EventHandlerFunc
	Args    common.EventArgs
}

type dispatcher[F any] struct {
	handlers map[string][]*eventHandler[F]
	mu       sync.RWMutex
}

func (d *dispatcher[F]) Register(key string, h *eventHandler[F]) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.handlers == nil {
		d.handlers = make(map[string][]*eventHandler[F])
	}
	d.handlers[key] = append(d.handlers[key], h)
}

func (d *dispatcher[F]) Clear(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.handlers, key)
}

func (d *dispatcher[F]) Has(key string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, ok := d.handlers[key]
	return ok
}

type emitOptions[F any] struct {
	shouldEmit func(F) bool
	stop       func() bool
}

func (d *dispatcher[F]) emit(opts emitOptions[F], params ...any) []*common.Action {
	d.mu.RLock()
	var hs []*eventHandler[F]
	for _, h := range d.handlers {
		hs = append(hs, h...)
	}
	d.mu.RUnlock()

	slices.SortStableFunc(hs, func(a, b *eventHandler[F]) int { return -1 * cmp.Compare(a.Args.Priority, b.Args.Priority) })

	var result []*common.Action
	for _, h := range hs {
		if opts.shouldEmit != nil && !opts.shouldEmit(h.Filter) {
			continue
		}
		a := runEventHandler(h.Execute, h.Args, params...)
		if a != nil {
			result = append(result, a)
		}
		if opts.stop != nil && opts.stop() {
			break
		}
	}

	return result
}

func runEventHandler(f common.EventHandlerFunc, args common.EventArgs, params ...any) *common.Action {
	if f == nil {
		return nil
	}

	action := &common.Action{
		Tags: args.Tags,
	}
	start := time.Now()
	logs := len(action.Logs)

	ctx := &common.EventContext{
		EventLogger: action.AppendLog,
		Args:        params,
	}

	if b, err := f(ctx); err != nil {
		log.Errorf("unable to execute event handler: %v", err)
		action.Error = &common.Error{Message: err.Error()}
	} else if !b && logs == len(action.Logs) {
		return nil
	}
	log.WithField("handler", action).Debug("processed event handler")

	action.Parameters = getDeepCopy(params)
	action.Duration = time.Now().Sub(start).Milliseconds()
	return action
}

func newEventHandler[F any](sh *scriptHost, event string, filter F, f common.EventHandlerFunc, args common.EventArgs) *eventHandler[F] {
	addDefaultTags(&args, sh, event)
	return &eventHandler[F]{Filter: filter, Execute: f, Args: args}
}
