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

func TestModule_Patch(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "null patch with null",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.patch(null)
				`)
				r.NoError(t, err)
				r.Equal(t, nil, v.Export())
			},
		},
		{
			name: "null patch with string",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.patch(null, 'foo')
				`)
				r.NoError(t, err)
				r.Equal(t, "foo", v.Export())
			},
		},
		{
			name: "string patch with string",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.patch('foo', 'bar')
				`)
				r.NoError(t, err)
				r.Equal(t, "bar", v.Export())
			},
		},
		{
			name: "string patch with int",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.patch('foo', 12)
				`)
				r.NoError(t, err)
				r.Equal(t, int64(12), v.Export())
			},
		},
		{
			name: "object patch with int",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.patch({ value: 'foo' }, 12)
				`)
				r.NoError(t, err)
				r.Equal(t, int64(12), v.Export())
			},
		},
		{
			name: "replace property value",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.patch({ value: 'foo' }, { value: 'bar' })
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]any{"value": "bar"}, v.Export())
			},
		},
		{
			name: "add property value2",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.patch({ value: 'foo' }, { value2: 'bar' })
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]any{"value": "foo", "value2": "bar"}, v.Export())
			},
		},
		{
			name: "no property defined",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.patch({ value: 'foo' }, { })
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]any{"value": "foo"}, v.Export())
			},
		},
		{
			name: "patch one property",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.patch({ value: 'foo', x: 1 }, { value: 'bar' })
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]any{"value": "bar", "x": int64(1)}, v.Export())
			},
		},
		{
			name: "nested property",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.patch({ nested: { value: 'foo' }}, { nested: { value: 'bar' }})
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]any{"nested": map[string]any{"value": "bar"}}, v.Export())
			},
		},
		{
			name: "remove property",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi')
					m.patch({ value: 'foo' }, { value: m.Delete })
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]any{}, v.Export())
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
