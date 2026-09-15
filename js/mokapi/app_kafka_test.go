package mokapi_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/mokapi"
	"mokapi/js/require"
	"testing"
	"time"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

func TestModule_AppKafka(t *testing.T) {
	type handler struct {
		Filter  common.KafkaFilter
		Execute common.EventHandler
	}

	testcases := []struct {
		name          string
		script        string
		logger        *enginetest.Logger
		produceResult *common.KafkaProduceResult
		test          func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error)
	}{
		{
			name: "topic",
			script: `
const m = require('mokapi')
m.app.kafka().topic('foo').message((msg) => {})
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
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
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
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
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
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
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
				r.Len(t, produces, 1)
				r.Equal(t, "foo", produces[0].Topic)
				r.Equal(t, "mokapi-script", produces[0].ClientId)
				r.Equal(t, "", produces[0].Cluster)
				r.Equal(t, common.RetryArgs{MaxRetryTime: 180000000000, InitialRetryTime: 500000000, Factor: 2, Retries: 10}, produces[0].Retry)
				r.NotEmpty(t, produces[0].ScriptFile)
				r.Equal(t, "key", produces[0].Messages[0].Key)
				r.Equal(t, "val", string(produces[0].Messages[0].Value))
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
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
				r.Len(t, produces, 1)
				r.Equal(t, "foo", produces[0].Topic)
				r.Equal(t, "key1", produces[0].Messages[0].Key)
				r.Equal(t, "key2", produces[0].Messages[1].Key)
			},
		},
		{
			name: "produce key is number",
			script: `
const m = require('mokapi')
m.app.kafka().produce('foo', { key: 123 })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
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
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
				r.Len(t, produces, 1)
				r.Equal(t, "foo", produces[0].Topic)
				r.Len(t, produces[0].Messages, 5)
				r.Equal(t, "bar", string(produces[0].Messages[0].Value))
				r.Equal(t, "123", string(produces[0].Messages[1].Value))
				r.Equal(t, "12.3", string(produces[0].Messages[2].Value))
				r.Equal(t, "true", string(produces[0].Messages[3].Value))
				r.Equal(t, "", string(produces[0].Messages[4].Value))
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
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
				r.Equal(t, []byte{255, 128, 64, 0, 0, 0, 0, 0}, produces[0].Messages[0].Value)
			},
		},
		{
			name: "produce value using object",
			script: `
const m = require('mokapi')
m.app.kafka().produce('foo', { value: { foo: 'bar' } })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
				r.Equal(t, `{"foo":"bar"}`, string(produces[0].Messages[0].Value))
			},
		},
		{
			name: "produce data using object",
			script: `
const m = require('mokapi')
m.app.kafka().produce('foo', { data: { foo: 'bar' } })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
				r.Equal(t, map[string]any{"foo": "bar"}, produces[0].Messages[0].Data)
			},
		},
		{
			name: "produce data using header",
			script: `
const m = require('mokapi')
m.app.kafka().produce('foo', { headers: { foo: 'bar', bar: { name: 'carol' } } })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
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
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
				v := vm.Get("result").Export()
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
			name: "produceAsync using result function",
			produceResult: &common.KafkaProduceResult{
				Messages: []common.KafkaMessageResult{{Key: "foo"}},
			},
			script: `
const m = require('mokapi')
let result;
const p = m.app.kafka().produceAsync('foo', { }, { result: (r) => result = r })
`,
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
				p := vm.Get("p").Export().(*goja.Promise)
				counter := 0
				for p.State() == goja.PromiseStatePending && counter < 10 {
					time.Sleep(300 * time.Millisecond)
					counter++
				}
				v := vm.Get("result").Export()
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
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
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
			test: func(t *testing.T, handlers []handler, produces []*common.KafkaProduceArgs, vm *goja.Runtime, err error) {
				r.Equal(t, 1*time.Second, produces[0].Retry.MaxRetryTime)
				r.Equal(t, 2*time.Second, produces[0].Retry.InitialRetryTime)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			reg, err := require.NewRegistry()
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

			var handlers []handler
			host.OnKafkaFunc = func(filter common.KafkaFilter, do common.EventHandler, args common.EventArgs) {
				handlers = append(handlers, handler{
					Filter:  filter,
					Execute: do,
				})
			}
			var produces []*common.KafkaProduceArgs
			host.KafkaClientTest = &enginetest.KafkaClient{
				ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					produces = append(produces, args)
					return tc.produceResult, nil
				},
			}

			_, err = vm.RunString(tc.script)
			tc.test(t, handlers, produces, vm, err)
		})
	}
}
