package store_test

import (
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/createTopics"
	"mokapi/kafka/kafkatest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateTopic(t *testing.T) {
	defaultConfig := asyncapi3test.NewConfig(
		asyncapi3test.WithServer("foo", "kafka", "127.0.0.1:9092"),
		asyncapi3test.WithServer("bar", "kafka", "127.0.0.1:9093"),
		asyncapi3test.WithChannel("already.exists"),
		asyncapi3test.WithChannel("orders.{region}.{eventType}",
			asyncapi3test.WithParameter(
				"region", &asyncapi3.Parameter{
					Enum: []string{"us", "eu"},
				},
			),
			asyncapi3test.WithParameter(
				"eventType", &asyncapi3.Parameter{
					Enum: []string{"created", "processed"},
				},
			),
		),
		asyncapi3test.WithChannel("products.{region}",
			asyncapi3test.AssignToServer("foo"),
			asyncapi3test.WithParameter(
				"region", &asyncapi3.Parameter{
					Enum: []string{"us", "eu"},
				},
			),
		),
	)

	testcases := []struct {
		name string
		cfg  *asyncapi3.Config
		test func(t *testing.T, s *store.Store)
	}{
		{
			name: "valid topic name and num partitions",
			test: func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &createTopics.Request{
					Topics: []createTopics.Topic{
						{
							Name:          "orders.eu.created",
							NumPartitions: 3,
						},
					},
				}))
				res, ok := rr.Message.(*createTopics.Response)
				require.True(t, ok)
				require.Len(t, res.Topics, 1)
				require.Equal(t, "orders.eu.created", res.Topics[0].Name)
				require.Equal(t, kafka.None, res.Topics[0].ErrorCode)
				require.Equal(t, int32(3), res.Topics[0].NumPartitions)
			},
		},
		{
			name: "valid topic name no num partitions should use default number",
			test: func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &createTopics.Request{
					Topics: []createTopics.Topic{
						{
							Name:          "orders.us.created",
							NumPartitions: -1,
						},
					},
				}))
				res, ok := rr.Message.(*createTopics.Response)
				require.True(t, ok)
				require.Len(t, res.Topics, 1)
				require.Equal(t, "orders.us.created", res.Topics[0].Name)
				require.Equal(t, kafka.None, res.Topics[0].ErrorCode)
				require.Equal(t, int32(1), res.Topics[0].NumPartitions)
			},
		},
		{
			name: "invalid topic name",
			test: func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &createTopics.Request{
					Topics: []createTopics.Topic{
						{
							Name:          "test",
							NumPartitions: -1,
						},
					},
				}))
				res, ok := rr.Message.(*createTopics.Response)
				require.True(t, ok)
				require.Len(t, res.Topics, 1)
				require.Equal(t, "test", res.Topics[0].Name)
				require.Equal(t, kafka.PolicyViolation, res.Topics[0].ErrorCode)
				require.Equal(t, `topic 'test' does not match any channel available on server 'foo':
- does not match channel address expression 'orders.{region}.{eventType}'
- does not match channel address expression 'products.{region}'`, res.Topics[0].ErrorMessage)
			},
		},
		{
			name: "create twice",
			test: func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &createTopics.Request{
					Topics: []createTopics.Topic{
						{
							Name: "orders.us.created",
						},
						{
							Name: "orders.us.created",
						},
					},
				}))
				res, ok := rr.Message.(*createTopics.Response)
				require.True(t, ok)
				require.Len(t, res.Topics, 2)
				require.Equal(t, "orders.us.created", res.Topics[0].Name)
				require.Equal(t, kafka.None, res.Topics[0].ErrorCode)
				require.Equal(t, int32(0), res.Topics[0].NumPartitions)
				require.Equal(t, "orders.us.created", res.Topics[1].Name)
				require.Equal(t, kafka.TopicAlreadyExists, res.Topics[1].ErrorCode)
				require.Equal(t, "", res.Topics[1].ErrorMessage)
			},
		},
		{
			name: "already exists",
			test: func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &createTopics.Request{
					Topics: []createTopics.Topic{
						{
							Name: "already.exists",
						},
					},
				}))
				res, ok := rr.Message.(*createTopics.Response)
				require.True(t, ok)
				require.Len(t, res.Topics, 1)
				require.Equal(t, "already.exists", res.Topics[0].Name)
				require.Equal(t, kafka.TopicAlreadyExists, res.Topics[0].ErrorCode)
				require.Equal(t, "", res.Topics[0].ErrorMessage)
			},
		},
		{
			name: "channel with parameter is not enabled for broker",
			test: func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				r := kafkatest.NewRequest("kafkatest", 3, &createTopics.Request{
					Topics: []createTopics.Topic{
						{
							Name: "products.us",
						},
					},
				})
				r.Host = "127.0.0.1:9093"
				s.ServeMessage(rr, r)
				res, ok := rr.Message.(*createTopics.Response)
				require.True(t, ok)
				require.Len(t, res.Topics, 1)
				require.Equal(t, "products.us", res.Topics[0].Name)
				require.Equal(t, kafka.PolicyViolation, res.Topics[0].ErrorCode)
				require.Equal(t, `topic 'products.us' does not match any channel available on server 'bar':
- channel 'products.{region}' is not available for Kafka protocol
- does not match channel address expression 'orders.{region}.{eventType}'`, res.Topics[0].ErrorMessage)
			},
		},
		{
			name: "channel is not for Kafka available",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithServer("foo", "mqtt", "127.0.0.1:1883"),
				asyncapi3test.WithChannel("products.{region}",
					asyncapi3test.AssignToServer("foo"),
					asyncapi3test.WithParameter(
						"region", &asyncapi3.Parameter{
							Enum: []string{"us", "eu"},
						},
					),
				),
			),
			test: func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				r := kafkatest.NewRequest("kafkatest", 3, &createTopics.Request{
					Topics: []createTopics.Topic{
						{
							Name: "products.eu",
						},
					},
				})
				s.ServeMessage(rr, r)
				res, ok := rr.Message.(*createTopics.Response)
				require.True(t, ok)
				require.Len(t, res.Topics, 1)
				require.Equal(t, "products.eu", res.Topics[0].Name)
				require.Equal(t, kafka.PolicyViolation, res.Topics[0].ErrorCode)
				require.Equal(t, `channel 'products.{region}' is not available for Kafka protocol`, res.Topics[0].ErrorMessage)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.cfg
			if cfg == nil {
				cfg = defaultConfig
			}
			s := store.New(cfg, enginetest.NewEngine(), &eventstest.Handler{}, monitor.NewKafka())
			defer s.Close()

			tc.test(t, s)
		})
	}
}
