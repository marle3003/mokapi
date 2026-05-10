package mcp_test

import (
	"context"
	"mokapi/kafka"
	"mokapi/mcp"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/generator"
	jsonSchema "mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestService_Run_Kafka(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service)
	}{
		{
			name: "get Kafka APIs",
			app: runtimetest.NewKafkaApp(
				asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "", ""),
				),
				asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("bar", "", ""),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApis()`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, []mcp.ApiSummary{
					{Name: "bar", Type: "kafka"},
					{Name: "foo", Type: "kafka"},
				}, r.Result)
			},
		},
		{
			name: "get Kafka API",
			app: runtimetest.NewKafkaApp(
				asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "", ""),
					asyncapi3test.WithServer("bar", "kafka", "foo.bar", asyncapi3test.WithServerDescription("server description")),
				),
				asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("bar", "", ""),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo')`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, &mcp.Kafka{}, r.Result)
				kafka := r.Result.(*mcp.Kafka)
				require.Equal(t, "foo", kafka.Name)
				require.Equal(t, "kafka", kafka.Type)
				require.Equal(t, []mcp.Broker{{Name: "bar", Host: "foo.bar", Description: "server description"}}, kafka.Brokers)
			},
		},
		{
			name: "get topics",
			app: runtimetest.NewKafkaApp(
				asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "", ""),
					asyncapi3test.WithChannel("channel-1",
						asyncapi3test.WithChannelTitle("title-1"),
						asyncapi3test.WithChannelSummary("channel-1 summary"),
					),
					asyncapi3test.WithChannel("channel-2",
						asyncapi3test.WithChannelTitle("title-2"),
					),
				),
				asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("bar", "", ""),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo').getTopics()`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, []mcp.TopicSummary{}, r.Result)
				topics := r.Result.([]mcp.TopicSummary)
				require.Len(t, topics, 2)
				require.Equal(t, "channel-1", topics[0].Name)
				require.Equal(t, "title-1", topics[0].Title)
				require.Equal(t, "channel-1 summary", topics[0].Summary)
				require.Equal(t, "channel-2", topics[1].Name)
				require.Equal(t, "title-2", topics[1].Title)
				require.Equal(t, "", topics[1].Summary)
			},
		},
		{
			name: "get topic",
			app: func() *runtime.App {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithMessageName("msg-name-1"),
					asyncapi3test.WithMessageTitle("msg-title-1"),
					asyncapi3test.WithMessageSummary("msg-summary-1"),
					asyncapi3test.WithMessageDescription("msg-description-1"),
					asyncapi3test.WithContentType("application/json"),
					asyncapi3test.WithPayload(
						schematest.New("object",
							schematest.WithProperty("foo", schematest.New("string")),
						),
					),
				)

				ch := asyncapi3test.NewChannel(
					asyncapi3test.WithChannelTitle("title-1"),
					asyncapi3test.WithChannelSummary("channel-1 summary"),
					asyncapi3test.WithChannelDescription("description"),
					asyncapi3test.UseMessage("foo", &asyncapi3.MessageRef{Value: msg}),
				)

				return runtimetest.NewKafkaApp(
					asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "", ""),
						asyncapi3test.AddChannel("channel-1", ch),
						asyncapi3test.WithOperation("publish",
							asyncapi3test.WithOperationAction("send"),
							asyncapi3test.WithOperationTitle("op-title-1"),
							asyncapi3test.WithOperationSummary("op-summary-1"),
							asyncapi3test.WithOperationDescription("op-description-1"),
							asyncapi3test.WithOperationChannel(ch),
							asyncapi3test.UseOperationMessage(msg),
						),
						asyncapi3test.WithOperation("consume",
							asyncapi3test.WithOperationAction("receive"),
							asyncapi3test.WithOperationChannel(ch),
							asyncapi3test.UseOperationMessage(msg),
						),
					),
				)
			}(),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo').getTopic('channel-1')`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, mcp.Topic{}, r.Result)
				topic := r.Result.(mcp.Topic)
				require.Equal(t, "channel-1", topic.Name)
				require.Equal(t, "title-1", topic.Title)
				require.Equal(t, "channel-1 summary", topic.Summary)
				require.Equal(t, "description", topic.Description)
				require.Len(t, topic.Operations, 2)

				require.Equal(t, "receive", topic.Operations[0].Action)

				require.Equal(t, "send", topic.Operations[1].Action)
				require.Equal(t, "op-title-1", topic.Operations[1].Title)
				require.Equal(t, "op-summary-1", topic.Operations[1].Summary)
				require.Equal(t, "op-description-1", topic.Operations[1].Description)
				require.Len(t, topic.Operations[1].Messages, 1)
				require.Equal(t, "msg-name-1", topic.Operations[1].Messages[0].Name)
				require.Equal(t, "msg-title-1", topic.Operations[1].Messages[0].Title)
				require.Equal(t, "msg-summary-1", topic.Operations[1].Messages[0].Summary)
				require.Equal(t, "msg-description-1", topic.Operations[1].Messages[0].Description)
				require.Equal(t, "application/json", topic.Operations[1].Messages[0].ContentType)
				require.IsType(t, &jsonSchema.Schema{}, topic.Operations[1].Messages[0].Payload)
				payload := topic.Operations[1].Messages[0].Payload.(*jsonSchema.Schema)
				require.Equal(t, "object", payload.Type.String())
			},
		},
		{
			name: "consume from topic",
			app: func() *runtime.App {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithMessageName("msg-name-1"),
					asyncapi3test.WithMessageTitle("msg-title-1"),
					asyncapi3test.WithMessageSummary("msg-summary-1"),
					asyncapi3test.WithMessageDescription("msg-description-1"),
					asyncapi3test.WithContentType("application/json"),
					asyncapi3test.WithPayload(
						schematest.New("object",
							schematest.WithProperty("foo", schematest.New("string")),
						),
					),
				)

				ch := asyncapi3test.NewChannel(
					asyncapi3test.WithChannelTitle("title-1"),
					asyncapi3test.WithChannelSummary("channel-1 summary"),
					asyncapi3test.WithChannelDescription("description"),
					asyncapi3test.UseMessage("foo", &asyncapi3.MessageRef{Value: msg}),
				)

				app := runtimetest.NewKafkaApp(
					asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "", ""),
						asyncapi3test.AddChannel("channel-1", ch),
						asyncapi3test.WithOperation("publish",
							asyncapi3test.WithOperationAction("send"),
							asyncapi3test.WithOperationTitle("op-title-1"),
							asyncapi3test.WithOperationSummary("op-summary-1"),
							asyncapi3test.WithOperationDescription("op-description-1"),
							asyncapi3test.WithOperationChannel(ch),
							asyncapi3test.UseOperationMessage(msg),
						),
						asyncapi3test.WithOperation("consume",
							asyncapi3test.WithOperationAction("receive"),
							asyncapi3test.WithOperationChannel(ch),
							asyncapi3test.UseOperationMessage(msg),
						),
					),
				)

				_, err := app.Kafka.Get("foo").Store.Topic("channel-1").Partition(0).Write(
					kafka.RecordBatch{
						Records: []*kafka.Record{
							{
								Offset:  0,
								Time:    time.Time{},
								Key:     kafka.NewBytes([]byte("foo")),
								Value:   kafka.NewBytes([]byte(`{"foo":"bar"}`)),
								Headers: nil,
							},
						},
					},
				)
				require.NoError(t, err)

				return app
			}(),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo').getTopic('channel-1').consume(0, 0, 10)`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, []mcp.KafkaRecord{}, r.Result)
				records := r.Result.([]mcp.KafkaRecord)
				require.Len(t, records, 1)
				require.Equal(t, int64(0), records[0].Offset)
				require.Equal(t, "foo", records[0].Key)
				require.Equal(t, `{"foo":"bar"}`, records[0].Value)

				r, err = s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `const topic = mokapi.getApi('foo').getTopic('channel-1')
const p0 = topic.partitions.find(p => p.index === 0);
let lastMessage = null
if (p0 && p0.offset > 0) {
    lastMessage = topic.consume(0, p0.offset - 1, 1);
}
lastMessage
`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, []mcp.KafkaRecord{}, r.Result)
				records = r.Result.([]mcp.KafkaRecord)
				require.Len(t, records, 1)
				require.Equal(t, int64(0), records[0].Offset)
				require.Equal(t, "foo", records[0].Key)
				require.Equal(t, `{"foo":"bar"}`, records[0].Value)
			},
		},
		{
			name: "produce into topic",
			app: func() *runtime.App {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithMessageName("msg-name-1"),
					asyncapi3test.WithMessageTitle("msg-title-1"),
					asyncapi3test.WithMessageSummary("msg-summary-1"),
					asyncapi3test.WithMessageDescription("msg-description-1"),
					asyncapi3test.WithContentType("application/json"),
					asyncapi3test.WithPayload(
						schematest.New("object",
							schematest.WithProperty("foo", schematest.New("string")),
						),
					),
				)

				ch := asyncapi3test.NewChannel(
					asyncapi3test.WithChannelTitle("title-1"),
					asyncapi3test.WithChannelSummary("channel-1 summary"),
					asyncapi3test.WithChannelDescription("description"),
					asyncapi3test.UseMessage("foo", &asyncapi3.MessageRef{Value: msg}),
				)

				return runtimetest.NewKafkaApp(
					asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "", ""),
						asyncapi3test.AddChannel("channel-1", ch),
						asyncapi3test.WithOperation("publish",
							asyncapi3test.WithOperationAction("send"),
							asyncapi3test.WithOperationTitle("op-title-1"),
							asyncapi3test.WithOperationSummary("op-summary-1"),
							asyncapi3test.WithOperationDescription("op-description-1"),
							asyncapi3test.WithOperationChannel(ch),
							asyncapi3test.UseOperationMessage(msg),
						),
						asyncapi3test.WithOperation("consume",
							asyncapi3test.WithOperationAction("receive"),
							asyncapi3test.WithOperationChannel(ch),
							asyncapi3test.UseOperationMessage(msg),
						),
					),
				)
			}(),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `const t = mokapi.getApi('foo').getTopic('channel-1')
t.produce(0, {foo: 'bar'}, 'foo');
t`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, mcp.Topic{}, r.Result)
				topic := r.Result.(mcp.Topic)
				require.Len(t, topic.Partitions, 1)
				require.Equal(t, int64(1), topic.Partitions[0].Offset)
			},
		},
		{
			name: "produce into topic but invalid message",
			app: func() *runtime.App {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithMessageName("msg-name-1"),
					asyncapi3test.WithMessageTitle("msg-title-1"),
					asyncapi3test.WithMessageSummary("msg-summary-1"),
					asyncapi3test.WithMessageDescription("msg-description-1"),
					asyncapi3test.WithContentType("application/json"),
					asyncapi3test.WithPayload(
						schematest.New("object",
							schematest.WithProperty("foo", schematest.New("string")),
						),
					),
				)

				ch := asyncapi3test.NewChannel(
					asyncapi3test.WithChannelTitle("title-1"),
					asyncapi3test.WithChannelSummary("channel-1 summary"),
					asyncapi3test.WithChannelDescription("description"),
					asyncapi3test.UseMessage("foo", &asyncapi3.MessageRef{Value: msg}),
				)

				return runtimetest.NewKafkaApp(
					asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "", ""),
						asyncapi3test.AddChannel("channel-1", ch),
						asyncapi3test.WithOperation("publish",
							asyncapi3test.WithOperationAction("send"),
							asyncapi3test.WithOperationTitle("op-title-1"),
							asyncapi3test.WithOperationSummary("op-summary-1"),
							asyncapi3test.WithOperationDescription("op-description-1"),
							asyncapi3test.WithOperationChannel(ch),
							asyncapi3test.UseOperationMessage(msg),
						),
						asyncapi3test.WithOperation("consume",
							asyncapi3test.WithOperationAction("receive"),
							asyncapi3test.WithOperationChannel(ch),
							asyncapi3test.UseOperationMessage(msg),
						),
					),
				)
			}(),
			test: func(t *testing.T, s *mcp.Service) {
				_, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `const t = mokapi.getApi('foo').getTopic('channel-1')
t.produce(0, { foo: 123 }, 'foo');
t`,
					},
				)
				require.EqualError(t, err, `no matching message configuration found for the given value: {"foo":123}
hint:
encoding data to 'application/json' failed: error count 1:
	- #/foo/type: invalid type, expected string but got integer
`)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(123456)

			s := mcp.NewService(tc.app)
			tc.test(t, s)
		})
	}
}
