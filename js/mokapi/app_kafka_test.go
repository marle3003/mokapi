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

func TestModule_AppKafka(t *testing.T) {
	type handler struct {
		Filter  common.KafkaFilter
		Execute common.EventHandlerFunc
	}

	testcases := []struct {
		name          string
		script        string
		awaitPromise  string // global variable holding a promise to settle before the test runs
		logger        *enginetest.Logger
		produceResult *common.KafkaProduceResult
		test          func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error)
	}{
		{
			name: "topic",
			script: `
const m = require('mokapi')
m.app.kafka().topic('foo').message((msg) => {})
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				r.Len(t, handlers, 1)
				r.Equal(t, "foo", handlers[0].Filter.Topic)
			},
		},
		{
			name: "api and topic",
			script: `
const m = require('mokapi')
m.app.api("foo").kafka().topic('bar').message((msg) => {})
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				r.Len(t, handlers, 1)
				r.Equal(t, "foo", handlers[0].Filter.Api)
				r.Equal(t, "bar", handlers[0].Filter.Topic)
			},
		},
		{
			name: "api on Kafka object",
			script: `
const m = require('mokapi')
m.app.kafka().api("foo").message((msg) => {})
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				r.Len(t, handlers, 1)
				r.Equal(t, "foo", handlers[0].Filter.Api)
				r.Equal(t, "", handlers[0].Filter.Topic)
			},
		},
		{
			name: "produce to topic",
			script: `
const m = require('mokapi')
m.app.kafka().produce('foo', { key: 'key', value: 'val' })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				r.Len(t, produces, 1)
				r.Equal(t, "foo", produces[0].Topic)
				r.Equal(t, "mokapi-script", produces[0].ClientId)
				r.Equal(t, "", produces[0].Cluster)
				r.Equal(t, common.RetryArgs{MaxRetryTime: 180000000000, InitialRetryTime: 500000000, Factor: 2, Retries: 10}, produces[0].Retry)
				r.NotEmpty(t, produces[0].ScriptFile)
				r.Equal(t, "key", produces[0].Messages[0].Key)
				r.Equal(t, "val", string(produces[0].Messages[0].Value.Value))
				r.Equal(t, -1, produces[0].Messages[0].Partition)
				r.Nil(t, produces[0].Messages[0].Headers)

			},
		},
		{
			name: "produce two messages to topic",
			script: `
const m = require('mokapi')
m.app.kafka().produce('foo', [{ key: 'key1' },{ key: 'key2' }])
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				r.Len(t, produces, 1)
				r.Equal(t, "foo", produces[0].Topic)
				r.Equal(t, "key1", produces[0].Messages[0].Key)
				r.Equal(t, "key2", produces[0].Messages[1].Key)
			},
		},
		{
			name: "produce messages with messageId",
			script: `
const m = require('mokapi')
m.app.kafka().produce('foo', [{ messageId: 'orderCreated' }, { messageId: 'orderDeleted', value: null }, { key: 'k' }, { messageId: undefined }])
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				r.NoError(t, err)
				r.Len(t, produces, 1)
				r.Equal(t, "orderCreated", produces[0].Messages[0].MessageId)
				r.Equal(t, "orderDeleted", produces[0].Messages[1].MessageId)
				r.True(t, produces[0].Messages[1].Value.IsSet)
				r.Equal(t, "", produces[0].Messages[2].MessageId)
				r.Equal(t, "", produces[0].Messages[3].MessageId)
			},
		},
		{
			name: "produce key is number",
			script: `
const m = require('mokapi')
m.app.kafka().produce('foo', { key: 123 })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				r.Len(t, produces, 1)
				r.Equal(t, "foo", produces[0].Topic)
				r.Equal(t, int64(123), produces[0].Messages[0].Key)
			},
		},
		{
			name: "produce value using simple types",
			script: `
const m = require('mokapi')
m.app.kafka().produce('foo', ['bar', 123, 12.3, true, null].map(x => { return { value: x }}))
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				r.Len(t, produces, 1)
				r.Equal(t, "foo", produces[0].Topic)
				r.Len(t, produces[0].Messages, 5)
				r.Equal(t, "bar", string(produces[0].Messages[0].Value.Value))
				r.Equal(t, "123", string(produces[0].Messages[1].Value.Value))
				r.Equal(t, "12.3", string(produces[0].Messages[2].Value.Value))
				r.Equal(t, "true", string(produces[0].Messages[3].Value.Value))
				r.Equal(t, "", string(produces[0].Messages[4].Value.Value))
			},
		},
		{
			name: "produce value using bytes",
			script: `
const m = require('mokapi')
let buffer = new ArrayBuffer(8);
let uint8View = new Uint8Array(buffer);
uint8View[0] = 255;
uint8View[1] = 128;
uint8View[2] = 64;
m.app.kafka().produce('foo', { value: buffer })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				r.Equal(t, []byte{255, 128, 64, 0, 0, 0, 0, 0}, produces[0].Messages[0].Value.Value)
			},
		},
		{
			name: "produce value using object",
			script: `
const m = require('mokapi')
m.app.kafka().produce('foo', { value: { foo: 'bar' } })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				r.Equal(t, `{"foo":"bar"}`, string(produces[0].Messages[0].Value.Value))
			},
		},
		{
			name: "produce data using object",
			script: `
const m = require('mokapi')
m.app.kafka().produce('foo', { data: { foo: 'bar' } })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				r.Equal(t, map[string]any{"foo": "bar"}, produces[0].Messages[0].Data.Value)
			},
		},
		{
			name: "produce data using header",
			script: `
const m = require('mokapi')
m.app.kafka().produce('foo', { headers: { foo: 'bar', bar: { name: 'carol' } } })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				r.Equal(t, map[string]string{"bar": "{\"name\":\"carol\"}", "foo": "bar"}, produces[0].Messages[0].Headers)
			},
		},
		{
			name: "produce using result function",
			produceResult: &common.KafkaProduceResult{
				Messages: []common.KafkaMessageResult{{Key: "foo"}},
			},
			script: `
const m = require('mokapi')
let result;
m.app.kafka().produce('foo', { }, { result: (r) => result = r })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				v, err := evalOnLoop(loop, "result")
				r.NoError(t, err)
				r.Equal(t, common.KafkaMessageResult{
					Key:       "foo",
					Value:     "",
					Offset:    0,
					Headers:   map[string]string(nil),
					Partition: 0,
				}, v)
			},
		},
		{
			name:         "produceAsync using result function",
			awaitPromise: "p",
			produceResult: &common.KafkaProduceResult{
				Messages: []common.KafkaMessageResult{{Key: "foo"}},
			},
			script: `
const m = require('mokapi')
let result;
const p = m.app.kafka().produceAsync('foo', { }, { result: (r) => result = r })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				v, err := evalOnLoop(loop, "result")
				r.NoError(t, err)
				r.Equal(t, common.KafkaMessageResult{
					Key:       "foo",
					Value:     "",
					Offset:    0,
					Headers:   map[string]string(nil),
					Partition: 0,
				}, v)
			},
		},
		{
			name: "produce using retry with numbers",
			script: `
const m = require('mokapi')
let result;
m.app.kafka().produce('foo', { }, { retry: { maxRetryTime: 1, initialRetryTime: 2, factor: 3, retries: 4 } })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				r.Equal(t, 1*time.Millisecond, produces[0].Retry.MaxRetryTime)
				r.Equal(t, 2*time.Millisecond, produces[0].Retry.InitialRetryTime)
				r.Equal(t, 3, produces[0].Retry.Factor)
				r.Equal(t, 4, produces[0].Retry.Retries)
			},
		},
		{
			name: "produce using retry with strings",
			script: `
const m = require('mokapi')
let result;
m.app.kafka().produce('foo', { }, { retry: { maxRetryTime: '1s', initialRetryTime: '2s' } })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, loop *eventloop.EventLoop, err error) {
				r.Equal(t, 1*time.Second, produces[0].Retry.MaxRetryTime)
				r.Equal(t, 2*time.Second, produces[0].Retry.InitialRetryTime)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			reg, err := require.NewRegistry(nil)
			reg.RegisterNativeModule("mokapi", mokapi.Require)
			r.NoError(t, err)

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			loop := eventloop.New(vm, host)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			// the host calls these functions on the event loop or on a send goroutine
			var mu sync.Mutex
			var handlers []handler
			host.OnKafkaFunc = func(filter common.KafkaFilter, do common.EventHandlerFunc, args common.EventArgs) {
				mu.Lock()
				defer mu.Unlock()
				handlers = append(handlers, handler{
					Filter:  filter,
					Execute: do,
				})
			}
			var produces []*common.KafkaProduceArgs
			host.KafkaClientTest = &enginetest.KafkaClient{
				ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					mu.Lock()
					defer mu.Unlock()
					produces = append(produces, args)
					return tc.produceResult, nil
				},
			}

			_, err = evalOnLoop(loop, tc.script)
			if tc.awaitPromise != "" {
				r.NoError(t, err)
				r.Equal(t, goja.PromiseStateFulfilled, awaitPromise(t, loop, tc.awaitPromise))
			}
			mu.Lock()
			defer mu.Unlock()
			tc.test(t, handlers, produces, loop, err)
		})
	}
}

