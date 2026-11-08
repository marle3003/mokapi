package mqtt_test

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/mqtt"
	"mokapi/js/require"
	"testing"
	"time"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

func TestMqtt(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "publish no parameter",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.MqttClientTest = &enginetest.MqttClient{PublishFunc: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
					r.Equal(t, "mokapi-script", args.ClientId)
					r.Equal(t, "64613435-3062-6462-3033-316532633233", args.ScriptFile)
					return &common.MqttPublishResult{}, nil
				}}

				_, err := vm.RunString(`
					const mqtt = require("mokapi/mqtt")
					mqtt.publish()
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "publish empty args",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.MqttClientTest = &enginetest.MqttClient{PublishFunc: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
					return &common.MqttPublishResult{}, nil
				}}

				_, err := vm.RunString(`
					const mqtt = require("mokapi/mqtt")
					mqtt.publish({})
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "publish maps all parameters",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var published *common.MqttPublishArgs
				host.MqttClientTest = &enginetest.MqttClient{PublishFunc: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
					published = args
					return &common.MqttPublishResult{}, nil
				}}

				_, err := vm.RunString(`
					const mqtt = require("mokapi/mqtt")
					mqtt.publish({ cluster: 'c', topic: 't', value: 'v', retain: true,
						retry: { maxRetryTime: '1s', initialRetryTime: 2, retries: 3, factor: 4 } })
				`)
				r.NoError(t, err)
				r.Equal(t, "c", published.Cluster)
				r.Equal(t, "t", published.Topic)
				r.Equal(t, []byte("v"), published.Value)
				r.True(t, published.Retain)
				r.Equal(t, common.RetryArgs{
					MaxRetryTime:     time.Second,
					InitialRetryTime: 2 * time.Millisecond,
					Retries:          3,
					Factor:           4,
				}, published.Retry)
			},
		},
		{
			name: "publish with messageId",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var published *common.MqttPublishArgs
				host.MqttClientTest = &enginetest.MqttClient{PublishFunc: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
					published = args
					return &common.MqttPublishResult{}, nil
				}}

				_, err := vm.RunString(`
					const mqtt = require("mokapi/mqtt")
					mqtt.publish({ topic: 't', messageId: 'alarm' })
				`)
				r.NoError(t, err)
				r.Equal(t, "alarm", published.MessageId)
				r.Nil(t, published.Value)
			},
		},
		{
			name: "publish null value is an empty payload",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var published *common.MqttPublishArgs
				host.MqttClientTest = &enginetest.MqttClient{PublishFunc: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
					published = args
					return &common.MqttPublishResult{}, nil
				}}

				_, err := vm.RunString(`
					const mqtt = require("mokapi/mqtt")
					mqtt.publish({ topic: 't', value: null })
				`)
				r.NoError(t, err)
				r.NotNil(t, published.Value)
				r.Empty(t, published.Value)
			},
		},
		{
			name: "messageId of wrong type",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.MqttClientTest = &enginetest.MqttClient{PublishFunc: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
					return &common.MqttPublishResult{}, nil
				}}

				_, err := vm.RunString(`
					const mqtt = require("mokapi/mqtt")
					mqtt.publish({ topic: 't', messageId: 12 })
				`)
				r.EqualError(t, err, "unexpected type for 'messageId': expected string, got Integer at mokapi/js/mqtt.(*Module).Publish-fm (native)")
			},
		},
		{
			name: "publish client error throws",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.MqttClientTest = &enginetest.MqttClient{PublishFunc: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
					return nil, fmt.Errorf("TEST ERROR")
				}}

				_, err := vm.RunString(`
					const mqtt = require("mokapi/mqtt")
					mqtt.publish({})
				`)
				r.EqualError(t, err, "TEST ERROR at mokapi/js/mqtt.(*Module).Publish-fm (native)")
			},
		},
		{
			name: "panic should not crash",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.MqttClientTest = &enginetest.MqttClient{PublishFunc: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
					panic("TEST")
				}}

				_, err := vm.RunString(`
					const mqtt = require("mokapi/mqtt")
					mqtt.publish({})
				`)
				r.EqualError(t, err, "TEST at mokapi/js/mqtt.(*Module).Publish-fm (native)")
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
			req.RegisterNativeModule("mokapi/mqtt", mqtt.Require)

			tc.test(t, vm, host)
		})
	}
}

func TestMqtt_PublishAsync(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		client func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error)
		state  string // "fulfilled" or "rejected"
		result any    // exported value or rejection reason of the promise
		thrown string // part of the exception when the call itself must throw
	}{
		{
			name:   "resolves with the client result",
			script: `mqtt.publishAsync({ topic: 'foo', value: 'hello' })`,
			client: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
				return &common.MqttPublishResult{}, nil
			},
			state:  "fulfilled",
			result: &common.MqttPublishResult{},
		},
		{
			name:   "invalid argument throws synchronously",
			script: `mqtt.publishAsync({ retry: 'foo' })`,
			thrown: "interface conversion",
		},
		{
			name:   "invalid retry throws synchronously",
			script: `mqtt.publishAsync({ retry: { maxRetryTime: 'foo' } })`,
			thrown: "parse maxRetryTime failed",
		},
		{
			name:   "client error rejects the promise",
			script: `mqtt.publishAsync({ topic: 'foo', value: 'hello' })`,
			client: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
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
			req.RegisterNativeModule("mokapi/mqtt", mqtt.Require)
			host.MqttClientTest = &enginetest.MqttClient{PublishFunc: tc.client}

			type outcome struct {
				state string
				value any
			}
			settled := make(chan outcome, 1)
			_, err = loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
				_ = vm.Set("settled", func(state string, v any) { settled <- outcome{state, v} })
				return vm.RunString(`const mqtt = require("mokapi/mqtt");` + tc.script +
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
