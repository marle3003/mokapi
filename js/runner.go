package js

import (
	"github.com/dop251/goja"
	"time"
)

type timer struct {
	fn        func()
	timer     *time.Timer
	cancelled bool
}

type runner struct {
	vm *goja.Runtime

	exports goja.Value

	queueChan chan func()
	stopChan  chan struct{}
	running   bool
	jobCount  int
}

func newRunner(vm *goja.Runtime) *runner {
	r := &runner{
		vm:        vm,
		queueChan: make(chan func()),
		stopChan:  make(chan struct{}),
	}

	r.exports = vm.NewObject()
	_ = vm.Set("exports", r.exports)
	_ = vm.Set("setTimeout", r.setTimeout)

	return r
}

func (r *runner) Run(fn func(vm *goja.Runtime)) {
	if r.running {
		r.queueChan <- func() { fn(r.vm) }
	} else {
		fn(r.vm)
	}
}

func (r *runner) StartLoop() {
	r.running = true
	go func() {
	LOOP:
		for {
			select {
			case job := <-r.queueChan:
				job()
			case <-r.stopChan:
				break LOOP
			}
		}
	}()
}

func (r *runner) Stop() {
	if r.running {
		r.stopChan <- struct{}{}
	}
}

func (r *runner) HasJobs() bool {
	return r.jobCount > 0
}

func (r *runner) setTimeout(call goja.FunctionCall) goja.Value {
	r.jobCount++
	delay, f := r.getScheduledFunc(call)
	t := &timer{
		fn: f,
	}
	t.timer = time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
		t.cancelled = true
		r.jobCount--
		r.queueChan <- f
	})

	return r.vm.ToValue(t)
}

func (r *runner) getScheduledFunc(call goja.FunctionCall) (int64, func()) {
	if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
		delay := call.Argument(1).ToInteger()
		var args []goja.Value
		if len(call.Arguments) > 2 {
			args = append(args, call.Arguments[2:]...)
		}
		f := func() {
			_, err := fn(nil, args...)
			if err != nil {
				panic(r.vm.ToValue(err.Error()))
			}
		}
		return delay, f
	}
	return 0, nil
}
