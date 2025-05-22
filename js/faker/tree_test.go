package faker_test

import (
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

func TestTree(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "findByName nil",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker');
					m.findByName('foo');
				`)
				r.NoError(t, err)
				r.Nil(t, v.Export())
			},
		},
		{
			name: "findByName root",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.FindFakerNodeFunc = func(name string) *common.FakerTree {
					if name == "root" {
						return common.NewFakerTree(generator.NewNode("root"))
					}
					return nil
				}

				v, err := vm.RunString(`
					const m = require('faker');
					m.findByName('root');
				`)
				r.NoError(t, err)
				r.NotNil(t, v.Export())
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
