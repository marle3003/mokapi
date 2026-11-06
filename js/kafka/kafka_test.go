package kafka_test

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/kafka"
	"mokapi/js/require"
	"testing"
	"time"

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
			name: "produce client error throws",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return nil, fmt.Errorf("TEST ERROR")
				}}

				_, err := vm.RunString(`
					const kafka = require("mokapi/kafka")
					kafka.produce({})
				`)
				r.EqualError(t, err, "TEST ERROR at mokapi/js/kafka.(*Module).Produce-fm (native)")
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

		{
			name: "auto create topic",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, true, args.AllowAutoTopicCreation)
					return &common.KafkaProduceResult{}, nil
				}}

				_, err := vm.RunString(`
					const kafka = require("mokapi/kafka")
					kafka.produce({ allowAutoTopicCreation: true })
				`)
				r.NoError(t, err)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			vm := goja.New()
			host := &enginetest.Host{}
			js.EnableInternal(vm, host, &eventloop.EventLoop{}, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			req, err := require.NewRegistry(nil)
			r.NoError(t, err)
			req.Enable(vm)
			req.RegisterNativeModule("mokapi/kafka", kafka.Require)

			tc.test(t, vm, host)
		})
	}
}

func TestKafka_ProduceAsync(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		client func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error)
		state  string // "fulfilled" or "rejected"
		result any    // exported value or rejection reason of the promise
		thrown string // part of the exception when the call itself must throw
	}{
		{
			name:   "resolves with the client result",
			script: `kafka.produceAsync({ topic: 'foo', messages: [{ value: 'hello' }] })`,
			client: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
				return &common.KafkaProduceResult{}, nil
			},
			state:  "fulfilled",
			result: &common.KafkaProduceResult{},
		},
		{
			name:   "invalid argument throws synchronously",
			script: `kafka.produceAsync({ messages: [ [] ] })`,
			thrown: "invalid type in messages: expected Object but got Array",
		},
		{
			name:   "invalid retry throws synchronously",
			script: `kafka.produceAsync({ retry: { maxRetryTime: 'foo' } })`,
			thrown: "parse maxRetryTime failed",
		},
		{
			name:   "client error rejects the promise",
			script: `kafka.produceAsync({ topic: 'foo', messages: [{ value: 'hello' }] })`,
			client: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
				return nil, fmt.Errorf("TEST ERROR")
			},
			state:  "rejected",
			result: "TEST ERROR",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			vm := goja.New()
			host := &enginetest.Host{}
			loop := eventloop.New(vm, host)
			loop.StartLoop()
			defer loop.Stop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			req, err := require.NewRegistry(nil)
			r.NoError(t, err)
			req.Enable(vm)
			req.RegisterNativeModule("mokapi/kafka", kafka.Require)
			host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: tc.client}

			type outcome struct {
				state string
				value any
			}
			settled := make(chan outcome, 1)
			_, err = loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
				_ = vm.Set("settled", func(state string, v any) { settled <- outcome{state, v} })
				return vm.RunString(`const kafka = require("mokapi/kafka");` + tc.script +
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
