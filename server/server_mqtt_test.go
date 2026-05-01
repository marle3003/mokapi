package server

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/schema/json/schema"
	"mokapi/try"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMqttServer(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, m *MqttManager)
	}{
		{
			name: "TestMqttServer",
			test: func(t *testing.T, m *MqttManager) {
				port := try.GetFreePort()
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				c := asyncapi3test.NewConfig(
					asyncapi3test.WithTitle("foo"),
					asyncapi3test.WithServer("mqtt12", "mqtt", addr),
					asyncapi3test.WithChannel("foo",
						asyncapi3test.WithMessage("foo",
							asyncapi3test.WithPayload(
								&schema.Schema{Type: schema.Types{"string"}},
							),
						),
					),
				)

				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}, Data: c}})

				// wait for mqtt start
				time.Sleep(500 * time.Millisecond)

				client := mqtttest.NewClient(addr)
				defer client.Close()
				_, err := client.Send(&mqtt.Message{
					Header:  &mqtt.Header{Type: mqtt.CONNECT},
					Payload: &mqtt.ConnectRequest{},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "kafka topic should not be available",
			test: func(t *testing.T, m *MqttManager) {
				port := try.GetFreePort()
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				cfg := asyncapi3test.NewConfig(
					asyncapi3test.WithTitle("foo"),
					asyncapi3test.WithServer("mqtt12", "mqtt", addr),
					asyncapi3test.WithServer("kafka", "kafka", addr),
					asyncapi3test.WithChannel("foo",
						asyncapi3test.WithMessage("foo",
							asyncapi3test.WithPayload(
								&schema.Schema{Type: schema.Types{"string"}},
							),
						),
						asyncapi3test.AssignToServer("#/servers/mqtt12"),
					),
					asyncapi3test.WithChannel("bar",
						asyncapi3test.WithMessage("bar",
							asyncapi3test.WithPayload(
								&schema.Schema{Type: schema.Types{"string"}},
							),
						),
						asyncapi3test.AssignToServer("#/servers/kafka"),
					),
				)
				c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}, Data: cfg}
				err := cfg.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)

				m.UpdateConfig(dynamic.ConfigEvent{Config: c})

				// wait for mqtt start
				time.Sleep(500 * time.Millisecond)

				client := mqtttest.NewClient(addr)
				defer client.Close()
				msg, err := client.Send(&mqtt.Message{
					Header:  &mqtt.Header{Type: mqtt.CONNECT},
					Payload: &mqtt.ConnectRequest{},
				})
				require.NoError(t, err)
				require.IsType(t, &mqtt.ConnectResponse{}, msg.Payload)

				msg, err = client.Send(&mqtt.Message{
					Header:  &mqtt.Header{Type: mqtt.PUBLISH, QoS: 1},
					Payload: &mqtt.PublishRequest{Topic: "bar", MessageId: uint16(123)},
				})
				require.NoError(t, err)
				require.IsType(t, &mqtt.PublishResponse{}, msg.Payload)
				require.Equal(t, mqtt.TopicNameInvalid, msg.Payload.(*mqtt.PublishResponse).ReasonCode)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := NewMqttManager(nil, runtime.New(&static.Config{}, &dynamictest.Reader{}))
			defer m.Stop()

			tc.test(t, m)
		})
	}
}
