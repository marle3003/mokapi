package mcp_test

import (
	"context"
	"mokapi/mcp"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_ProduceKafkaMessage(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service)
	}{
		{
			name: "Produce Kafka Message",
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
				r, err := s.ProduceKafkaMessage(context.Background(), mcp.ProduceKafkaMessageInput{
					APIName:   "foo",
					Topic:     "topic-1",
					Partition: 0,
					Key:       nil,
					Value:     "hello world",
				})
				require.NoError(t, err)
				require.Equal(t, int64(0), r.Offset)
				r, err = s.ProduceKafkaMessage(context.Background(), mcp.ProduceKafkaMessageInput{
					APIName:   "foo",
					Topic:     "topic-1",
					Partition: 0,
					Key:       nil,
					Value:     "hello world 2",
				})
				require.NoError(t, err)
				require.Equal(t, int64(1), r.Offset)
			},
		},
		{
			name: "Produce Kafka Message but topic does not exist",
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
					Topic:     "topic-2",
					Partition: 0,
					Key:       nil,
					Value:     "hello world",
				})
				require.EqualError(t, err, "kafka topic 'topic-2' not found")
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
