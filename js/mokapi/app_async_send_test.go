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

// The goja runtime is not goroutine-safe. The async send methods must read the message and
// arguments on the event loop and only hand plain Go values to the goroutine that sends them.
// Each script keeps working on the loop (mutating the message object) right after calling
// the async method, so any access to the message from another goroutine is a data race
// that is reported when run with `go test -race`.
func TestModule_AppAsyncSend_DoesNotAccessRuntimeOffEventLoop(t *testing.T) {
	testcases := []struct {
		name    string
		message string
		call    string
	}{
		{name: "websocket sendAsync", message: `{ text: 'hello' }`, call: `m.app.websocket().sendAsync('/foo', msg)`},
		{name: "websocket channel sendAsync", message: `{ text: 'hello' }`, call: `m.app.websocket().channel('/foo').sendAsync(msg)`},
		{name: "kafka produceAsync", message: `{ value: 'hello' }`, call: `m.app.kafka().produceAsync('foo', msg)`},
		{name: "kafka topic produceAsync", message: `{ value: 'hello' }`, call: `m.app.kafka().topic('foo').produceAsync(msg)`},
		{name: "mqtt publishAsync", message: `{ value: 'hello' }`, call: `m.app.mqtt().publishAsync('foo', msg)`},
		{name: "mqtt topic publishAsync", message: `{ value: 'hello' }`, call: `m.app.mqtt().topic('foo').publishAsync(msg)`},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			reg, err := require.NewRegistry(nil)
			r.NoError(t, err)
			reg.RegisterNativeModule("mokapi", mokapi.Require)

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			loop := eventloop.New(vm, host)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			host.WebsocketClientTest = &enginetest.WebsocketClient{
				SendFunc: func(args common.WebsocketSendArgs) (*common.WebsocketSendResult, error) {
					return nil, nil
				},
			}
			host.KafkaClientTest = &enginetest.KafkaClient{
				ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				},
			}
			host.MqttClientTest = &enginetest.MqttClient{
				PublishFunc: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
					return &common.MqttPublishResult{}, nil
				},
			}

			// the script and its promise callbacks run on the event loop; the test only
			// waits on a channel and never touches the runtime itself
			settled := make(chan string, 1)
			_, err = loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
				_ = vm.Set("settled", func(state string) { settled <- state })
				return vm.RunString(`
const m = require('mokapi')
const msg = ` + tc.message + `
const p = ` + tc.call + `
// keep the loop busy with work on the message while the send is in flight
for (let i = 0; i < 20000; i++) {
  msg.value = 'v' + i
  msg.text = 'v' + i
  msg['key' + (i % 100)] = i
}
p.then(() => settled('fulfilled'), (e) => settled('rejected: ' + e))
`)
			})
			r.NoError(t, err)

			select {
			case state := <-settled:
				r.Equal(t, "fulfilled", state)
			case <-time.After(5 * time.Second):
				t.Fatal("promise was not settled")
			}
		})
	}
}
