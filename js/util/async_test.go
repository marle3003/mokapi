package util_test

import (
	"fmt"
	"mokapi/engine/enginetest"
	"mokapi/js/eventloop"
	"mokapi/js/util"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

type outcome struct {
	state string
	value any
}

// runAsync starts the promise returned by start on the event loop, like a JavaScript API
// method would, and waits until it is settled.
func runAsync(t *testing.T, start func(vm *goja.Runtime, loop *eventloop.EventLoop) *goja.Promise) outcome {
	t.Helper()
	vm := goja.New()
	loop := eventloop.New(vm, &enginetest.Host{})
	loop.StartLoop()
	defer loop.Stop()

	settled := make(chan outcome, 1)
	_, err := loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
		_ = vm.Set("settled", func(state string, value any) { settled <- outcome{state, value} })
		_ = vm.Set("p", start(vm, loop))
		return vm.RunString(`p.then(v => settled('fulfilled', v), e => settled('rejected', e))`)
	})
	require.NoError(t, err)

	select {
	case o := <-settled:
		return o
	case <-time.After(5 * time.Second):
		t.Fatal("promise was not settled")
		return outcome{}
	}
}

func goroutineID() string {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	return strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
}

func TestRunAsync(t *testing.T) {
	testcases := []struct {
		name     string
		work     func() (string, error)
		settle   func(string) (any, error)
		expected outcome
	}{
		{
			name:     "resolves with the work result without settle",
			work:     func() (string, error) { return "foo", nil },
			expected: outcome{"fulfilled", "foo"},
		},
		{
			name:     "resolves with the value of settle",
			work:     func() (string, error) { return "foo", nil },
			settle:   func(result string) (any, error) { return result + "bar", nil },
			expected: outcome{"fulfilled", "foobar"},
		},
		{
			name:     "rejects with the error message of work",
			work:     func() (string, error) { return "", fmt.Errorf("TEST ERROR") },
			expected: outcome{"rejected", "TEST ERROR"},
		},
		{
			name:     "rejects when work panics",
			work:     func() (string, error) { panic("TEST PANIC") },
			expected: outcome{"rejected", "TEST PANIC"},
		},
		{
			name:     "does not call settle when work fails",
			work:     func() (string, error) { return "", fmt.Errorf("TEST ERROR") },
			settle:   func(string) (any, error) { panic("settle must not run") },
			expected: outcome{"rejected", "TEST ERROR"},
		},
		{
			name:     "rejects with the error message of settle",
			work:     func() (string, error) { return "foo", nil },
			settle:   func(string) (any, error) { return nil, fmt.Errorf("TEST SETTLE ERROR") },
			expected: outcome{"rejected", "TEST SETTLE ERROR"},
		},
		{
			name:     "rejects when settle panics",
			work:     func() (string, error) { return "foo", nil },
			settle:   func(string) (any, error) { panic("TEST SETTLE PANIC") },
			expected: outcome{"rejected", "TEST SETTLE PANIC"},
		},
		{
			name: "rejects when settle dereferences a nil result",
			work: func() (string, error) { return "foo", nil },
			settle: func(string) (any, error) {
				var m map[string]string
				m["foo"] = "bar"
				return nil, nil
			},
			expected: outcome{"rejected", "assignment to entry in nil map"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			o := runAsync(t, func(vm *goja.Runtime, loop *eventloop.EventLoop) *goja.Promise {
				return util.RunAsync(vm, loop, tc.work, tc.settle)
			})
			require.Equal(t, tc.expected, o)
		})
	}
}

func TestRunAsync_SettleRunsOnLoop(t *testing.T) {
	var loopID, workID, settleID string

	o := runAsync(t, func(vm *goja.Runtime, loop *eventloop.EventLoop) *goja.Promise {
		loopID = goroutineID()
		return util.RunAsync(vm, loop,
			func() (string, error) {
				workID = goroutineID()
				return "foo", nil
			},
			func(result string) (any, error) {
				settleID = goroutineID()
				return result, nil
			})
	})

	require.Equal(t, "fulfilled", o.state)
	require.Equal(t, loopID, settleID, "settle must run on the event loop")
	require.NotEqual(t, loopID, workID, "work must not block the event loop")
}

func TestRunAsync_EventLoopStaysResponsive(t *testing.T) {
	vm := goja.New()
	loop := eventloop.New(vm, &enginetest.Host{})
	loop.StartLoop()
	defer loop.Stop()

	release := make(chan struct{})
	_, err := loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
		util.RunAsync(vm, loop, func() (string, error) {
			<-release
			return "foo", nil
		}, nil)
		return nil, nil
	})
	require.NoError(t, err)

	// the loop runs other scripts while the work is still in flight
	v, err := loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
		return vm.RunString(`1 + 1`)
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), v.Export())
	close(release)
}

func TestThrowPanic(t *testing.T) {
	vm := goja.New()
	call := func(f func()) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r.(goja.Value))
			}
		}()
		f()
		return nil
	}

	t.Run("converts a panic to a value the script can catch", func(t *testing.T) {
		err := call(func() {
			defer util.ThrowPanic(vm)
			panic("TEST PANIC")
		})
		require.EqualError(t, err, "TEST PANIC")
	})

	t.Run("passes through when nothing panics", func(t *testing.T) {
		err := call(func() {
			defer util.ThrowPanic(vm)
		})
		require.NoError(t, err)
	})
}
