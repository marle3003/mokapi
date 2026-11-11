package mokapi_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/jstest"
	"mokapi/js/mokapi"
	"mokapi/js/require"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

func TestModule(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "sleep 1s",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				now := time.Now()
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.sleep('1s')
				`)
				r.NoError(t, err)
				r.Greater(t, time.Now().Sub(now), 1*time.Second)
			},
		},
		{
			name: "sleep invalid duration",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.sleep('1')
				`)
				r.EqualError(t, err, "time: missing unit in duration \"1\" at mokapi/js/mokapi.(*Module).Sleep-fm (native)")
			},
		},
		{
			name: "sleep 500 (ms)",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				now := time.Now()
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.sleep(500)
				`)
				r.NoError(t, err)
				r.Greater(t, time.Now().Sub(now), 500*time.Millisecond)
			},
		},
		{
			name: "sleep invalid argument",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.sleep(1.5)
				`)
				r.EqualError(t, err, "unexpected type for time: 1.5 at mokapi/js/mokapi.(*Module).Sleep-fm (native)")
			},
		},
		{
			name: "env",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_ = os.Setenv("mokapi_foo_env_var", "bar")
				defer func() { _ = os.Unsetenv("mokapi_foo_env_var") }()

				v, err := vm.RunString(`
					const m = require('mokapi')
					m.env('mokapi_foo_env_var')
				`)
				r.NoError(t, err)
				r.Equal(t, "bar", v.Export())
			},
		},
		{
			name: "not set env",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.env('mokapi_foo_env_var_not_set')
				`)
				r.NoError(t, err)
				r.Equal(t, "", v.Export())
			},
		},
		{
			name: "date",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.date({timestamp:  new Date(Date.UTC(2022, 5, 9, 12, 0, 0, 0)).getTime()}); // january is 0
				`)
				r.NoError(t, err)
				expected := time.Date(2022, 6, 9, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
				r.Equal(t, expected, v.Export())
			},
		},
		{
			name: "marshal text/plain",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.marshal('foo', { schema: {type: 'string'}, contentType: 'text/plain' })
				`)
				r.NoError(t, err)
				r.Equal(t, "foo", v.Export())
			},
		},
		{
			name: "marshal json",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.marshal('foo', { schema: {type: 'string'}, contentType: 'application/json' })
				`)
				r.NoError(t, err)
				r.Equal(t, `"foo"`, v.Export())
			},
		},
		{
			name: "marshal uses json as default",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.marshal('foo', { schema: {type: 'string'} })
				`)
				r.NoError(t, err)
				r.Equal(t, `"foo"`, v.Export())
			},
		},
		{
			name: "marshal with error",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.marshal('foo', { schema: {type: 'integer'}, contentType: 'application/json' })
				`)
				r.EqualError(t, err, "Validation error count 1:\n\t- #/type: invalid type, expected integer but got string at mokapi/js/mokapi.(*Module).Marshal-fm (native)")
			},
		},
		{
			name: "invalid schema",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.marshal('foo', { schema: {type: {}}, contentType: 'application/json' })
				`)
				r.EqualError(t, err, "unexpected type for 'type': Object at mokapi/js/mokapi.(*Module).Marshal-fm (native)")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg, err := require.NewRegistry(nil)
			reg.RegisterNativeModule("mokapi", mokapi.Require)
			r.NoError(t, err)

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			loop := eventloop.New(vm, host)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			tc.test(t, vm, host)
		})
	}
}

func TestDelay(t *testing.T) {
	delayed := []struct {
		name string
		arg  string
		min  time.Duration
	}{
		{name: "milliseconds", arg: `100`, min: 100 * time.Millisecond},
		{name: "duration string", arg: `'50ms'`, min: 50 * time.Millisecond},
		{name: "float is truncated to whole milliseconds", arg: `1.7`},
		{name: "zero resolves right away", arg: `0`},
		{name: "negative resolves right away", arg: `-5`},
	}
	for _, tc := range delayed {
		t.Run("delay "+tc.name, func(t *testing.T) {
			loop, _ := newModuleLoop(t)
			_, err := evalOnLoop(loop, `const m = require('mokapi')`)
			r.NoError(t, err)

			start := time.Now()
			_, err = evalOnLoop(loop, `const p = m.delay(`+tc.arg+`)`)
			r.NoError(t, err)
			state, value := awaitPromiseResult(t, loop, "p")

			r.Equal(t, goja.PromiseStateFulfilled, state)
			r.Nil(t, value)
			r.GreaterOrEqual(t, time.Since(start), tc.min)
		})
	}

	t.Run("delay is pending until the time has passed", func(t *testing.T) {
		loop, _ := newModuleLoop(t)
		_, err := evalOnLoop(loop, `const m = require('mokapi'); const p = m.delay(300)`)
		r.NoError(t, err)

		state := pendingState(t, loop, "p")

		r.Equal(t, goja.PromiseStatePending, state)
	})

	t.Run("loop runs other jobs while a delay is pending", func(t *testing.T) {
		loop, _ := newModuleLoop(t)
		_, err := evalOnLoop(loop, `const m = require('mokapi'); const p = m.delay('300ms')`)
		r.NoError(t, err)
		start := time.Now()

		v, err := evalOnLoop(loop, `1 + 1`)

		r.NoError(t, err)
		r.Equal(t, int64(2), v)
		r.Less(t, time.Since(start), 200*time.Millisecond)
		r.Equal(t, goja.PromiseStatePending, pendingState(t, loop, "p"))
		r.Equal(t, goja.PromiseStateFulfilled, awaitPromise(t, loop, "p"))
	})

	t.Run("invalid arguments throw synchronously", func(t *testing.T) {
		errorCases := []struct {
			name     string
			arg      string
			expected string
		}{
			{name: "duration without unit", arg: `'1'`, expected: `time: missing unit in duration "1"`},
			{name: "unexpected type", arg: `{}`, expected: `unexpected type for time: map[]`},
		}
		for _, tc := range errorCases {
			t.Run(tc.name, func(t *testing.T) {
				loop, _ := newModuleLoop(t)
				_, err := evalOnLoop(loop, `const m = require('mokapi')`)
				r.NoError(t, err)

				_, err = evalOnLoop(loop, `m.delay(`+tc.arg+`)`)

				r.Error(t, err)
				// goja appends the name of the native function, which differs from sleep
				message, _, _ := strings.Cut(err.Error(), " at ")
				r.Equal(t, tc.expected, message)
			})
		}
	})

	t.Run("pending delay leaves no goroutine after Stop", func(t *testing.T) {
		before := runtime.NumGoroutine()
		loop, _ := newModuleLoop(t)
		_, err := evalOnLoop(loop, `const m = require('mokapi'); m.delay('50ms')`)
		r.NoError(t, err)

		loop.Stop()
		// the timer fires after the loop is gone
		time.Sleep(100 * time.Millisecond)

		jstest.RequireGoroutinesAtMost(t, before)
	})
}

// pendingState returns the state of the promise stored in the global variable name without waiting.
func pendingState(t *testing.T, loop *eventloop.EventLoop, name string) goja.PromiseState {
	t.Helper()
	var state goja.PromiseState
	_, err := loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
		state = vm.Get(name).Export().(*goja.Promise).State()
		return nil, nil
	})
	r.NoError(t, err)
	return state
}
