package mcp_test

import (
	"context"
	"mokapi/mcp"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_GetEvents(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service)
	}{
		{
			name: "Get Events",
			app: runtimetest.NewApp(
				runtimetest.WithKafka(asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "", ""),
					asyncapi3test.WithChannel("topic-1",
						asyncapi3test.WithMessage("msg",
							asyncapi3test.WithPayload(schematest.New("string")),
						),
					),
				)),
			),
			test: func(t *testing.T, s *mcp.Service) {
				_, err := s.ProduceKafkaMessage(context.Background(), mcp.ProduceKafkaMessageInput{
					APIName:   "foo",
					Topic:     "topic-1",
					Partition: 0,
					Key:       nil,
					Value:     "hello world",
				})
				require.NoError(t, err)

				r, err := s.GetEvents(context.Background(), mcp.GetEventsInput{
					APIName: "foo",
					Type:    "kafka",
				})
				require.NoError(t, err)
				require.Len(t, r.Events, 1)
				require.IsType(t, &store.KafkaMessageLog{}, r.Events[0].Data)
				require.Equal(t, "hello world", r.Events[0].Data.(*store.KafkaMessageLog).Message.Value)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(12345)

			s := mcp.NewService(tc.app)
			tc.test(t, s)
		})
	}
}
