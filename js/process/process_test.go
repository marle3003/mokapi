package process_test

import (
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/process"
	"testing"
)

func TestProcess(t *testing.T) {
	testcases := []struct {
		name string
		env  map[string]string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "env exists",
			env: map[string]string{
				"foo": "bar",
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				}}

				v, err := vm.RunString(`
					process.env.foo
				`)
				r.NoError(t, err)
				r.Equal(t, "bar", v.Export())
			},
		},
		{
			name: "env does not exist",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				}}

				v, err := vm.RunString(`
					process.env.foo
				`)
				r.NoError(t, err)
				r.Equal(t, nil, v.Export())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			vm := goja.New()
			host := &enginetest.Host{}
			js.EnableInternal(vm, host, &eventloop.EventLoop{}, &dynamic.Config{})

			for name, value := range tc.env {
				t.Setenv(name, value)
			}
			process.Enable(vm)

			tc.test(t, vm, host)
		})
	}
}
