package store_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
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

// A message without value (`messages: {order: }` in YAML) is a nil entry in the channel.
func TestTopic_ChannelWithEmptyMessage(t *testing.T) {
	cfg := asyncapi3test.NewConfig(
		asyncapi3test.WithChannel("foo", asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 1})),
	)
	cfg.Channels["foo"].Value.Messages = map[string]*asyncapi3.MessageRef{"order": nil}
	require.NoError(t, cfg.Parse(&dynamic.Config{Data: cfg}, &dynamictest.Reader{}))

	var s *store.Store
	require.NotPanics(t, func() { s = newStoreFromConfig(cfg) })
	defer s.Close()

	require.NotNil(t, s.Topic("foo"))
}

func TestStore_EmptyChannelNextToValidChannel(t *testing.T) {
	cfg := asyncapi3test.NewConfig(
		asyncapi3test.WithChannel("foo", asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 1})),
	)
	cfg.Channels["empty"] = nil
	require.NoError(t, cfg.Parse(&dynamic.Config{Data: cfg}, &dynamictest.Reader{}))

	s := newStoreFromConfig(cfg)
	defer s.Close()

	require.NotNil(t, s.Topic("foo"))
	require.Nil(t, s.Topic("empty"))
}
