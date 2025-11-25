package faker_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/faker"
	"mokapi/js/require"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

func TestNode(t *testing.T) {
	cleanup := func(host *enginetest.Host) {
		for index := len(host.CleanupFuncs) - 1; index >= 0; index-- {
			host.CleanupFuncs[index]()
		}
	}

	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "overwrite fake function",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				street := generator.NewNode("street")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "street" {
						return street
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const street = m.findByName('street');
					street.fake = () => 123;
				`)
				r.NoError(t, err)
				v, err := street.Fake(nil)
				r.NoError(t, err)
				r.Equal(t, int64(123), v)
				cleanup(host)
				r.Nil(t, street.Fake)
			},
		},
		{
			name: "call fake function",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				street := generator.NewNode("street")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "street" {
						return street
					}
					return nil
				}

				v, err := vm.RunString(`
					const m = require('faker');
					const street = m.findByName('street');
					street.fake = () => 123;
					street.fake()
				`)
				r.NoError(t, err)
				r.Equal(t, int64(123), v.Export())
			},
		},
		{
			name: "call existing fake function",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				street := generator.NewNode("street")
				street.Fake = func(r *generator.Request) (any, error) {
					return 123, nil
				}
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "street" {
						return street
					}
					return nil
				}

				v, err := vm.RunString(`
					const m = require('faker');
					const street = m.findByName('street');
					street.fake()
				`)
				r.NoError(t, err)
				r.Equal(t, int64(123), v.Export())
			},
		},
		{
			name: "call fake function with error",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				street := generator.NewNode("street")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "street" {
						return street
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const street = m.findByName('street');
					street.fake = () => { throw new Error('TEST') };
				`)
				r.NoError(t, err)
				_, err = street.Fake(nil)
				r.EqualError(t, err, "Error: TEST at <eval>:4:34(3)")
			},
		},
		{
			name: "call fake function access request parameter",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				street := generator.NewNode("street")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "street" {
						return street
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const street = m.findByName('street');
					street.fake = (r) => { return r.schema.type };
				`)
				r.NoError(t, err)
				v, err := street.Fake(generator.NewRequest(nil, schematest.New("string"), nil))
				r.NoError(t, err)
				r.Equal(t, &schema.Types{"string"}, v)
			},
		},
		{
			name: "change name",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				street := generator.NewNode("street")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "street" {
						return street
					}
					return nil
				}

				v, err := vm.RunString(`
					const m = require('faker');
					const street = m.findByName('street');
					street.name = 'foo';
					street.name;
				`)
				r.NoError(t, err)
				r.Equal(t, "foo", v.Export())
			},
		},
		{
			name: "read-write attributes",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				street := generator.NewNode("street")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "street" {
						return street
					}
					return nil
				}

				v, err := vm.RunString(`
					const m = require('faker');
					const street = m.findByName('street');
					street.attributes = ['foo', 'bar'];
					street.attributes;
				`)
				r.NoError(t, err)
				var attributes []string
				err = vm.ExportTo(v, &attributes)
				r.NoError(t, err)
				r.Equal(t, []string{"foo", "bar"}, attributes)
			},
		},
		{
			name: "read-write weight",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				street := generator.NewNode("street")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "street" {
						return street
					}
					return nil
				}

				v, err := vm.RunString(`
					const m = require('faker');
					const street = m.findByName('street');
					street.weight = 10;
					street.weight;
				`)
				r.NoError(t, err)
				r.Equal(t, int64(10), v.Export())
			},
		},
		{
			name: "read-write dependsOn",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				street := generator.NewNode("street")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "street" {
						return street
					}
					return nil
				}

				v, err := vm.RunString(`
					const m = require('faker');
					const street = m.findByName('street');
					street.dependsOn = ['foo', 'bar'];
					street.dependsOn;
				`)
				r.NoError(t, err)
				var dependsOn []string
				err = vm.ExportTo(v, &dependsOn)
				r.NoError(t, err)
				r.Equal(t, []string{"foo", "bar"}, dependsOn)
			},
		},
		{
			name: "read-write children",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				street := generator.NewNode("street")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "street" {
						return street
					}
					return nil
				}

				v, err := vm.RunString(`
					const m = require('faker');
					const street = m.findByName('street');
					street.children = [{name: 'foo'}, {name: 'bar'}];
					street.children;
				`)
				r.NoError(t, err)
				var children []*generator.Node
				err = vm.ExportTo(v, &children)
				r.NoError(t, err)
				r.Equal(t, "foo", children[0].Name)
				r.True(t, children[0].Custom)
				r.Equal(t, "bar", children[1].Name)
				r.True(t, children[1].Custom)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)

			reg, err := require.NewRegistry()
			reg.RegisterNativeModule("faker", faker.Require)
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
