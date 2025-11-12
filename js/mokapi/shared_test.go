package mokapi_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/mokapi"
	"mokapi/js/require"
	"testing"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

func TestModule_Shared(t *testing.T) {
	vmFactory := func(reg *require.Registry, store *engine.Store) func() *goja.Runtime {
		return func() *goja.Runtime {
			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{StoreTest: store}
			loop := eventloop.New(vm, host)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)
			return vm
		}
	}

	testcases := []struct {
		name string
		test func(t *testing.T, newVm func() *goja.Runtime)
	}{
		{
			name: "get",
			test: func(t *testing.T, newVm func() *goja.Runtime) {
				vm1 := newVm()
				vm2 := newVm()

				_, err := vm1.RunString(`
					const m = require('mokapi');
					m.shared.set('test', 'hello world');
				`)
				r.NoError(t, err)

				v, err := vm2.RunString(`
					const m = require('mokapi');
					m.shared.get('test');
				`)
				r.NoError(t, err)

				r.Equal(t, "hello world", mokapi.Export(v))
			},
		},
		{
			name: "has",
			test: func(t *testing.T, newVm func() *goja.Runtime) {
				vm1 := newVm()
				vm2 := newVm()

				_, err := vm1.RunString(`
					const m = require('mokapi');
					m.shared.set('test', 'hello world');
				`)
				r.NoError(t, err)

				v, err := vm2.RunString(`
					const m = require('mokapi');
					const result = {
						test: m.shared.has('test'),
						bar: m.shared.has('bar')
					};
					result
				`)
				r.NoError(t, err)

				r.Equal(t, map[string]any{"bar": false, "test": true}, mokapi.Export(v))
			},
		},
		{
			name: "clear",
			test: func(t *testing.T, newVm func() *goja.Runtime) {
				vm1 := newVm()
				vm2 := newVm()

				_, err := vm1.RunString(`
					const m = require('mokapi');
					m.shared.set('test', 'hello world');
				`)
				r.NoError(t, err)

				_, err = vm2.RunString(`
					const m = require('mokapi');
					m.shared.clear()
				`)
				r.NoError(t, err)

				v, err := vm1.RunString(`
					m.shared.get('bar')
				`)
				r.NoError(t, err)
				r.Equal(t, nil, mokapi.Export(v))

				_, err = vm1.RunString(`
					m.shared.set('bar', 123)
				`)
				r.NoError(t, err)

				v, err = vm2.RunString(`
					m.shared.get('bar')
				`)
				r.NoError(t, err)
				r.Equal(t, int64(123), mokapi.Export(v))
			},
		},
		{
			name: "update",
			test: func(t *testing.T, newVm func() *goja.Runtime) {
				vm1 := newVm()
				vm2 := newVm()

				_, err := vm1.RunString(`
					const m = require('mokapi');
					m.shared.update('counter', c => (c ?? 0) + 1);
				`)
				r.NoError(t, err)

				v, err := vm2.RunString(`
					const m = require('mokapi');
					m.shared.update('counter', c => (c ?? 0) + 1);
				`)
				r.NoError(t, err)

				r.Equal(t, int64(2), mokapi.Export(v))
			},
		},
		{
			name: "keys",
			test: func(t *testing.T, newVm func() *goja.Runtime) {
				vm1 := newVm()
				vm2 := newVm()

				_, err := vm1.RunString(`
					const m = require('mokapi');
					m.shared.set('foo', 123);
					m.shared.set('1', '-');
				`)
				r.NoError(t, err)

				v, err := vm2.RunString(`
					const m = require('mokapi');
					m.shared.set('bar', undefined);
					m.shared.set('100', '-');
					m.shared.keys();
				`)
				r.NoError(t, err)

				r.Equal(t, []string{"1", "100", "bar", "foo"}, mokapi.Export(v))
			},
		},
		{
			name: "namespace",
			test: func(t *testing.T, newVm func() *goja.Runtime) {
				vm1 := newVm()
				vm2 := newVm()

				_, err := vm1.RunString(`
					const m = require('mokapi');
					const petstore = m.shared.namespace('petstore');
					petstore.set('foo', 123);
				`)
				r.NoError(t, err)

				v, err := vm2.RunString(`
					const m = require('mokapi');
					const petstore = m.shared.namespace('petstore');
					petstore.get('foo');
				`)
				r.NoError(t, err)

				r.Equal(t, int64(123), mokapi.Export(v))
			},
		},
		{
			name: "update with objects",
			test: func(t *testing.T, newVm func() *goja.Runtime) {
				vm1 := newVm()
				vm2 := newVm()

				_, err := vm1.RunString(`
					const m = require('mokapi');
					const foo = m.shared.update('foo', (v) => v ?? {});
					foo.bar = '123'
				`)
				r.NoError(t, err)

				v, err := vm2.RunString(`
					const m = require('mokapi');
					const foo = m.shared.update('foo', (v) => v ?? {});
					foo;
				`)
				r.NoError(t, err)

				r.Equal(t, map[string]any{"bar": "123"}, mokapi.Export(v))
			},
		},
		{
			name: "update with array",
			test: func(t *testing.T, newVm func() *goja.Runtime) {
				vm1 := newVm()

				v, err := vm1.RunString(`
					const m = require('mokapi');
					const foo = m.shared.update('foo', (v) => v ?? { items: [] });
					foo.items.push(123)
					foo
				`)
				r.NoError(t, err)
				m := map[string]interface{}{}
				err = vm1.ExportTo(v, &m)
				r.Equal(t, map[string]any{"items": []any{int64(123)}}, mokapi.Export(v))
			},
		},
		{
			name: "enumerate object",
			test: func(t *testing.T, newVm func() *goja.Runtime) {
				vm1 := newVm()

				v, err := vm1.RunString(`
					const m = require('mokapi');
					const foo = m.shared.update('foo', (v) => v ?? { foo: 'bar' });
					const result = []
					for (let k in foo) {
						result.push(k)
					}
					result
				`)
				r.NoError(t, err)
				r.Equal(t, []any{"foo"}, mokapi.Export(v))
			},
		},
		{
			name: "spread object",
			test: func(t *testing.T, newVm func() *goja.Runtime) {
				vm1 := newVm()

				v, err := vm1.RunString(`
					const m = require('mokapi');
					const shared = m.shared.update('foo', (v) => v ?? { foo: 'bar' });
					const { foo } = shared
					foo
				`)
				r.NoError(t, err)
				r.Equal(t, "bar", mokapi.Export(v))
			},
		},
		{
			name: "splice array",
			test: func(t *testing.T, newVm func() *goja.Runtime) {
				vm1 := newVm()

				v, err := vm1.RunString(`
					const m = require('mokapi');
					const shared = m.shared.update('foo', (v) => v ?? { items: [1,2,3] });
					shared.items.splice(1, 1)
					shared.items
				`)
				r.NoError(t, err)
				r.Equal(t, []any{int64(1), int64(3)}, mokapi.Export(v))
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

			tc.test(t, vmFactory(reg, engine.NewStore()))
		})
	}
}
