package store_test

import (
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTopic_Config_AfterUpdate(t *testing.T) {
	s := newStoreFromConfig(asyncapi3test.NewConfig(
		asyncapi3test.WithChannel("foo", asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 1})),
	))
	defer s.Close()
	before := s.Topic("foo").Config()
	require.Equal(t, 1, before.Bindings.Kafka.Partitions)

	s.Update(asyncapi3test.NewConfig(
		asyncapi3test.WithChannel("foo", asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 3})),
	))

	after := s.Topic("foo").Config()
	require.NotSame(t, before, after)
	require.Equal(t, 3, after.Bindings.Kafka.Partitions)
	require.Equal(t, 1, before.Bindings.Kafka.Partitions)
}
