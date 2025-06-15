package mokapi_test

import (
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/mokapi"
	"mokapi/js/require"
	"testing"
)

func TestModule_Every(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "register every handler",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var every string
				var handler func()
				host.EveryFunc = func(e string, do func(), opt common.JobOptions) {
					every = e
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					let test = ''
					m.every('5s', () => { test = 'foo' })
				`)
				r.NoError(t, err)
				r.Equal(t, "5s", every)
				handler()
				v, err := vm.RunString("test")
				r.NoError(t, err)
				r.Equal(t, "foo", v.Export())
			},
		},
		{
			name: "every handler throws error",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var handler func()
				host.EveryFunc = func(every string, do func(), opt common.JobOptions) {
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.every('', () => { throw new Error('TEST') })
				`)
				r.NoError(t, err)
				r.Panics(t, handler)
			},
		},
		{
			name: "every handler with tags",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var tags map[string]string
				host.EveryFunc = func(evt string, do func(), opts common.JobOptions) {
					tags = opts.Tags
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.every('', () => {}, { tags: { foo: 'bar', bar: null } })
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]string{"foo": "bar", "bar": "null"}, tags)
			},
		},
		{
			name: "every handler with tags but invalid type",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.every('', () => {}, { tags: 'foo' })
				`)
				r.EqualError(t, err, "unexpected type for tags: String at mokapi/js/mokapi.(*Module).Every-fm (native)")
			},
		},
		{
			name: "every handler invalid type for args",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.every('', () => {}, 'foo')
				`)
				r.EqualError(t, err, "unexpected type for args: String at mokapi/js/mokapi.(*Module).Every-fm (native)")
			},
		},
		{
			name: "every handler with times options",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var times int
				host.EveryFunc = func(evt string, do func(), opts common.JobOptions) {
					times = opts.Times
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.every('', () => {}, { times: 10 })
				`)
				r.NoError(t, err)
				r.Equal(t, 10, times)
			},
		},
		{
			name: "every handler with invalid type for times",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.every('', () => {}, { times: [] })
				`)
				r.EqualError(t, err, "unexpected type for times: Array at mokapi/js/mokapi.(*Module).Every-fm (native)")
			},
		},
		{
			name: "every handler with skipImmediateFirstRun",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var skip bool
				host.EveryFunc = func(evt string, do func(), opts common.JobOptions) {
					skip = opts.SkipImmediateFirstRun
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.every('', () => {}, { skipImmediateFirstRun: true })
				`)
				r.NoError(t, err)
				r.Equal(t, true, skip)
			},
		},
		{
			name: "every handler with invalid type for skipImmediateFirstRun",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.every('', () => {}, { skipImmediateFirstRun: 'true' })
				`)
				r.EqualError(t, err, "unexpected type for skipImmediateFirstRun: String at mokapi/js/mokapi.(*Module).Every-fm (native)")
			},
		},
		{
			name: "async event handler",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var f func()
				host.EveryFunc = func(every string, do func(), opt common.JobOptions) {
					f = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					let result;
					m.every('1m', async () => {
						result = await getMessage();
					})

					let getMessage = async () => {
						return new Promise(async (resolve, reject) => {
						  setTimeout(() => {
							resolve('foo');
						  }, 200);
						});
					}
				`)
				r.NoError(t, err)
				f()
				r.NoError(t, err)
				v, err := vm.RunString("result")
				r.NoError(t, err)
				r.Equal(t, "foo", v.Export())
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
