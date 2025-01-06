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

func TestModule_Cron(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "register cron handler",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var cron string
				var handler func()
				host.CronFunc = func(every string, do func(), opt common.JobOptions) {
					cron = every
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					let test = ''
					m.cron('* * * * *', () => { test = 'foo' })
				`)
				r.NoError(t, err)
				r.Equal(t, "* * * * *", cron)
				handler()
				v, err := vm.RunString("test")
				r.NoError(t, err)
				r.Equal(t, "foo", v.Export())
			},
		},
		{
			name: "cron handler throws error",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var handler func()
				host.CronFunc = func(every string, do func(), opt common.JobOptions) {
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.cron('', () => { throw new Error('TEST') })
				`)
				r.NoError(t, err)
				r.Panics(t, handler)
			},
		},
		{
			name: "cron handler with tags",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var tags map[string]string
				host.CronFunc = func(evt string, do func(), opts common.JobOptions) {
					tags = opts.Tags
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.cron('', () => {}, { tags: { foo: 'bar', bar: null } })
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]string{"foo": "bar", "bar": "null"}, tags)
			},
		},
		{
			name: "cron handler with tags but invalid type",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.cron('', () => {}, { tags: 'foo' })
				`)
				r.EqualError(t, err, "unexpected type for tags: String at mokapi/js/mokapi.(*Module).Cron-fm (native)")
			},
		},
		{
			name: "cron handler invalid type for args",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.cron('', () => {}, 'foo')
				`)
				r.EqualError(t, err, "unexpected type for args: String at mokapi/js/mokapi.(*Module).Cron-fm (native)")
			},
		},
		{
			name: "cron handler with times options",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var times int
				host.CronFunc = func(evt string, do func(), opts common.JobOptions) {
					times = opts.Times
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.cron('', () => {}, { times: 10 })
				`)
				r.NoError(t, err)
				r.Equal(t, 10, times)
			},
		},
		{
			name: "cron handler with invalid type for times",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.cron('', () => {}, { times: [] })
				`)
				r.EqualError(t, err, "unexpected type for times: Array at mokapi/js/mokapi.(*Module).Cron-fm (native)")
			},
		},
		{
			name: "cron handler with skipImmediateFirstRun",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var skip bool
				host.CronFunc = func(evt string, do func(), opts common.JobOptions) {
					skip = opts.SkipImmediateFirstRun
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.cron('', () => {}, { skipImmediateFirstRun: true })
				`)
				r.NoError(t, err)
				r.Equal(t, true, skip)
			},
		},
		{
			name: "cron handler with invalid type for skipImmediateFirstRun",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.cron('', () => {}, { skipImmediateFirstRun: 'true' })
				`)
				r.EqualError(t, err, "unexpected type for skipImmediateFirstRun: String at mokapi/js/mokapi.(*Module).Cron-fm (native)")
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
