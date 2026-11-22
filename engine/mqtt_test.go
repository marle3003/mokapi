package engine_test

import (
	"io"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/mqtt/store"
	"mokapi/runtime"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema/schematest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestMqttClient_Publish(t *testing.T) {
	logrus.SetOutput(io.Discard)

	// topic "sensors/events" with the messages reading and alarm, topic "sensors/ping" with a
	// message without payload
	newConfig := func() *asyncapi3.Config {
		reading := asyncapi3test.NewMessage(
			asyncapi3test.WithPayload(schematest.New("object",
				schematest.WithProperty("kind", schematest.New("string", schematest.WithEnumValues("reading"))),
				schematest.WithRequired("kind"),
			)),
		)
		alarm := asyncapi3test.NewMessage(
			asyncapi3test.WithPayload(schematest.New("object",
				schematest.WithProperty("kind", schematest.New("string", schematest.WithEnumValues("alarm"))),
				schematest.WithRequired("kind"),
			)),
		)

		events := asyncapi3test.NewChannel(
			asyncapi3test.WithChannelAddress("sensors/events"),
			asyncapi3test.UseMessage("reading", &asyncapi3.MessageRef{Value: reading}),
			asyncapi3test.UseMessage("alarm", &asyncapi3.MessageRef{Value: alarm}),
		)
		nulls := asyncapi3test.NewChannel(
			asyncapi3test.WithChannelAddress("sensors/null"),
			asyncapi3test.UseMessage("nullable", &asyncapi3.MessageRef{Value: asyncapi3test.NewMessage(asyncapi3test.WithPayload(schematest.New("null")))}),
		)
		ping := asyncapi3test.NewChannel(
			asyncapi3test.WithChannelAddress("sensors/ping"),
			asyncapi3test.UseMessage("ping", &asyncapi3.MessageRef{Value: asyncapi3test.NewMessage()}),
		)
		return asyncapi3test.NewConfig(
			asyncapi3test.WithInfo("foo", "", ""),
			asyncapi3test.AddChannel("events", events),
			asyncapi3test.AddChannel("nulls", nulls),
			asyncapi3test.AddChannel("ping", ping),
		)
	}

	testcases := []struct {
		name string
		args common.MqttPublishArgs
		test func(t *testing.T, retained string, result *common.MqttPublishResult, err error)
	}{
		{
			name: "payload is generated from the selected message",
			args: common.MqttPublishArgs{MessageId: "alarm"},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kind":"alarm"}`, retained)
				require.Equal(t, `{"kind":"alarm"}`, result.Value)
			},
		},
		{
			name: "data is encoded with the selected message",
			args: common.MqttPublishArgs{MessageId: "reading", Data: common.Optional[any]{IsSet: true, Value: map[string]any{"kind": "reading"}}},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kind":"reading"}`, retained)
			},
		},
		{
			name: "data is validated against the selected message only",
			// valid for reading, but alarm is selected
			args: common.MqttPublishArgs{MessageId: "alarm", Data: common.Optional[any]{IsSet: true, Value: map[string]any{"kind": "reading"}}},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.Error(t, err)
				require.Empty(t, retained)
			},
		},
		{
			name: "data without messageId is validated against the messages of the topic",
			args: common.MqttPublishArgs{Data: common.Optional[any]{IsSet: true, Value: map[string]any{"kind": "alarm"}}},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kind":"alarm"}`, retained)
			},
		},
		{
			name: "data null is validated against the selected message",
			args: common.MqttPublishArgs{Topic: "sensors/null", MessageId: "nullable", Data: common.Optional[any]{IsSet: true}},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.NoError(t, err)
				require.Equal(t, `null`, retained)
			},
		},
		{
			name: "data null is rejected if the selected message does not accept null",
			args: common.MqttPublishArgs{MessageId: "alarm", Data: common.Optional[any]{IsSet: true}},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.Error(t, err)
				require.Empty(t, retained)
			},
		},
		{
			name: "data null without messageId is validated against the messages of the topic",
			args: common.MqttPublishArgs{Topic: "sensors/null", Data: common.Optional[any]{IsSet: true}},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.NoError(t, err)
				require.Equal(t, `null`, retained)
			},
		},
		{
			name: "value null is an empty payload without validation",
			args: common.MqttPublishArgs{MessageId: "alarm", Value: []byte{}},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.NoError(t, err)
				require.Empty(t, retained)
			},
		},
		{
			name: "value empty string is an empty payload without validation",
			args: common.MqttPublishArgs{MessageId: "alarm", Value: []byte("")},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.NoError(t, err)
				require.Empty(t, retained)
			},
		},
		{
			name: "messageId on a topic without channel is an error",
			args: common.MqttPublishArgs{Topic: "$SYS/broker/version", MessageId: "alarm"},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.EqualError(t, err, "message 'alarm' not found in channel '$SYS/broker/version'")
			},
		},
		{
			name: "unknown messageId is an error without fallback",
			args: common.MqttPublishArgs{MessageId: "alrm"},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.EqualError(t, err, "message 'alrm' not found in channel 'sensors/events'")
				require.Empty(t, retained)
			},
		},
		{
			name: "selected message with null payload schema generates null",
			args: common.MqttPublishArgs{Topic: "sensors/null", MessageId: "nullable"},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.NoError(t, err)
				require.Equal(t, `null`, retained)
			},
		},
		{
			name: "payload is generated from a message of the topic without messageId, data and value",
			args: common.MqttPublishArgs{},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.NoError(t, err)
				require.Regexp(t, `^\{"kind":"(reading|alarm)"\}$`, retained)
			},
		},
		{
			name: "value is published as given",
			args: common.MqttPublishArgs{Value: []byte(`{"kind":"alarm"}`)},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kind":"alarm"}`, retained)
			},
		},
		{
			name: "value with messageId is published as given",
			args: common.MqttPublishArgs{MessageId: "alarm", Value: []byte(`{"kind":"reading"}`)},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kind":"reading"}`, retained)
			},
		},
		{
			name: "selected message without payload publishes an empty payload",
			args: common.MqttPublishArgs{Topic: "sensors/ping", MessageId: "ping"},
			test: func(t *testing.T, retained string, result *common.MqttPublishResult, err error) {
				require.NoError(t, err)
				require.Empty(t, retained)
				require.Empty(t, result.Value)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(11)

			app := runtime.New(&static.Config{}, &dynamictest.Reader{})
			e := enginetest.NewEngine()
			info, err := app.Mqtt.Add(getConfig(newConfig()), e)
			require.NoError(t, err)

			args := tc.args
			if args.Topic == "" {
				args.Topic = "sensors/events"
			}
			args.Retain = true
			result, err := engine.NewMqttClient(app).Publish(&args)

			retained := ""
			if topic, ok := info.Topic(args.Topic); ok && topic.Retained != nil {
				retained = string(topic.Retained.Data)
			}
			tc.test(t, retained, result, err)
		})
	}
}

