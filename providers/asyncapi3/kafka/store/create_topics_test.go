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
	s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngine(), &eventstest.Handler{}, monitor.NewKafka())
	defer s.Close()

	rr := kafkatest.NewRecorder()
	s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &createTopics.Request{
		Topics: []createTopics.Topic{
			{
				Name:              "test",
				NumPartitions:     2,
				ReplicationFactor: 1,
			},
		},
	}))

	res, ok := rr.Message.(*createTopics.Response)
	require.True(t, ok)
	require.Equal(t, "test", res.Topics[0].Name)
	require.Equal(t, kafka.None, res.Topics[0].ErrorCode)
}

func TestCreateTopic_AlreadyExists(t *testing.T) {
	s := store.New(asyncapi3test.NewConfig(asyncapi3test.AddChannel("test", &asyncapi3.Channel{})), enginetest.NewEngine(), &eventstest.Handler{}, monitor.NewKafka())
	defer s.Close()

	rr := kafkatest.NewRecorder()
	s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &createTopics.Request{
		Topics: []createTopics.Topic{
			{
				Name:              "test",
				NumPartitions:     2,
				ReplicationFactor: 1,
			},
		},
	}))

	res, ok := rr.Message.(*createTopics.Response)
	require.True(t, ok)
	require.Equal(t, "test", res.Topics[0].Name)
	require.Equal(t, kafka.TopicAlreadyExists, res.Topics[0].ErrorCode)
}
