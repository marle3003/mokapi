package server_test

import (
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/schema/json/schema"
	"mokapi/server"
	"mokapi/try"
	"net"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestMqttServer(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name string
		test func(t *testing.T, m *server.MqttManager)
	}{
		{
			name: "TestMqttServer",
			test: func(t *testing.T, m *server.MqttManager) {
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

				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}, Data: c}})

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
			test: func(t *testing.T, m *server.MqttManager) {
				port := try.GetFreePort()
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				cfg := asyncapi3test.NewConfig(
					asyncapi3test.WithTitle("foo"),
					asyncapi3test.WithServer("mqtt12", "mqtt", addr),
					asyncapi3test.WithServer("kafka", "kafka", fmt.Sprintf("127.0.0.1:%v", try.GetFreePort())),
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
				c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}, Data: cfg}
				err := cfg.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)

				m.UpdateConfig(dynamic.ConfigEvent{Config: c})

				// wait for mqtt start
				time.Sleep(500 * time.Millisecond)

				client := mqtttest.NewClient(addr)
				defer client.Close()
				msg, err := client.Send(&mqtt.Message{
					Header:  &mqtt.Header{Type: mqtt.CONNECT},
					Payload: &mqtt.ConnectRequest{ClientId: "client"},
				})
				require.NoError(t, err)
				require.IsType(t, &mqtt.ConnectResponse{}, msg.Payload)
				require.Equal(t, mqtt.Success.Code, msg.Payload.(*mqtt.ConnectResponse).ReasonCode.Code)

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

			m := server.NewMqttManager(nil, runtime.New(&static.Config{}, &dynamictest.Reader{}))
			defer m.Stop()

			tc.test(t, m)
		})
	}
}

func TestMqttManager_PortIsBoundWhenUpdateConfigReturns(t *testing.T) {
	m := server.NewMqttManager(nil, runtime.New(&static.Config{}, &dynamictest.Reader{}))
	defer m.Stop()

	requireBoundWhenUpdateReturns(t, func(port int) { m.UpdateConfig(mqttConfigEvent(port)) })
}

func TestMqttManager_BindErrorIsLoggedAndServerStaysRegistered(t *testing.T) {
	m := server.NewMqttManager(nil, runtime.New(&static.Config{}, &dynamictest.Reader{}))
	defer m.Stop()

	requireBindErrorIsLoggedAndServerStaysRegistered(t,
		func(port int) { m.UpdateConfig(mqttConfigEvent(port)) },
		func(port int) string { return fmt.Sprintf("listen tcp :%d: bind: address already in use", port) },
	)
}

func mqttConfigEvent(port int) dynamic.ConfigEvent {
	cfg := asyncapi3test.NewConfig(
		asyncapi3test.WithTitle("foo"),
		asyncapi3test.WithServer("mqtt", "mqtt", fmt.Sprintf("127.0.0.1:%d", port)),
	)
	return dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}, Data: cfg}}
}

func TestMqttManager_DeleteRemovesConfigAndStopsServer(t *testing.T) {
	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	m := server.NewMqttManager(nil, app)
	defer m.Stop()

	port := try.GetFreePort()
	e := mqttConfigEvent(port)
	m.UpdateConfig(e)
	require.NotNil(t, app.Mqtt.Get("foo"))

	e.Event = dynamic.Delete
	require.NotPanics(t, func() { m.UpdateConfig(e) })

	require.Nil(t, app.Mqtt.Get("foo"), "config must be removed from the MQTT store")
	requirePortClosed(t, port)
}

func TestMqttManager_DeleteDoesNotTouchKafkaClusterWithSameName(t *testing.T) {
	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	m := server.NewMqttManager(nil, app)
	defer m.Stop()

	kafkaCfg := &dynamic.Config{
		Info: dynamic.ConfigInfo{Url: try.MustUrl("kafka.yml")},
		Data: asyncapi3test.NewConfig(
			asyncapi3test.WithTitle("foo"),
			asyncapi3test.WithServer("kafka", "kafka", fmt.Sprintf("127.0.0.1:%d", try.GetFreePort())),
		),
	}
	_, err := app.Kafka.Add(kafkaCfg, nil)
	require.NoError(t, err)

	port := try.GetFreePort()
	e := mqttConfigEvent(port)
	m.UpdateConfig(e)
	e.Event = dynamic.Delete
	require.NotPanics(t, func() { m.UpdateConfig(e) })

	require.NotNil(t, app.Kafka.Get("foo"))
	require.NotNil(t, app.Kafka.Get("foo").Config, "kafka config must survive a MQTT delete")
	require.Nil(t, app.Mqtt.Get("foo"), "config must be removed from the MQTT store")
	requirePortClosed(t, port)
}

func requirePortClosed(t *testing.T, port int) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err == nil {
		_ = conn.Close()
	}
	require.Error(t, err, "port must be closed after delete")
}
