package file_test

import (
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
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
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
						Data: openapitest.NewConfig("3.0.0",
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
		{
			name: "open",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				dir := t.TempDir()
				err := os.WriteFile(filepath.Join(dir, "foo.txt"), []byte("Hello World"), 0o644)
				r.NoError(t, err)

				host.CwdFunc = func() string {
					return dir
				}

				v, err := vm.RunString(`
					const m = require("mokapi/file")
					m.read('foo.txt');
				`)
				r.NoError(t, err)
				r.Equal(t, "Hello World", v.Export(), dir)
			},
		},
		{
			name: "write file",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				dir := t.TempDir()
				host.CwdFunc = func() string {
					return dir
				}

				_, err := vm.RunString(`
					const m = require("mokapi/file")
					m.writeString('foo.txt', 'Hello World');
				`)
				r.NoError(t, err)

				b, err := os.ReadFile(filepath.Join(dir, "foo.txt"))
				r.NoError(t, err)
				r.Equal(t, "Hello World", string(b), dir)
			},
		},
		{
			name: "append string to file",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				dir := t.TempDir()
				err := os.WriteFile(filepath.Join(dir, "foo.txt"), []byte("Hello World"), 0o644)
				r.NoError(t, err)

				host.CwdFunc = func() string {
					return dir
				}

				_, err = vm.RunString(`
					const m = require("mokapi/file")
					m.appendString('foo.txt', '!');
				`)
				r.NoError(t, err)

				b, err := os.ReadFile(filepath.Join(dir, "foo.txt"))
				r.NoError(t, err)
				r.Equal(t, "Hello World!", string(b), dir)
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
			loop := eventloop.New(vm, host)
			defer loop.Stop()
			loop.StartLoop()
			source := &dynamic.Config{Info: dynamictest.NewConfigInfo()}
			js.EnableInternal(vm, host, loop, source)
			reg.Enable(vm)
			file.Enable(vm, host, source)
			reg.RegisterNativeModule("mokapi/file", file.Require)

			tc.test(t, vm, host)
		})
	}
}
