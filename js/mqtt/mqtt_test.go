package mqtt_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/mqtt"
	"mokapi/js/require"
	"testing"

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
			req, err := require.NewRegistry()
			r.NoError(t, err)
			req.Enable(vm)
			req.RegisterNativeModule("mokapi/mqtt", mqtt.Require)

			tc.test(t, vm, host)
		})
	}
}
