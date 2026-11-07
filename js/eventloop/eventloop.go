package eventloop

import (
	"errors"
	"fmt"
	"mokapi/engine/common"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dop251/goja"
)

type timeout struct {
	timer *time.Timer
}

type interval struct {
	run    func()
	ticker *time.Ticker
	stop   chan struct{}
}

type JobFunc func(vm *goja.Runtime) (goja.Value, error)

type JobContext struct {
	EventLogger func(level, message string)
}

// ErrStopped is returned by RunAsync when the loop is stopped before the job's result is available.
var ErrStopped = errors.New("event loop stopped")

// outcome is the result of a RunAsync job. It is written to the job's result channel exactly once.
type outcome struct {
	value goja.Value
	err   error
}

type EventLoop struct {
	vm   *goja.Runtime
	host common.Host

	exports goja.Value

	queueChan chan func()
	stopChan  chan struct{}
	// stopped is closed by the loop goroutine when it exits
	stopped  chan struct{}
	running  atomic.Bool
	jobCount atomic.Int32

	loopGoID uint64
}

func New(vm *goja.Runtime, host common.Host) *EventLoop {
	r := &EventLoop{
		vm:        vm,
		host:      host,
		queueChan: make(chan func(), 1),
		stopChan:  make(chan struct{}, 1),
		stopped:   make(chan struct{}),
	}

	r.exports = vm.NewObject()
	_ = vm.Set("exports", r.exports)
	_ = vm.Set("setTimeout", r.setTimeout)
	_ = vm.Set("clearTimeout", r.clearTimeout)
	_ = vm.Set("setInterval", r.setInterval)
	_ = vm.Set("clearInterval", r.clearInterval)

	return r
}

func goID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	field := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, _ := strconv.ParseUint(field, 10, 64)
	return id
}

func (loop *EventLoop) Run(fn func(vm *goja.Runtime)) {
	if loop.running.Load() {
		loop.queueChan <- func() { fn(loop.vm) }
	} else {
		fn(loop.vm)
	}
}

func (loop *EventLoop) RunSync(fn func(vm *goja.Runtime) (goja.Value, error)) (goja.Value, error) {
	if loop.onLoopGoroutine() {
		return fn(loop.vm)
	}

	done := make(chan struct{})
	var result goja.Value
	var err error
	loop.queueChan <- func() {
		result, err = fn(loop.vm)
		done <- struct{}{}
	}

	<-done
	return result, err
}

func (loop *EventLoop) RunAsync(fn JobFunc, ctx *JobContext) (goja.Value, error) {
	if loop.onLoopGoroutine() {
		return fn(loop.vm)
	}

	if !loop.running.Load() {
		return nil, fmt.Errorf("event loop not started")
	}

	select {
	case <-loop.stopped:
		return nil, ErrStopped
	default:
	}

	if ctx.EventLogger != nil {
		loop.host.SetEventLogger(ctx.EventLogger)
		defer loop.host.SetEventLogger(nil)
	}

	result := make(chan outcome, 1)
	select {
	case loop.queueChan <- loop.newJob(fn, result):
	case <-loop.stopped:
		return nil, ErrStopped
	}

	select {
	case o := <-result:
		return o.value, o.err
	case <-loop.stopped:
		return nil, ErrStopped
	}
}

// newJob wraps fn so that everything touching goja values, including the promise
// state, happens on the loop goroutine.
func (loop *EventLoop) newJob(fn JobFunc, result chan<- outcome) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				result <- outcome{err: err}
			}
		}()

		v, err := fn(loop.vm)
		if err != nil {
			result <- outcome{err: err}
			return
		}
		loop.deliver(v, result)
	}
}

func (loop *EventLoop) deliver(v goja.Value, result chan<- outcome) {
	if v == nil {
		v = goja.Undefined()
	}

	p, ok := v.Export().(*goja.Promise)
	switch {
	case !ok:
		result <- outcome{value: v}
	case p.State() == goja.PromiseStatePending:
		loop.deliverWhenSettled(v, p, result)
	default:
		result <- settledOutcome(p)
	}
}

// deliverWhenSettled registers Go reactions on the pending promise. goja offers no Go API
// for that, so it calls the promise's JS "then". jobCount stays above zero while the
// promise is pending, so HasJobs reports the loop as busy.
func (loop *EventLoop) deliverWhenSettled(v goja.Value, p *goja.Promise, result chan<- outcome) {
	loop.jobCount.Add(1)
	onSettled := loop.vm.ToValue(func(goja.FunctionCall) goja.Value {
		loop.jobCount.Add(-1)
		result <- settledOutcome(p)
		return goja.Undefined()
	})

	then, _ := goja.AssertFunction(v.ToObject(loop.vm).Get("then"))
	if _, err := then(v, onSettled, onSettled); err != nil {
		loop.jobCount.Add(-1)
		result <- outcome{err: err}
	}
}

func settledOutcome(p *goja.Promise) outcome {
	if p.State() == goja.PromiseStateRejected {
		return outcome{err: fmt.Errorf("%v", p.Result().ToString())}
	}
	return outcome{value: p.Result()}
}

func (loop *EventLoop) StartLoop() {
	loop.running.Store(true)
	go func() {
		defer close(loop.stopped)
		atomic.StoreUint64(&loop.loopGoID, goID())
	LOOP:
		for {
			select {
			case job := <-loop.queueChan:
				job()
			case <-loop.stopChan:
				break LOOP
			}
		}
	}()
}

func (loop *EventLoop) Stop() {
	if loop.running.Load() {
		select {
		case loop.stopChan <- struct{}{}:
		default:
		}
	}
}

func (loop *EventLoop) HasJobs() bool {
	return loop.jobCount.Load() > 0
}

func (loop *EventLoop) setTimeout(call goja.FunctionCall) goja.Value {
	loop.jobCount.Add(1)
	delay, run := loop.getScheduledFunc(call)
	t := &timeout{}
	t.timer = time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
		loop.jobCount.Add(-1)
		loop.queueChan <- run
	})

	return loop.vm.ToValue(t)
}

func (loop *EventLoop) clearTimeout(t *timeout) {
	t.timer.Stop()
	loop.jobCount.Add(-1)
}

func (loop *EventLoop) setInterval(call goja.FunctionCall) goja.Value {
	loop.jobCount.Add(1)
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
	loop.jobCount.Add(-1)
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

func (loop *EventLoop) onLoopGoroutine() bool {
	return goID() == atomic.LoadUint64(&loop.loopGoID)
}
