package mokapi_test

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/mokapi"
	"mokapi/js/require"
	"sync"
	"testing"
	"time"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

func TestModule_AppMqtt(t *testing.T) {
	type handler struct {
		Filter  common.MqttFilter
		Execute common.EventHandlerFunc
	}

	testcases := []struct {
		name          string
		script        string
		awaitPromise  string // global variable holding a promise to settle before the test runs
		logger        *enginetest.Logger
		publishResult *common.MqttPublishResult
		test          func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error)
	}{
		{
			name: "topic",
			script: `
const m = require('mokapi')
m.app.mqtt().topic('foo').message((msg) => {})
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				r.Len(t, handlers, 1)
				r.Equal(t, "foo", handlers[0].Filter.Topic)
			},
		},
		{
			name: "api and topic",
			script: `
const m = require('mokapi')
m.app.api("foo").mqtt().topic('bar').message((msg) => {})
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				r.Len(t, handlers, 1)
				r.Equal(t, "foo", handlers[0].Filter.Api)
				r.Equal(t, "bar", handlers[0].Filter.Topic)
			},
		},
		{
			name: "api on Kafka object",
			script: `
const m = require('mokapi')
m.app.mqtt().api("foo").message((msg) => {})
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				r.Len(t, handlers, 1)
				r.Equal(t, "foo", handlers[0].Filter.Api)
				r.Equal(t, "", handlers[0].Filter.Topic)
			},
		},
		{
			name: "publish to topic",
			script: `
const m = require('mokapi')
m.app.mqtt().publish('foo', { value: 'val', retain: true })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				r.Len(t, publishes, 1)
				r.Equal(t, "foo", publishes[0].Topic)
				r.Equal(t, "mokapi-script", publishes[0].ClientId)
				r.Equal(t, "", publishes[0].Cluster)
				r.Equal(t, common.RetryArgs{MaxRetryTime: 180000000000, InitialRetryTime: 500000000, Factor: 2, Retries: 10}, publishes[0].Retry)
				r.NotEmpty(t, publishes[0].ScriptFile)
				r.Equal(t, "val", string(publishes[0].Value))
				r.True(t, publishes[0].Retain)

			},
		},
		{
			name: "publish value using simple types",
			script: `
const m = require('mokapi');
['bar', 123, 12.3, true, null].map(x => m.app.mqtt().publish('foo', { value: x }))

`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				r.NoError(t, err)
				r.Len(t, publishes, 5)
				r.Equal(t, "foo", publishes[0].Topic)
				r.Equal(t, "bar", string(publishes[0].Value))
				r.Equal(t, "123", string(publishes[1].Value))
				r.Equal(t, "12.3", string(publishes[2].Value))
				r.Equal(t, "true", string(publishes[3].Value))
				r.Equal(t, "", string(publishes[4].Value))
			},
		},
		{
			name: "publish value using bytes",
			script: `
const m = require('mokapi')
let buffer = new ArrayBuffer(8);
let uint8View = new Uint8Array(buffer);
uint8View[0] = 255;
uint8View[1] = 128;
uint8View[2] = 64;
m.app.mqtt().publish('foo', { value: buffer })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				r.Equal(t, []byte{255, 128, 64, 0, 0, 0, 0, 0}, publishes[0].Value)
			},
		},
		{
			name: "publish value using object",
			script: `
const m = require('mokapi')
m.app.mqtt().publish('foo', { value: { foo: 'bar' } })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				r.Equal(t, `{"foo":"bar"}`, string(publishes[0].Value))
			},
		},
		{
			name: "publish with messageId",
			script: `
const m = require('mokapi')
m.app.mqtt().publish('foo', { messageId: 'alarm', data: { level: 'critical' } })
m.app.mqtt().publish('foo', { messageId: undefined })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				r.NoError(t, err)
				r.Len(t, publishes, 2)
				r.Equal(t, "alarm", publishes[0].MessageId)
				r.Equal(t, common.Optional[any]{IsSet: true, Value: map[string]any{"level": "critical"}}, publishes[0].Data)
				r.Equal(t, "", publishes[1].MessageId)
			},
		},
		{
			name: "publish null value is an empty payload, not a generated one",
			script: `
const m = require('mokapi')
m.app.mqtt().publish('foo', { value: null })
m.app.mqtt().publish('foo', {})
m.app.mqtt().publish('foo', { value: '' })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				r.NoError(t, err)
				r.NotNil(t, publishes[0].Value)
				r.Empty(t, publishes[0].Value)
				r.Nil(t, publishes[1].Value)
				r.NotNil(t, publishes[2].Value)
				r.Empty(t, publishes[2].Value)
			},
		},
		{
			name: "publish data null is set, no data is not",
			script: `
const m = require('mokapi')
m.app.mqtt().publish('foo', { data: null })
m.app.mqtt().publish('foo', {})
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				r.NoError(t, err)
				r.Equal(t, common.Optional[any]{IsSet: true}, publishes[0].Data)
				r.Equal(t, common.Optional[any]{}, publishes[1].Data)
			},
		},
		{
			name: "publish data using object",
			script: `
const m = require('mokapi')
m.app.mqtt().publish('foo', { data: { foo: 'bar' } })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				r.Equal(t, common.Optional[any]{IsSet: true, Value: map[string]any{"foo": "bar"}}, publishes[0].Data)
			},
		},
		{
			name: "publish using result function",
			publishResult: &common.MqttPublishResult{
				Value: "foo",
			},
			script: `
const m = require('mokapi')
let result;
m.app.mqtt().publish('foo', { }, { result: (r) => result = r })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				v, err := evalOnLoop(loop, "result")
				r.NoError(t, err)
				r.Equal(t, &common.MqttPublishResult{
					Value: "foo",
				}, v)
			},
		},
		{
			name:         "publishAsync using result function",
			awaitPromise: "p",
			publishResult: &common.MqttPublishResult{
				Value: "foo",
			},
			script: `
const m = require('mokapi')
let result;
const p = m.app.mqtt().publishAsync('foo', { }, { result: (r) => result = r })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				r.NoError(t, err)
				v, err := evalOnLoop(loop, "result")
				r.NoError(t, err)
				r.Equal(t, &common.MqttPublishResult{
					Value: "foo",
				}, v)
			},
		},
		{
			name: "publish calls result function with api, topic and value",
			publishResult: &common.MqttPublishResult{
				Api:   "broker",
				Topic: "foo",
				Value: "bar",
			},
			script: `
const m = require('mokapi')
let result;
m.app.mqtt().publish('foo', { }, { result: (r) => result = { api: r.api, topic: r.topic, value: r.value, cluster: r.cluster } })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				v, err := evalOnLoop(loop, "result")
				r.NoError(t, err)
				r.Equal(t, map[string]any{"api": "broker", "topic": "foo", "value": "bar", "cluster": nil}, v)
			},
		},
		{
			name: "publish using retry with numbers",
			script: `
const m = require('mokapi')
let result;
m.app.mqtt().publish('foo', { }, { retry: { maxRetryTime: 1, initialRetryTime: 2, factor: 3, retries: 4 } })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				r.Equal(t, 1*time.Millisecond, publishes[0].Retry.MaxRetryTime)
				r.Equal(t, 2*time.Millisecond, publishes[0].Retry.InitialRetryTime)
				r.Equal(t, 3, publishes[0].Retry.Factor)
				r.Equal(t, 4, publishes[0].Retry.Retries)
			},
		},
		{
			name: "publish using retry with strings",
			script: `
const m = require('mokapi')
let result;
m.app.mqtt().publish('foo', { }, { retry: { maxRetryTime: '1s', initialRetryTime: '2s' } })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, loop *eventloop.EventLoop, err error) {
				r.Equal(t, 1*time.Second, publishes[0].Retry.MaxRetryTime)
				r.Equal(t, 2*time.Second, publishes[0].Retry.InitialRetryTime)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			reg, err := require.NewRegistry(nil)
			reg.RegisterNativeModule("mokapi", mokapi.Require)
			r.NoError(t, err)

			vm := goja.New()
			vm.SetFieldNameMapper(scriptFieldNameMapper{})
			host := &enginetest.Host{}
			loop := eventloop.New(vm, host)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			// the host calls these functions on the event loop or on a send goroutine
			var mu sync.Mutex
			var handlers []handler
			host.OnMqttFunc = func(filter common.MqttFilter, do common.EventHandlerFunc, args common.EventArgs) {
				mu.Lock()
				defer mu.Unlock()
				handlers = append(handlers, handler{
					Filter:  filter,
					Execute: do,
				})
			}
			var publishes []*common.MqttPublishArgs
			host.MqttClientTest = &enginetest.MqttClient{
				PublishFunc: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
					mu.Lock()
					defer mu.Unlock()
					publishes = append(publishes, args)
					return tc.publishResult, nil
				},
			}

			_, err = evalOnLoop(loop, tc.script)
			if tc.awaitPromise != "" {
				r.NoError(t, err)
				r.Equal(t, goja.PromiseStateFulfilled, awaitPromise(t, loop, tc.awaitPromise))
			}
			mu.Lock()
			defer mu.Unlock()
			tc.test(t, handlers, publishes, loop, err)
		})
	}
}

func TestModule_AppMqtt_PublishAsync_Errors(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		client func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error)
		reason string // part of the rejection reason of the promise
		thrown string // part of the exception when the call itself must throw
	}{
		{
			name:   "messageId of wrong type throws synchronously",
			script: `m.app.mqtt().publishAsync('foo', { messageId: 12 })`,
			thrown: "unexpected type for 'messageId': expected string, got Integer",
		},
		{
			name:   "invalid retry throws synchronously",
			script: `m.app.mqtt().publishAsync('foo', { value: 'hello' }, { retry: 'foo' })`,
			thrown: "invalid property type 'retry'",
		},
		{
			name:   "topic: invalid result callback throws synchronously",
			script: `m.app.mqtt().topic('foo').publishAsync({ value: 'hello' }, { result: 'foo' })`,
			thrown: "expected function, got String",
		},
		{
			name:   "client error rejects the promise",
			script: `m.app.mqtt().publishAsync('foo', { value: 'hello' })`,
			client: func(*common.MqttPublishArgs) (*common.MqttPublishResult, error) {
				return nil, fmt.Errorf("TEST ERROR")
			},
			reason: "TEST ERROR",
		},
		{
			name:   "result callback error rejects the promise",
			script: `m.app.mqtt().publishAsync('foo', { value: 'hello' }, { result: () => { throw new Error('CALLBACK ERROR') } })`,
			reason: "CALLBACK ERROR",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			loop, host := newModuleLoop(t)
			host.MqttClientTest = &enginetest.MqttClient{
				PublishFunc: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
					if tc.client != nil {
						return tc.client(args)
					}
					return &common.MqttPublishResult{}, nil
				},
			}

			_, err := evalOnLoop(loop, `const m = require('mokapi')`)
			r.NoError(t, err)
			_, err = evalOnLoop(loop, `const p = `+tc.script)
			if tc.thrown != "" {
				r.ErrorContains(t, err, tc.thrown)
				return
			}
			r.NoError(t, err)

			state, reason := awaitPromiseResult(t, loop, "p")
			r.Equal(t, goja.PromiseStateRejected, state)
			r.Contains(t, reason, tc.reason)
		})
	}
}

func TestModule_AppMqtt_PublishAsync_SendsMessageAsItWasAtCallTime(t *testing.T) {
	loop, host := newModuleLoop(t)
	var mu sync.Mutex
	var calls []*common.MqttPublishArgs
	host.MqttClientTest = &enginetest.MqttClient{
		PublishFunc: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
			mu.Lock()
			defer mu.Unlock()
			calls = append(calls, args)
			return &common.MqttPublishResult{}, nil
		},
	}

	_, err := evalOnLoop(loop, `
const m = require('mokapi')
const msg = { value: 'hello' }
const p = m.app.mqtt().publishAsync('foo', msg)
msg.value = 'changed'
`)
	r.NoError(t, err)
	r.Equal(t, goja.PromiseStateFulfilled, awaitPromise(t, loop, "p"))

	mu.Lock()
	defer mu.Unlock()
	r.Len(t, calls, 1)
	r.Equal(t, "hello", string(calls[0].Value))
}

func TestModule_AppMqtt_Publish_Errors(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		client func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error)
		thrown string
	}{
		{
			name:   "client error throws",
			script: `m.app.mqtt().publish('foo', { value: 'hello' })`,
			client: func(*common.MqttPublishArgs) (*common.MqttPublishResult, error) {
				return nil, fmt.Errorf("TEST ERROR")
			},
			thrown: "TEST ERROR",
		},
		{
			name:   "topic: client error throws",
			script: `m.app.mqtt().topic('foo').publish({ value: 'hello' })`,
			client: func(*common.MqttPublishArgs) (*common.MqttPublishResult, error) {
				return nil, fmt.Errorf("TEST ERROR")
			},
			thrown: "TEST ERROR",
		},
		{
			name:   "result callback error throws",
			script: `m.app.mqtt().publish('foo', { value: 'hello' }, { result: () => { throw new Error('CALLBACK ERROR') } })`,
			client: func(*common.MqttPublishArgs) (*common.MqttPublishResult, error) {
				return &common.MqttPublishResult{}, nil
			},
			thrown: "CALLBACK ERROR",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			loop, host := newModuleLoop(t)
			host.MqttClientTest = &enginetest.MqttClient{PublishFunc: tc.client}

			_, err := evalOnLoop(loop, `const m = require('mokapi')`)
			r.NoError(t, err)
			_, err = evalOnLoop(loop, tc.script)
			r.ErrorContains(t, err, tc.thrown)
		})
	}
}
