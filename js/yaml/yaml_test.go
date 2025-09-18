package yaml_test

import (
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/require"
	"mokapi/js/yaml"
	"testing"
)

func TestYaml(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "parse yaml",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				}}

				v, err := vm.RunString(`
					const m = require("mokapi/yaml")
					m.parse('foo: bar')
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{"foo": "bar"}, v.Export())
			},
		},
		{
			name: "stringify",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				}}

				v, err := vm.RunString(`
					const m = require("mokapi/yaml")
					m.stringify({ foo: "bar" })
				`)
				r.NoError(t, err)
				r.Equal(t, "foo: bar\n", v.Export())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			vm := goja.New()
			host := &enginetest.Host{}
			js.EnableInternal(vm, host, &eventloop.EventLoop{}, &dynamic.Config{})
			req, err := require.NewRegistry()
			r.NoError(t, err)
			req.Enable(vm)
			req.RegisterNativeModule("mokapi/yaml", yaml.Require)

			tc.test(t, vm, host)
		})
	}
}
