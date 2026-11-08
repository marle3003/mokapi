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
	"mokapi/runtime"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/sirupsen/logrus"
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
