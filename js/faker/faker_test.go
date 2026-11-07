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
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
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
				r.Equal(t, "fnsy", v.Export())
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
				r.Equal(t, map[string]any{"foo": "bar"}, v.Export())
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
			name: "FakerTree: request path of the root value is an empty array",
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
						fake: (r) => "path length " + r.path.length
					})
					m.fake({type: 'string'})
				`)
				r.NoError(t, err)
				r.Equal(t, "path length 0", v.Export())
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
							attributes: ['foo'],
							fake: (r) => {
								const value = 'hello';
								r.context.values['foo'] = value;
								return value;
							}
						},
						{
							name: 'bar',
							attributes: ['bar'],
							dependsOn: ['foo'],
							fake: (r) => {
								return r.context.values['foo'] + ' Carol';
							}
						},
					)
				    m.fake({ properties: { bar: {}, foo: {} }, required: ['foo', 'bar'] })
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]any{"bar": "hello Carol", "foo": "hello"}, v.Export())
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
						attributes: ['frequency'],
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
				m := v.Export().(map[string]any)
				frequencyItems := []string{"never", "daily", "weekly", "monthly", "yearly"}
				r.Contains(t, frequencyItems, m["frequency"])
			},
		},
		{
			name: "fake with required property",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({ type: 'object', properties: { foo: { type: 'string' }, bar: { type: 'string' }}, additionalProperties: { type: 'string' }, required: ['foo', 'bar','x', 'y', 'z'] } )
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]any{"bar": "aKAQyByPS<qbft", "class": "Clq kaieGDf", "fear": "itucts2mX", "foo": "yxlsTI", "heart": "qdeZwkx5", "pack": "DnQ", "trip": "VgPAl89Xbz v", "x": "IqiIQzyMKld", "y": "nI fpidfoDeHcd", "z": ""}, v.Export())
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_ = gofakeit.Seed(11)

			reg, err := require.NewRegistry(nil)
			reg.RegisterNativeModule("faker", faker.Require)
			r.NoError(t, err)

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			defer cleanup(host)
			loop := eventloop.New(vm, host)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			tc.test(t, vm, host)
		})
	}
}

func TestModule_FakeAsync(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		state  string // "fulfilled" or "rejected"
		result any    // exported value of the promise
		thrown string // part of the exception when the call itself must throw
	}{
		{
			name:   "resolves with the generated value",
			script: `m.fakeAsync({ type: 'object', example: { foo: 'bar' } })`,
			state:  "fulfilled",
			result: map[string]any{"foo": "bar"},
		},
		{
			name:   "invalid argument throws synchronously",
			script: `m.fakeAsync('foo')`,
			thrown: "expect object parameter but got: String",
		},
		{
			name:   "without argument resolves with undefined",
			script: `m.fakeAsync()`,
			state:  "fulfilled",
			result: nil,
		},
		{
			name:   "invalid schema throws synchronously",
			script: `m.fakeAsync({ type: 'object', properties: 12 })`,
			thrown: "expected object, got number",
		},
		{
			name: "uses the schema as it was at call time",
			script: `(() => {
				const schema = { type: 'object', example: { foo: 'bar' } }
				const p = m.fakeAsync(schema)
				schema.example = { foo: 'changed' }
				return p
			})()`,
			state:  "fulfilled",
			result: map[string]any{"foo": "bar"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_ = gofakeit.Seed(11)

			reg, err := require.NewRegistry(nil)
			reg.RegisterNativeModule("faker", faker.Require)
			r.NoError(t, err)

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			loop := eventloop.New(vm, host)
			loop.StartLoop()
			defer loop.Stop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			type outcome struct {
				state string
				value any
			}
			settled := make(chan outcome, 1)
			_, err = loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
				_ = vm.Set("settled", func(state string, v any) { settled <- outcome{state, v} })
				return vm.RunString(`const m = require('faker');` + tc.script +
					`.then(v => settled('fulfilled', v), e => settled('rejected', e))`)
			})
			if tc.thrown != "" {
				r.ErrorContains(t, err, tc.thrown)
				return
			}
			r.NoError(t, err)

			select {
			case o := <-settled:
				r.Equal(t, outcome{tc.state, tc.result}, o)
			case <-time.After(5 * time.Second):
				t.Fatal("promise was not settled")
			}
		})
	}
}
