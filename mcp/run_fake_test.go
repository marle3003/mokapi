package mcp_test

import (
	"context"
	"mokapi/mcp"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_Run_Fake(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service)
	}{
		{
			name: "fake string",
			app:  runtimetest.NewApp(),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.fake({ type: 'string' })`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, "P8", r.Result)
			},
		},
		{
			name: "fake object",
			app:  runtimetest.NewApp(),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.fake({ type: 'object', properties: { foo: { type: 'string' } } })`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "P8"}, r.Result)
			},
		},
		{
			name: "fake from message payload",
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
						),
					),
				)
			}(),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `const topic = mokapi.getApi('foo').getTopic('channel-1')
const operation = topic.operations.find(x => x.action === 'send')
mokapi.fake(operation.messages[0].payload)
`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, map[string]any{}, r.Result)
				data := r.Result.(map[string]any)
				require.Equal(t, map[string]any{
					"foo": "P8",
				}, data)
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