func TestMqttClient_Publish_ConfigRemovedWhileScriptIsRunning(t *testing.T) {
	logrus.SetOutput(io.Discard)

	// a cluster whose last file was removed keeps its store but has no config snapshot (nil)
	newApp := func() *runtime.App {
		cfg := asyncapi3test.NewConfig(
			asyncapi3test.WithInfo("foo", "", ""),
			asyncapi3test.WithChannel("sensors", asyncapi3test.WithMessage("msg")),
		)
		st := store.New(cfg, enginetest.NewEngine(), &eventstest.Handler{}, monitor.NewMqtt())
		return runtimetest.NewApp(runtimetest.WithMqttInfo("foo", runtime.NewMqttInfo(nil, st)))
	}

	t.Run("cluster given returns not found", func(t *testing.T) {
		require.NotPanics(t, func() {
			_, err := engine.NewMqttClient(newApp()).Publish(&common.MqttPublishArgs{Cluster: "foo", Topic: "sensors"})

			require.EqualError(t, err, "mqtt API 'foo' not found")
		})
	})

	t.Run("cluster searched by topic skips it", func(t *testing.T) {
		require.NotPanics(t, func() {
			_, err := engine.NewMqttClient(newApp()).Publish(&common.MqttPublishArgs{Topic: "sensors"})

			require.EqualError(t, err, "mqtt topic 'sensors' not found")
		})
	})
}

func TestMqttClient_Publish_NoApiOrTopic(t *testing.T) {
	logrus.SetOutput(io.Discard)

	t.Run("no API loaded returns not found", func(t *testing.T) {
		require.NotPanics(t, func() {
			_, err := engine.NewMqttClient(runtimetest.NewApp()).Publish(&common.MqttPublishArgs{})

			require.EqualError(t, err, "mqtt topic '' not found")
		})
	})

	// the $SYS topics of every API make a topic always ambiguous without a cluster
	t.Run("API without topics of its own is ambiguous", func(t *testing.T) {
		app := runtimetest.NewApp(runtimetest.WithMqtt(asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""))))

		_, err := engine.NewMqttClient(app).Publish(&common.MqttPublishArgs{})

		require.EqualError(t, err, "ambiguous topic . Specify the cluster")
	})
}

func TestMqttClient_Publish_Api(t *testing.T) {
	logrus.SetOutput(io.Discard)

	newApp := func() *runtime.App {
		return runtimetest.NewApp(runtimetest.WithMqtt(asyncapi3test.NewConfig(
			asyncapi3test.WithInfo("foo", "", ""),
			asyncapi3test.AddChannel("sensors", asyncapi3test.NewChannel(
				asyncapi3test.WithChannelAddress("sensors"),
				asyncapi3test.UseMessage("msg", &asyncapi3.MessageRef{Value: asyncapi3test.NewMessage()}),
			)),
		)))
	}

	testcases := []struct {
		name string
		args common.MqttPublishArgs
		err  string
	}{
		{name: "topic only", args: common.MqttPublishArgs{Topic: "sensors"}},
		{name: "cluster given", args: common.MqttPublishArgs{Cluster: "foo", Topic: "sensors"}},
		{name: "unknown cluster", args: common.MqttPublishArgs{Cluster: "bar", Topic: "sensors"}, err: "mqtt API 'bar' not found"},
		{name: "unknown topic with cluster", args: common.MqttPublishArgs{Cluster: "foo", Topic: "x"}, err: "mqtt topic 'x' not found"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.NewMqttClient(newApp()).Publish(&tc.args)

			if tc.err != "" {
				require.EqualError(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "foo", result.Api)
		})
	}
}

func TestMqttClient_Publish_RetriesTopicNotFound(t *testing.T) {
	logrus.SetOutput(io.Discard)
	app := runtimetest.NewApp(runtimetest.WithMqtt(asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""))))
	// set after the app is created, adding an API changes the log level
	logrus.SetLevel(logrus.DebugLevel)
	hook := logtest.NewGlobal()

	_, err := engine.NewMqttClient(app).Publish(&common.MqttPublishArgs{
		Cluster: "foo",
		Topic:   "x",
		Retry:   common.RetryArgs{Retries: 2, InitialRetryTime: time.Millisecond, MaxRetryTime: time.Second, Factor: 2},
	})

	require.EqualError(t, err, "mqtt topic 'x' not found")
	require.Contains(t, getMessages(hook), "mqtt topic 'x' not found. Retry in 1ms")
}
