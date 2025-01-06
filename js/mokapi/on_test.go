package mokapi_test

import (
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/mokapi"
	"mokapi/js/require"
	"testing"
)

func TestModule_On(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "register event handler",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var event string
				var handler func(args ...interface{}) (bool, error)
				host.OnFunc = func(evt string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					event = evt
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', () => true)
				`)
				r.NoError(t, err)
				r.Equal(t, "http", event)
				b, err := handler()
				r.NoError(t, err)
				r.Equal(t, true, b)
			},
		},
		{
			name: "event handler with parameter",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var handler func(args ...interface{}) (bool, error)
				host.OnFunc = func(evt string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', (param) => param === 'foo')
				`)
				r.NoError(t, err)
				b, err := handler("foo")
				r.NoError(t, err)
				r.Equal(t, true, b)
			},
		},
		{
			name: "event handler throws error",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var handler func(args ...interface{}) (bool, error)
				host.OnFunc = func(evt string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', () => { throw new Error('TEST') })
				`)
				r.NoError(t, err)
				_, err = handler()
				r.EqualError(t, err, "Error: TEST at <eval>:3:33(3)")
			},
		},
		{
			name: "event handler with tags",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var tags map[string]string
				host.OnFunc = func(evt string, do func(args ...interface{}) (bool, error), t map[string]string) {
					tags = t
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', () => true, { tags: { foo: 'bar', bar: null } })
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]string{"foo": "bar", "bar": "null"}, tags)
			},
		},
		{
			name: "event handler with tags but invalid type",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', () => true, { tags: 'foo' })
				`)
				r.EqualError(t, err, "unexpected type for tags: String at mokapi/js/mokapi.(*Module).On-fm (native)")
			},
		},
		{
			name: "event handler invalid type for args",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', () => true, 'foo')
				`)
				r.EqualError(t, err, "unexpected type for args: String at mokapi/js/mokapi.(*Module).On-fm (native)")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg, err := require.NewRegistry()
			reg.RegisterNativeModule("mokapi", mokapi.Require)
			r.NoError(t, err)

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			loop := eventloop.New(vm)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			tc.test(t, vm, host)
		})
	}
}
