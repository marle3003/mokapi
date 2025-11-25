package mokapi_test

import (
	"mokapi/engine/enginetest"
	"mokapi/js/eventloop"
	"mokapi/js/mokapi"
	"strings"
	"testing"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

func TestProxy(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime)
	}{
		{
			name: "set map value",
			test: func(t *testing.T, vm *goja.Runtime) {
				p := mokapi.NewProxy(map[string]any{}, vm)
				err := vm.Set("proxy", vm.NewDynamicObject(p))
				r.NoError(t, err)
				_, err = vm.RunString(`
					proxy.foo = 2;
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{"foo": int64(2)}, p.Export())
			},
		},
		{
			name: "get map value",
			test: func(t *testing.T, vm *goja.Runtime) {
				p := mokapi.NewProxy(map[string]any{"foo": 2}, vm)
				err := vm.Set("proxy", vm.NewDynamicObject(p))
				r.NoError(t, err)
				v, err := vm.RunString(`
					proxy.foo;
				`)
				r.NoError(t, err)
				r.Equal(t, 2, mokapi.Export(v))
			},
		},
		{
			name: "set map value with custom toJSValue",
			test: func(t *testing.T, vm *goja.Runtime) {
				p := mokapi.NewProxy(map[string]any{"foo": map[string]any{}}, vm)
				p.ToJSValue = func(vm *goja.Runtime, k string, v any) goja.Value {
					if k == "foo" {
						foo := mokapi.NewProxy(v, vm)
						foo.KeyNormalizer = func(s string) string {
							return strings.ToUpper(s)
						}
						return vm.NewDynamicObject(foo)
					}
					return vm.ToValue(v)
				}
				err := vm.Set("proxy", vm.NewDynamicObject(p))
				r.NoError(t, err)
				_, err = vm.RunString(`
					proxy.foo.bar = 1;
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{"foo": map[string]interface{}{"BAR": int64(1)}}, p.Export())
			},
		},
		{
			name: "set struct value",
			test: func(t *testing.T, vm *goja.Runtime) {
				type T struct {
					Foo int `json:"foo"`
				}
				s := &T{Foo: 1}

				p := mokapi.NewProxy(s, vm)
				err := vm.Set("proxy", vm.NewDynamicObject(p))
				r.NoError(t, err)
				_, err = vm.RunString(`
					proxy.foo = 2;
				`)
				r.NoError(t, err)
				r.Equal(t, T{Foo: 2}, p.Export())
			},
		},
		{
			name: "get struct value",
			test: func(t *testing.T, vm *goja.Runtime) {
				type T struct {
					Foo int `json:"foo"`
				}
				s := &T{Foo: 2}

				p := mokapi.NewProxy(s, vm)
				err := vm.Set("proxy", vm.NewDynamicObject(p))
				r.NoError(t, err)
				v, err := vm.RunString(`
					proxy.foo;
				`)
				r.NoError(t, err)
				r.Equal(t, 2, mokapi.Export(v))
			},
		},
		{
			name: "set map value with custom toJSValue",
			test: func(t *testing.T, vm *goja.Runtime) {
				type T struct {
					Foo map[string]any `json:"foo"`
				}
				s := &T{Foo: map[string]any{}}

				p := mokapi.NewProxy(s, vm)
				p.ToJSValue = func(vm *goja.Runtime, k string, v any) goja.Value {
					if k == "foo" {
						foo := mokapi.NewProxy(v, vm)
						foo.KeyNormalizer = func(s string) string {
							return strings.ToUpper(s)
						}
						return vm.NewDynamicObject(foo)
					}
					return vm.ToValue(v)
				}
				err := vm.Set("proxy", vm.NewDynamicObject(p))
				r.NoError(t, err)
				_, err = vm.RunString(`
					proxy.foo.bar = 1;
				`)
				r.NoError(t, err)
				i := int64(1)
				r.Equal(t, T{Foo: map[string]any{"BAR": &i}}, p.Export())
			},
		},
		{
			name: "array length",
			test: func(t *testing.T, vm *goja.Runtime) {
				p := mokapi.NewProxy([]any{1}, vm)
				err := vm.Set("proxy", vm.NewDynamicObject(p))
				r.NoError(t, err)
				v, err := vm.RunString(`
					proxy.length;
				`)
				r.NoError(t, err)
				r.Equal(t, int64(1), mokapi.Export(v))
			},
		},
		{
			name: "array pop",
			test: func(t *testing.T, vm *goja.Runtime) {
				p := mokapi.NewProxy([]any{1, 2, 3}, vm)
				err := vm.Set("proxy", vm.NewDynamicObject(p))
				r.NoError(t, err)
				v, err := vm.RunString(`
					proxy.pop();
				`)
				r.NoError(t, err)
				r.Equal(t, []any{1, 2}, mokapi.Export(p))
				r.Equal(t, 3, mokapi.Export(v))
			},
		},
		{
			name: "array shift",
			test: func(t *testing.T, vm *goja.Runtime) {
				p := mokapi.NewProxy([]any{1, 2, 3}, vm)
				err := vm.Set("proxy", vm.NewDynamicObject(p))
				r.NoError(t, err)
				v, err := vm.RunString(`
					proxy.shift();
				`)
				r.NoError(t, err)
				r.Equal(t, []any{2, 3}, mokapi.Export(p))
				r.Equal(t, 1, mokapi.Export(v))
			},
		},
		{
			name: "array unshift",
			test: func(t *testing.T, vm *goja.Runtime) {
				p := mokapi.NewProxy([]any{1, 2, 3}, vm)
				err := vm.Set("proxy", vm.NewDynamicObject(p))
				r.NoError(t, err)
				_, err = vm.RunString(`
					proxy.unshift(0);
				`)
				r.NoError(t, err)
				r.Equal(t, []any{int64(0), 1, 2, 3}, mokapi.Export(p))
			},
		},
		{
			name: "array unshift",
			test: func(t *testing.T, vm *goja.Runtime) {
				p := mokapi.NewProxy([]any{1, 3}, vm)
				err := vm.Set("proxy", vm.NewDynamicObject(p))
				r.NoError(t, err)
				_, err = vm.RunString(`
					proxy.splice(1, 0, 2);
				`)
				r.NoError(t, err)
				r.Equal(t, []any{1, int64(2), 3}, mokapi.Export(p))
			},
		},
		{
			name: "access array index",
			test: func(t *testing.T, vm *goja.Runtime) {
				p := mokapi.NewProxy([]any{1, 2, 3}, vm)
				err := vm.Set("proxy", vm.NewDynamicObject(p))
				r.NoError(t, err)
				v, err := vm.RunString(`
					proxy[1]
				`)
				r.NoError(t, err)
				r.Equal(t, int64(2), mokapi.Export(v))
			},
		},
		{
			name: "assign []string to []any",
			test: func(t *testing.T, vm *goja.Runtime) {
				p := mokapi.NewProxy(map[string][]any{"foo": {}}, vm)
				err := vm.Set("proxy", vm.NewDynamicObject(p))
				r.NoError(t, err)
				err = vm.Set("foo", vm.ToValue([]string{"bar", "yuh"}))
				r.NoError(t, err)
				_, err = vm.RunString(`
					proxy.foo = foo
				`)
				r.NoError(t, err)
				r.Equal(t, map[string][]any{"foo": {"bar", "yuh"}}, mokapi.Export(p))
			},
		},
		{
			name: "assign map[string][]string to map[string][]any",
			test: func(t *testing.T, vm *goja.Runtime) {
				p := mokapi.NewProxy(map[string]map[string][]any{"headers": {}}, vm)
				err := vm.Set("proxy", vm.NewDynamicObject(p))
				r.NoError(t, err)
				err = vm.Set("headers", vm.ToValue(map[string][]string{"foo": {"bar", "yuh"}}))
				r.NoError(t, err)
				_, err = vm.RunString(`
					proxy.headers = headers
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]map[string][]any{"headers": {"foo": {"bar", "yuh"}}}, mokapi.Export(p))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			loop := eventloop.New(vm, host)
			defer loop.Stop()
			loop.StartLoop()

			tc.test(t, vm)
		})
	}
}