func TestModule_AppKafka_ProduceAsync_Errors(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		client func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error)
		reason string // part of the rejection reason of the promise
		thrown string // part of the exception when the call itself must throw
	}{
		{
			name:   "messageId of wrong type throws synchronously",
			script: `m.app.kafka().produceAsync('foo', { messageId: 12 })`,
			thrown: "unexpected type for 'messageId': expected string, got Integer",
		},
		{
			name:   "invalid retry throws synchronously",
			script: `m.app.kafka().produceAsync('foo', { value: 'hello' }, { retry: 'foo' })`,
			thrown: "invalid property type 'retry'",
		},
		{
			name:   "topic: invalid result callback throws synchronously",
			script: `m.app.kafka().topic('foo').produceAsync({ value: 'hello' }, { result: 'foo' })`,
			thrown: "expected function, got String",
		},
		{
			name:   "client error rejects the promise",
			script: `m.app.kafka().produceAsync('foo', { value: 'hello' })`,
			client: func(*common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
				return nil, fmt.Errorf("TEST ERROR")
			},
			reason: "TEST ERROR",
		},
		{
			name:   "result callback error rejects the promise",
			script: `m.app.kafka().produceAsync('foo', { value: 'hello' }, { result: () => { throw new Error('CALLBACK ERROR') } })`,
			client: func(*common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
				return &common.KafkaProduceResult{Messages: []common.KafkaMessageResult{{Key: "foo"}}}, nil
			},
			reason: "CALLBACK ERROR",
		},
		{
			name:   "nil client result with result callback rejects the promise",
			script: `m.app.kafka().produceAsync('foo', { value: 'hello' }, { result: () => {} })`,
			client: func(*common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
				return nil, nil
			},
			reason: "nil pointer",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			loop, host := newModuleLoop(t)
			host.KafkaClientTest = &enginetest.KafkaClient{
				ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					if tc.client != nil {
						return tc.client(args)
					}
					return &common.KafkaProduceResult{}, nil
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

func TestModule_AppKafka_ProduceAsync_SendsMessageAsItWasAtCallTime(t *testing.T) {
	loop, host := newModuleLoop(t)
	var mu sync.Mutex
	var calls []*common.KafkaProduceArgs
	host.KafkaClientTest = &enginetest.KafkaClient{
		ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
			mu.Lock()
			defer mu.Unlock()
			calls = append(calls, args)
			return &common.KafkaProduceResult{}, nil
		},
	}

	_, err := evalOnLoop(loop, `
const m = require('mokapi')
const msg = { value: 'hello' }
const p = m.app.kafka().produceAsync('foo', msg)
msg.value = 'changed'
`)
	r.NoError(t, err)
	r.Equal(t, goja.PromiseStateFulfilled, awaitPromise(t, loop, "p"))

	mu.Lock()
	defer mu.Unlock()
	r.Len(t, calls, 1)
	r.Equal(t, "hello", string(calls[0].Messages[0].Value.Value))
}

func TestModule_AppKafka_Produce_Errors(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		client func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error)
		thrown string
	}{
		{
			name:   "client error throws",
			script: `m.app.kafka().produce('foo', { value: 'hello' })`,
			client: func(*common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
				return nil, fmt.Errorf("TEST ERROR")
			},
			thrown: "TEST ERROR",
		},
		{
			name:   "topic: client error throws",
			script: `m.app.kafka().topic('foo').produce({ value: 'hello' })`,
			client: func(*common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
				return nil, fmt.Errorf("TEST ERROR")
			},
			thrown: "TEST ERROR",
		},
		{
			name:   "result callback error throws",
			script: `m.app.kafka().produce('foo', { value: 'hello' }, { result: () => { throw new Error('CALLBACK ERROR') } })`,
			client: func(*common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
				return &common.KafkaProduceResult{Messages: []common.KafkaMessageResult{{Key: "foo"}}}, nil
			},
			thrown: "CALLBACK ERROR",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			loop, host := newModuleLoop(t)
			host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: tc.client}

			_, err := evalOnLoop(loop, `const m = require('mokapi')`)
			r.NoError(t, err)
			_, err = evalOnLoop(loop, tc.script)
			r.ErrorContains(t, err, tc.thrown)
		})
	}
}
