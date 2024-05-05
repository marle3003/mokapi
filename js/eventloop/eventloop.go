package eventloop

import (
	"fmt"
	"github.com/dop251/goja"
	"sync"
	"time"
)

type timer struct {
	fn        func()
	timer     *time.Timer
	cancelled bool
}

type EventLoop struct {
	vm *goja.Runtime

	exports goja.Value

	queueChan chan func()
	stopChan  chan struct{}
	running   bool
	jobCount  int

	waitCond *sync.Cond
	waitLock sync.Mutex
}

func New(vm *goja.Runtime) *EventLoop {
	r := &EventLoop{
		vm:        vm,
		queueChan: make(chan func(), 1),
		stopChan:  make(chan struct{}, 1),
	}
	r.waitCond = sync.NewCond(&r.waitLock)

	r.exports = vm.NewObject()
	_ = vm.Set("exports", r.exports)
	_ = vm.Set("setTimeout", r.setTimeout)

	return r
}

func (loop *EventLoop) Run(fn func(vm *goja.Runtime)) {
	if loop.running {
		loop.queueChan <- func() { fn(loop.vm) }
	} else {
		fn(loop.vm)
	}
}

func (loop *EventLoop) RunAsync(fn func(vm *goja.Runtime) (goja.Value, error)) (goja.Value, error) {
	if loop.running {
		var result goja.Value
		var err error
		done := make(chan struct{})
		loop.queueChan <- func() {
			result, err = fn(loop.vm)
			done <- struct{}{}
		}

		<-done

		if err != nil {
			return nil, err
		}

		if p, ok := result.Export().(*goja.Promise); ok {
			for p.State() == goja.PromiseStatePending && loop.running {
				loop.wait()
			}
			return p.Result(), nil
		}

		return result, nil
	}

	return nil, fmt.Errorf("runner not started")
}

func (loop *EventLoop) StartLoop() {
	loop.running = true
	go func() {
	LOOP:
		for {
			select {
			case job := <-loop.queueChan:
				job()
				loop.wakeup()
			case <-loop.stopChan:
				loop.wakeup()
				break LOOP
			}
		}
	}()
}

func (loop *EventLoop) Stop() {
	if loop.running {
		loop.stopChan <- struct{}{}
	}
}

func (loop *EventLoop) HasJobs() bool {
	return loop.jobCount > 0
}

func (loop *EventLoop) setTimeout(call goja.FunctionCall) goja.Value {
	loop.jobCount++
	delay, f := loop.getScheduledFunc(call)
	t := &timer{
		fn: f,
	}
	t.timer = time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
		t.cancelled = true
		loop.jobCount--
		loop.queueChan <- f
	})

	return loop.vm.ToValue(t)
}

func (loop *EventLoop) getScheduledFunc(call goja.FunctionCall) (int64, func()) {
	if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
		delay := call.Argument(1).ToInteger()
		var args []goja.Value
		if len(call.Arguments) > 2 {
			args = append(args, call.Arguments[2:]...)
		}
		f := func() {
			_, err := fn(nil, args...)
			if err != nil {
				panic(loop.vm.ToValue(err.Error()))
			}
		}
		return delay, f
	}
	return 0, nil
}

func (loop *EventLoop) wait() {
	loop.waitLock.Lock()
	defer loop.waitLock.Unlock()
	loop.jobCount++
	loop.waitCond.Wait()
	loop.jobCount--
}

func (loop *EventLoop) wakeup() {
	loop.waitLock.Lock()
	defer loop.waitLock.Unlock()
	loop.waitCond.Broadcast()
}
