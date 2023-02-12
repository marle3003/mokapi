package store_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/metaData"
	"strings"
	"testing"
)

func TestMetadata(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, s *store.Store)
	}{
		{
			"default",
			func(t *testing.T, s *store.Store) {
				//b.Apply(asyncapitest.NewConfig(asyncapitest.WithChannel("foo", asyncapitest.WithChannelKafka("partitions", "1"))))
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithServer("", "kafka", "127.0.0.1:9092"),
					asyncapitest.WithChannel("foo"),
				))
				rr := kafkatest.NewRecorder()
				r := kafkatest.NewRequest("kafkatest", 4, &metaData.Request{})
				s.ServeMessage(rr, r)

				res, ok := rr.Message.(*metaData.Response)
				require.True(t, ok)

				// controller
				require.Equal(t, int32(0), res.ControllerId)

				// brokers
				require.Len(t, res.Brokers, 1)
				require.Equal(t, int32(0), res.Brokers[0].NodeId)
				require.Equal(t, "127.0.0.1", res.Brokers[0].Host)
				require.Equal(t, int32(9092), res.Brokers[0].Port)
				require.Equal(t, "", res.Brokers[0].Rack)

				// topics
				require.Len(t, res.Topics, 1)
				require.Equal(t, "foo", res.Topics[0].Name)
				require.Equal(t, kafka.None, res.Topics[0].ErrorCode)
				require.Len(t, res.Topics[0].Partitions, 1)
				require.Equal(t, int32(0), res.Topics[0].Partitions[0].PartitionIndex)
				require.Equal(t, int32(0), res.Topics[0].Partitions[0].LeaderId) // default broker id is 0
				require.Len(t, res.Topics[0].Partitions[0].ReplicaNodes, 1)
				require.Equal(t, int32(0), res.Topics[0].Partitions[0].ReplicaNodes[0])
				require.Len(t, res.Topics[0].Partitions[0].IsrNodes, 1)
				require.Equal(t, int32(0), res.Topics[0].Partitions[0].IsrNodes[0])
				require.False(t, res.Topics[0].IsInternal)

				require.False(t, kafka.ClientFromContext(r).AllowAutoTopicCreation)
			},
		},
		{
			"with specific topic and two partitions",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithChannel("foo", asyncapitest.WithChannelKafka("partitions", "2")),
					asyncapitest.WithChannel("foo2", asyncapitest.WithChannelKafka("partitions", "1")),
				))

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr,
					kafkatest.NewRequest("kafkatest", 4,
						&metaData.Request{
							Topics: []metaData.TopicName{{Name: "foo"}, {Name: "bar"}},
						},
					),
				)

				res, ok := rr.Message.(*metaData.Response)
				require.True(t, ok)
				require.Len(t, res.Topics[0].Partitions, 2)
			},
		},
		{
			"with invalid topic",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithChannel("foo")))

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 4, &metaData.Request{
					Topics: []metaData.TopicName{{Name: "foo"}, {Name: "bar"}},
				}))

				res, ok := rr.Message.(*metaData.Response)
				require.True(t, ok)
				require.Len(t, res.Topics, 2)
				require.Equal(t, kafka.None, res.Topics[0].ErrorCode)
				require.Equal(t, kafka.UnknownTopicOrPartition, res.Topics[1].ErrorCode)
			},
		},
		{
			"create auto topic true",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithChannel("foo")))

				rr := kafkatest.NewRecorder()
				r := kafkatest.NewRequest("kafkatest", 4, &metaData.Request{
					AllowAutoTopicCreation: true,
				})
				s.ServeMessage(rr, r)

				require.True(t, kafka.ClientFromContext(r).AllowAutoTopicCreation)
			},
		},
		{
			"with invalid topic name",
			func(t *testing.T, s *store.Store) {
				for _, name := range []string{"", ".", "..", "event?", strings.Repeat("a", 250)} {
					testName := name
					if len(name) > 10 {
						testName = testName[:10] + "..."
					}
					t.Run(fmt.Sprintf("name %q", testName), func(t *testing.T) {
						rr := kafkatest.NewRecorder()
						s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 4, &metaData.Request{
							Topics: []metaData.TopicName{{Name: name}},
						}))

						res, ok := rr.Message.(*metaData.Response)
						require.True(t, ok)
						require.Len(t, res.Topics, 1)
						require.Equal(t, kafka.InvalidTopic, res.Topics[0].ErrorCode)
					})
				}
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := store.New(asyncapitest.NewConfig(), enginetest.NewEngine())
			defer s.Close()
			tc.fn(t, s)
		})
	}
}
