package engine_test

import (
	"mokapi/engine"
	"mokapi/engine/common"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvent_MQTT(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "api filtering",
			test: func(t *testing.T) {
				d := &engine.MqttEventDispatcher{}

				var called []string
				d.Register("foo", &engine.MqttEventHandler{
					Filter: common.MqttFilter{
						Api: "foo",
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "foo")
						return true, nil
					},
				})
				d.Register("bar", &engine.MqttEventHandler{
					Filter: common.MqttFilter{
						Api: "bar",
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "bar")
						return true, nil
					},
				})

				d.EmitMqtt(&common.MqttMessageEvent{
					Api: "foo",
				})

				require.Equal(t, []string{"foo"}, called)
			},
		},
		{
			name: "topic filtering",
			test: func(t *testing.T) {
				d := &engine.MqttEventDispatcher{}

				var called []string
				d.Register("foo", &engine.MqttEventHandler{
					Filter: common.MqttFilter{
						Topic: "foo",
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "foo")
						return true, nil
					},
				})
				d.Register("bar", &engine.MqttEventHandler{
					Filter: common.MqttFilter{
						Topic: "bar",
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "bar")
						return true, nil
					},
				})

				d.EmitMqtt(&common.MqttMessageEvent{
					Topic: "foo",
				})

				require.Equal(t, []string{"foo"}, called)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
