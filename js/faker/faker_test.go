package faker_test

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/faker"
	"mokapi/js/require"
	"mokapi/schema/json/generator"
	"testing"
)

func TestModule(t *testing.T) {
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
			name: "fake invalid argument",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake('foo')
				`)
				r.EqualError(t, err, "expect object parameter but got: String at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "fake string",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({ type: 'string' })
				`)
				r.NoError(t, err)
				r.Equal(t, "XidZuoWq ", v.Export())
			},
		},
		{
			name: "fake with example",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({ type: 'object', example: { foo: 'bar' } })
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{"foo": "bar"}, v.Export())
			},
		},
		{
			name: "FakerTree: findByName with existing name",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					return generator.FindByName(name)
				}

				vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
				_, err := vm.RunString(`
					const m = require('faker')
					const n = m.findByName('product')
					if (!n) {
						throw new Error('not found')
					}
					if (n.name !== 'product') {
						throw new Error('name does not match: '+n.name())
					}
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "FakerTree: add string node",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					return generator.FindByName(name)
				}

				vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
				v, err := vm.RunString(`
					const m = require('faker')
					const root = m.findByName('root')
				  	root.children.unshift({
						name: 'foo',
						attributes: ['*'],
						fake: () => {
							return 'foobar'
						}
				  	})
				    m.fake({type: 'string'})
				`)
				r.NoError(t, err)
				r.Equal(t, "foobar", v.Export())
			},
		},
		{
			name: "FakerTree: dependsOn",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					return generator.FindByName(name)
				}

				vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
				v, err := vm.RunString(`
					const m = require('faker')
					const root = m.findByName('root')
				  	root.children.unshift(
						{
							name: 'foo',
							fake: (r) => {
								const value = 'hello';
								r.context.values['foo'] = value;
								return value;
							}
						},
						{
							name: 'bar',
							dependsOn: 'foo',
							fake: (r) => {
								return r.context.values['foo'] + ' Carol';
							}
						},
					)
				    m.fake({ properties: { foo: {}, bar: {} } })
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{"bar": "hello Carol", "foo": "hello"}, v.Export())
			},
		},
		{
			name: "FakerTree: append custom faker",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					return generator.FindByName(name)
				}

				vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
				v, err := vm.RunString(`
					const m = require('faker')
					const n = m.findByName('root')
					const frequencyItems = ['never', 'daily', 'weekly', 'monthly', 'yearly']
					n.children.push({
						name: 'frequency',
						fake: (r) => {
							return frequencyItems[Math.floor(Math.random()*frequencyItems.length)]
						}
					})
					m.fake({
						properties: {
							frequency: { }
						}
					})
				`)
				r.NoError(t, err)
				m := v.Export().(map[string]interface{})
				frequencyItems := []string{"never", "daily", "weekly", "monthly", "yearly"}
				r.Contains(t, frequencyItems, m["frequency"])
			},
		},
		{
			name: "fake with required property",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({ type: 'object', properties: { foo: { type: 'string' }, bar: { type: 'string' }}, required: ['foo', 'bar','x', 'y', 'z'] } )
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{"bar": "", "foo": "XidZuoWq "}, v.Export())
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
			defer cleanup(host)
			loop := eventloop.New(vm)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			tc.test(t, vm, host)
		})
	}
}
