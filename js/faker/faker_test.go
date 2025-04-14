package faker_test

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/faker"
	"mokapi/js/require"
	"mokapi/schema/json/generator"
	"testing"
)

func TestModule(t *testing.T) {
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
				r.EqualError(t, err, "expect object parameter but got: String at mokapi/js/faker.(*Faker).Fake-fm (native)")
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
				host.FindFakerNodeFunc = func(name string) *common.FakerTree {
					return common.NewFakerTree(generator.FindByName(name))
				}

				vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
				_, err := vm.RunString(`
					const m = require('faker')
					const n = m.findByName('product')
					if (!n) {
						throw new Error('not found')
					}
					if (n.name() !== 'product') {
						throw new Error('name does not match: '+n.name())
					}
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "FakerTree: append custom faker",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.FindFakerNodeFunc = func(name string) *common.FakerTree {
					return common.NewFakerTree(generator.FindByName(name))
				}

				vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
				v, err := vm.RunString(`
					const m = require('faker')
					const n = m.findByName('')
					const frequencyItems = ['never', 'daily', 'weekly', 'monthly', 'yearly']
					n.append({
						name: 'frequency',
						fake: (r) => {
							return frequencyItems[Math.floor(Math.random()*frequencyItems.length)]
						}
					})
					m.fake({
						type: "object",
						properties: {
							frequency: { type: 'string' }
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
			loop := eventloop.New(vm)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			tc.test(t, vm, host)
		})
	}
}

type fakerTreeTest struct {
	name     string
	testFunc func(r *generator.Request) bool
	fakeFunc func(r *generator.Request) (interface{}, error)

	appendFunc   func(tree common.FakerNode)
	insertFunc   func(index int, tree common.FakerNode) error
	removeAtFunc func(index int) error
	removeFunc   func(name string) error
}

func (f *fakerTreeTest) Name() string {
	return f.name
}

func (f *fakerTreeTest) Test(r *generator.Request) bool {
	if f.testFunc != nil {
		return f.testFunc(r)
	}
	return false
}

func (f *fakerTreeTest) Fake(r *generator.Request) (interface{}, error) {
	if f.testFunc != nil {
		return f.fakeFunc(r)
	}
	return nil, fmt.Errorf("not implemented")
}

func (f *fakerTreeTest) Append(tree common.FakerNode) {
	if f.appendFunc != nil {
		f.appendFunc(tree)
	}
}

func (f *fakerTreeTest) Insert(index int, tree common.FakerNode) error {
	if f.insertFunc != nil {
		return f.insertFunc(index, tree)
	}
	return fmt.Errorf("not implemented")
}

func (f *fakerTreeTest) RemoveAt(index int) error {
	if f.removeAtFunc != nil {
		return f.removeAtFunc(index)
	}
	return fmt.Errorf("not implemented")
}

func (f *fakerTreeTest) Remove(name string) error {
	if f.removeFunc != nil {
		return f.removeFunc(name)
	}
	return fmt.Errorf("not implemented")
}
