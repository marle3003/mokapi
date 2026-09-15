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

func TestModule_AppMqtt(t *testing.T) {
	type handler struct {
		Filter  common.MqttFilter
		Execute common.EventHandler
	}

	testcases := []struct {
		name          string
		script        string
		logger        *enginetest.Logger
		publishResult *common.MqttPublishResult
		test          func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, vm *goja.Runtime, err error)
	}{
		{
			name: "topic",
			script: `
const m = require('mokapi')
m.app.mqtt().topic('foo').message((msg) => {})
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, vm *goja.Runtime, err error) {
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
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, vm *goja.Runtime, err error) {
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
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, vm *goja.Runtime, err error) {
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
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, vm *goja.Runtime, err error) {
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
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, vm *goja.Runtime, err error) {
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
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, vm *goja.Runtime, err error) {
				r.Equal(t, []byte{255, 128, 64, 0, 0, 0, 0, 0}, publishes[0].Value)
			},
		},
		{
			name: "publish value using object",
			script: `
const m = require('mokapi')
m.app.mqtt().publish('foo', { value: { foo: 'bar' } })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, vm *goja.Runtime, err error) {
				r.Equal(t, `{"foo":"bar"}`, string(publishes[0].Value))
			},
		},
		{
			name: "publish data using object",
			script: `
const m = require('mokapi')
m.app.mqtt().publish('foo', { data: { foo: 'bar' } })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, vm *goja.Runtime, err error) {
				r.Equal(t, map[string]any{"foo": "bar"}, publishes[0].Data)
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
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, vm *goja.Runtime, err error) {
				v := vm.Get("result").Export()
				r.Equal(t, &common.MqttPublishResult{
					Value: "foo",
				}, v)
			},
		},
		{
			name: "publishAsync using result function",
			publishResult: &common.MqttPublishResult{
				Value: "foo",
			},
			script: `
const m = require('mokapi')
let result;
const p = m.app.mqtt().publishAsync('foo', { }, { result: (r) => result = r })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, vm *goja.Runtime, err error) {
				r.NoError(t, err)
				p := vm.Get("p").Export().(*goja.Promise)
				counter := 0
				for p.State() == goja.PromiseStatePending && counter < 10 {
					time.Sleep(300 * time.Millisecond)
					counter++
				}
				v := vm.Get("result").Export()
				r.Equal(t, &common.MqttPublishResult{
					Value: "foo",
				}, v)
			},
		},
		{
			name: "publish using retry with numbers",
			script: `
const m = require('mokapi')
let result;
m.app.mqtt().publish('foo', { }, { retry: { maxRetryTime: 1, initialRetryTime: 2, factor: 3, retries: 4 } })
`,
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, vm *goja.Runtime, err error) {
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
			test: func(t *testing.T, handlers []handler, publishes []*common.MqttPublishArgs, vm *goja.Runtime, err error) {
				r.Equal(t, 1*time.Second, publishes[0].Retry.MaxRetryTime)
				r.Equal(t, 2*time.Second, publishes[0].Retry.InitialRetryTime)
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
			host.OnMqttFunc = func(filter common.MqttFilter, do common.EventHandler, args common.EventArgs) {
				handlers = append(handlers, handler{
					Filter:  filter,
					Execute: do,
				})
			}
			var publishes []*common.MqttPublishArgs
			host.MqttClientTest = &enginetest.MqttClient{
				PublishFunc: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
					publishes = append(publishes, args)
					return tc.publishResult, nil
				},
			}

			_, err = vm.RunString(tc.script)
			tc.test(t, handlers, publishes, vm, err)
		})
	}
}
