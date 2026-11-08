package util

import (
	"fmt"
	"mokapi/js/eventloop"

	"github.com/dop251/goja"
)

// RunAsync returns a promise that is settled with the outcome of work.
//
// work runs in its own goroutine and must not touch the goja runtime. settle (optional)
// runs on the event loop with the result of work and produces the value the promise is
// resolved with, e.g. to call a JavaScript `result` callback. Without settle the promise
// is resolved with the result of work. An error or panic in work or settle rejects the
// promise with the error message.
//
// The promise is created on the calling goroutine, which must be the event loop.
func RunAsync[T any](vm *goja.Runtime, loop *eventloop.EventLoop,
	work func() (T, error), settle func(T) (any, error)) *goja.Promise {
	p, resolve, reject := vm.NewPromise()
	go func() {
		result, err := recoverToError(work)
		loop.Run(func(vm *goja.Runtime) {
			var value any = result
			if err == nil && settle != nil {
				value, err = recoverToError(func() (any, error) { return settle(result) })
			}
			if err != nil {
				_ = reject(vm.ToValue(err.Error()))
				return
			}
			_ = resolve(value)
		})
	}()
	return p
}

func recoverToError[T any](f func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	return f()
}

// ThrowPanic turns a Go panic into a JavaScript exception, so that e.g. a failed type
// assertion on script input does not crash the event loop. It has to be deferred directly:
//
//	defer util.ThrowPanic(vm)
func ThrowPanic(vm *goja.Runtime) {
	if r := recover(); r != nil {
		panic(vm.ToValue(fmt.Sprintf("%v", r)))
	}
}
