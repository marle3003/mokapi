package file_test

import (
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/file"
	"mokapi/js/require"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"testing"
)

func TestModule_Open(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "open",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.OpenFunc = func(file, hint string) (*dynamic.Config, error) {
					return &dynamic.Config{Raw: []byte(file)}, nil
				}

				v, err := vm.RunString(`
					open('foo.txt')
				`)
				r.NoError(t, err)
				r.Equal(t, "foo.txt", v.Export())
			},
		},
		{
			name: "resolve",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.OpenFunc = func(file, hint string) (*dynamic.Config, error) {
					return &dynamic.Config{
						Data: openapitest.NewConfig("2.0.0",
							openapitest.WithComponentSchema("Foo", schematest.New("string")),
						),
					}, nil
				}

				v, err := vm.RunString(`
					const api = open('foo.yaml', { as: 'resolved' })
					api.components.schemas.Foo
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]any{"type": "string"}, v.Export())
			},
		},
		{
			name: "resolve with circular reference",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.OpenFunc = func(file, hint string) (*dynamic.Config, error) {
					s := &schema.Schema{Properties: &schema.Schemas{}}
					s.Properties.Set("foo", s)

					return &dynamic.Config{
						Data: openapitest.NewConfig("2.0.0",
							openapitest.WithComponentSchema("Foo", s),
						),
					}, nil
				}

				v, err := vm.RunString(`
					const api = open('foo.yaml', { as: 'resolved' })
					api.components.schemas.Foo.properties.foo
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]any{"description": "circular reference"}, v.Export())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg, err := require.NewRegistry()
			r.NoError(t, err)

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			loop := eventloop.New(vm)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)
			file.Enable(vm, host)

			tc.test(t, vm, host)
		})
	}
}
