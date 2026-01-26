package kafka_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/kafka"
	"mokapi/js/require"
	"testing"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

func TestKafka(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "produce no parameter",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, "mokapi-script", args.ClientId)
					r.Equal(t, "64613435-3062-6462-3033-316532633233", args.ScriptFile)
					return &common.KafkaProduceResult{}, nil
				}}

				_, err := vm.RunString(`
					const kafka = require("mokapi/kafka")
					kafka.produce()
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "produce empty args",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				}}

				_, err := vm.RunString(`
					const kafka = require("mokapi/kafka")
					kafka.produce({})
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "messages contains wrong type",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				}}

				_, err := vm.RunString(`
					const kafka = require("mokapi/kafka")
					kafka.produce({ messages: [ [] ] })
				`)
				r.EqualError(t, err, "invalid type in messages: expected Object but got Array at mokapi/js/kafka.(*Module).Produce-fm (native)")
			},
		},

		{
			name: "panic should not crash",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					panic("TEST")
				}}

				_, err := vm.RunString(`
					const kafka = require("mokapi/kafka")
					kafka.produce({})
				`)
				r.EqualError(t, err, "TEST at mokapi/js/kafka.(*Module).Produce-fm (native)")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			vm := goja.New()
			host := &enginetest.Host{}
			js.EnableInternal(vm, host, &eventloop.EventLoop{}, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			req, err := require.NewRegistry()
			r.NoError(t, err)
			req.Enable(vm)
			req.RegisterNativeModule("mokapi/kafka", kafka.Require)

			tc.test(t, vm, host)
		})
	}
}
