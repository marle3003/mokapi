package mokapi_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/mokapi"
	"mokapi/js/require"
	"testing"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

func TestModule_Webhook(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "webhook missing name",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.webhook()
				`)
				r.ErrorContains(t, err, "webhook name must not be empty")
			},
		},
		{
			name: "webhook missing url",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.webhook('foo')
				`)
				r.ErrorContains(t, err, "webhook url must not be empty")
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
			loop := eventloop.New(vm, host)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			tc.test(t, vm, host)
		})
	}
}
