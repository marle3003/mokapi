package js_test

import (
	"mokapi/js/jstest"
	"testing"
	"time"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

// The goja runtime is not goroutine-safe: tests must read script values on the event loop
// instead of from the test goroutine.

// onLoop runs fn on the script's event loop and waits for its result.
func onLoop[T any](t *testing.T, s *jstest.Script, fn func(vm *goja.Runtime) T) T {
	t.Helper()
	result := make(chan T, 1)
	r.NoError(t, s.RunFunc(func(vm *goja.Runtime) { result <- fn(vm) }))
	select {
	case v := <-result:
		return v
	case <-time.After(5 * time.Second):
		t.Fatal("event loop did not run the function")
		panic("unreachable")
	}
}

// awaitGlobal waits until the script has set the global variable name (for example from
// a timer or a promise callback) and returns its exported value.
func awaitGlobal(t *testing.T, s *jstest.Script, name string) any {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		v := onLoop(t, s, func(vm *goja.Runtime) any {
			if value := vm.Get(name); value != nil && !goja.IsUndefined(value) {
				return value.Export()
			}
			return nil
		})
		if v != nil {
			return v
		}
		if time.Now().After(deadline) {
			t.Fatalf("global %q was not set", name)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
