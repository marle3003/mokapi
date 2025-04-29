package console_test

import (
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/console"
	"mokapi/js/eventloop"
	"testing"
)

func TestConsole(t *testing.T) {
	testcases := []struct {
		name string
		host *enginetest.Host
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "string",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var logs []any
				host.InfoFunc = func(args ...interface{}) {
					logs = args
				}

				_, err := vm.RunString(`
					console.log('hello');
				`)
				r.NoError(t, err)
				r.Equal(t, "hello", logs[0])
			},
		},
		{
			name: "two string",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var logs []any
				host.InfoFunc = func(args ...interface{}) {
					logs = args
				}

				_, err := vm.RunString(`
					console.log('hello', 'world');
				`)
				r.NoError(t, err)
				r.Equal(t, "hello", logs[0])
				r.Equal(t, "world", logs[1])
			},
		},
		{
			name: "object",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var logs []any
				host.InfoFunc = func(args ...interface{}) {
					logs = args
				}

				_, err := vm.RunString(`
					console.log({ foo: 'bar' });
				`)
				r.NoError(t, err)
				r.Equal(t, `{"foo":"bar"}`, logs[0])
			},
		},
		{
			name: "format",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var logs []any
				host.InfoFunc = func(args ...interface{}) {
					logs = args
				}

				_, err := vm.RunString(`
					console.log('hello %s!', 'world');
				`)
				r.NoError(t, err)
				r.Len(t, logs, 1)
				r.Equal(t, "hello world!", logs[0])
			},
		},
		{
			name: "format with object",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var logs []any
				host.InfoFunc = func(args ...interface{}) {
					logs = args
				}

				_, err := vm.RunString(`
					console.log('hello %s!', { foo: 'bar' });
				`)
				r.NoError(t, err)
				r.Equal(t, `hello {"foo":"bar"}!`, logs[0])
			},
		},
		{
			name: "format with decimal",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var logs []any
				host.InfoFunc = func(args ...interface{}) {
					logs = args
				}

				_, err := vm.RunString(`
					console.log('hello %d!', 123);
				`)
				r.NoError(t, err)
				r.Equal(t, `hello 123!`, logs[0])
			},
		},
		{
			name: "invalid format",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var logs []any
				host.InfoFunc = func(args ...interface{}) {
					logs = args
				}

				_, err := vm.RunString(`
					console.log('hello %', 123);
				`)
				r.NoError(t, err)
				r.Equal(t, `hello %`, logs[0])
				r.Equal(t, int64(123), logs[1])
			},
		},
		{
			name: "format with missing",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var logs []any
				host.InfoFunc = func(args ...interface{}) {
					logs = args
				}

				_, err := vm.RunString(`
					console.log('hello %s');
				`)
				r.NoError(t, err)
				r.Equal(t, `hello %s`, logs[0])
			},
		},
		{
			name: "format %9.2f",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var logs []any
				host.InfoFunc = func(args ...interface{}) {
					logs = args
				}

				_, err := vm.RunString(`
					console.log('hello %9.2f', 3.141);
				`)
				r.NoError(t, err)
				r.Equal(t, `hello      3.14`, logs[0])
			},
		},
		{
			name: "format %9.2f with missing",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var logs []any
				host.InfoFunc = func(args ...interface{}) {
					logs = args
				}

				_, err := vm.RunString(`
					console.log('hello %9.2f');
				`)
				r.NoError(t, err)
				r.Equal(t, `hello %9.2f`, logs[0])
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := tc.host
			if host == nil {
				host = &enginetest.Host{}
			}
			js.EnableInternal(vm, host, &eventloop.EventLoop{}, &dynamic.Config{})
			console.Enable(vm)

			tc.test(t, vm, host)
		})
	}
}
