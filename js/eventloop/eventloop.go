package eventloop

import (
	"fmt"
	"github.com/dop251/goja"
	"sync"
	"time"
)

type timeout struct {
	timer *time.Timer
}

type interval struct {
	run    func()
	ticker *time.Ticker
	stop   chan struct{}
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
	_ = vm.Set("clearTimeout", r.clearTimeout)
	_ = vm.Set("setInterval", r.setInterval)
	_ = vm.Set("clearInterval", r.clearInterval)

	return r
}

func (loop *EventLoop) Run(fn func(vm *goja.Runtime)) {
	if loop.running {
		loop.queueChan <- func() { fn(loop.vm) }
	} else {
		fn(loop.vm)
	}
}

func (loop *EventLoop) RunSync(fn func(vm *goja.Runtime)) {
	done := make(chan struct{})
	loop.queueChan <- func() {
		fn(loop.vm)
		done <- struct{}{}
	}

	<-done
}

func (loop *EventLoop) RunAsync(fn func(vm *goja.Runtime) (goja.Value, error)) (goja.Value, error) {
	if loop.running {
		var result goja.Value
		var err error
		done := make(chan struct{})
		loop.queueChan <- func() {
			defer func() {
				if r := recover(); r != nil {
					err = r.(error)
					done <- struct{}{}
				}
			}()

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

	return nil, fmt.Errorf("eventloop not started")
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
	delay, run := loop.getScheduledFunc(call)
	t := &timeout{}
	t.timer = time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
		loop.jobCount--
		loop.queueChan <- run
	})

	return loop.vm.ToValue(t)
}

func (loop *EventLoop) clearTimeout(t *timeout) {
	t.timer.Stop()
	loop.jobCount--
}

func (loop *EventLoop) setInterval(call goja.FunctionCall) goja.Value {
	loop.jobCount++
	v, run := loop.getScheduledFunc(call)
	milliseconds := time.Duration(v) * time.Millisecond
	// https://nodejs.org/api/timers.html#timers_setinterval_callback_delay_args
	if milliseconds <= 0 {
		milliseconds = time.Millisecond
	}

	i := &interval{
		run:  run,
		stop: make(chan struct{}),
	}
	i.ticker = time.NewTicker(milliseconds)
	go i.start(loop)

	return loop.vm.ToValue(i)
}

func (loop *EventLoop) clearInterval(i *interval) {
	loop.jobCount--
	close(i.stop)
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

func (i *interval) start(loop *EventLoop) {
Stop:
	for {
		select {
		case <-i.stop:
			i.ticker.Stop()
			break Stop
		case <-i.ticker.C:
			loop.queueChan <- i.run
		}
	}
}
