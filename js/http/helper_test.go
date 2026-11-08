package http_test

import (
	"mokapi/js/eventloop"
	"testing"
	"time"

	"github.com/dop251/goja"
)

// The goja runtime is not goroutine-safe: tests must run scripts and read values on the
// event loop instead of from the test goroutine.

// evalOnLoop runs the script on the event loop and returns the exported result.
func evalOnLoop(loop *eventloop.EventLoop, script string) (any, error) {
	var exported any
	_, err := loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
		v, err := vm.RunString(script)
		if err == nil && v != nil {
			exported = v.Export()
		}
		return v, err
	})
	return exported, err
}

// awaitResult waits until the script has set the global variable result, for example in a
// promise callback, and returns its exported value.
func awaitResult(t *testing.T, loop *eventloop.EventLoop) any {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		v, err := evalOnLoop(loop, "result")
		if err != nil {
			t.Fatal(err)
		}
		if v != nil {
			return v
		}
		if time.Now().After(deadline) {
			t.Fatal("result was not set")
		}
		time.Sleep(5 * time.Millisecond)
	}
}
