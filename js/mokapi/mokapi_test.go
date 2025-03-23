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
	"os"
	"testing"
	"time"
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
				os.Setenv("mokapi_foo_env_var", "bar")
				defer os.Unsetenv("mokapi_foo_env_var")

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
				r.EqualError(t, err, "encoding data to 'application/json' failed: error count 1:\n- #/type: invalid type, expected integer but got string at mokapi/js/mokapi.(*Module).Marshal-fm (native)")
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
