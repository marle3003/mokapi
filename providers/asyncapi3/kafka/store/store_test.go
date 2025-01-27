package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestStore(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			"empty",
			func(t *testing.T) {
				s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngine())
				defer s.Close()
				require.Equal(t, 0, len(s.Brokers()))
				require.Equal(t, 0, len(s.Topics()))
				require.Equal(t, 0, len(s.Groups()))
				require.Nil(t, s.Topic("foo"), "topic not exists")
			},
		},
		{
			"server",
			func(t *testing.T) {
				s := store.New(asyncapi3test.NewConfig(
					asyncapi3test.WithServer("foo", "kafka", "foo:9092"),
				), enginetest.NewEngine())
				defer s.Close()
				require.Equal(t, 1, len(s.Brokers()))
				require.Equal(t, 0, len(s.Topics()))
				require.Equal(t, 0, len(s.Groups()))
				b, ok := s.Broker(0)
				require.Equal(t, true, ok)
				require.Equal(t, "foo", b.Name)
			},
		},
		{
			"topic",
			func(t *testing.T) {
				s := store.New(asyncapi3test.NewConfig(
					asyncapi3test.WithChannel("foo"),
				), enginetest.NewEngine())
				defer s.Close()
				require.Equal(t, 0, len(s.Brokers()))
				require.Equal(t, 1, len(s.Topics()))
				require.Equal(t, 0, len(s.Groups()))
				topic := s.Topic("foo")
				require.NotNil(t, topic, "topic is not nil")
				require.Equal(t, "foo", topic.Name)
				require.Len(t, topic.Partitions, 1)
			},
		},
		{
			"create topic",
			func(t *testing.T) {
				s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngine())
				defer s.Close()
				topic, err := s.NewTopic("foo", asyncapi3test.NewChannel(), []*asyncapi3.Operation{})
				require.NoError(t, err)
				require.Equal(t, "foo", topic.Name)
				require.Equal(t, 1, len(topic.Partitions))
			},
		},
		{
			"create topic, already exists",
			func(t *testing.T) {
				s := store.New(asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo")), enginetest.NewEngine())
				defer s.Close()
				_, err := s.NewTopic("foo", asyncapi3test.NewChannel(), []*asyncapi3.Operation{})
				require.Error(t, err, "topic foo already exists")
			},
		},
		{
			"update topic add partition",
			func(t *testing.T) {
				s := store.New(asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo", asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 1}))), enginetest.NewEngine())
				defer s.Close()

				s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo", asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 2}))))

				require.Len(t, s.Topic("foo").Partitions, 2)
			},
		},
		{
			"update topic remove partition",
			func(t *testing.T) {
				s := store.New(asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo", asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 2}))), enginetest.NewEngine())
				defer s.Close()

				s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo", asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 1}))))

				require.Len(t, s.Topic("foo").Partitions, 1)
			},
		},
		{
			"update topic change schema",
			func(t *testing.T) {
				s := store.New(asyncapi3test.NewConfig(
					asyncapi3test.WithChannel("foo",
						asyncapi3test.WithMessage("foo",
							asyncapi3test.WithPayload(schematest.New("integer")),
							asyncapi3test.WithContentType("application/json"),
						),
					)),
					enginetest.NewEngine())
				defer s.Close()

				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithChannel("foo",
						asyncapi3test.WithMessage("foo",
							asyncapi3test.WithPayload(schematest.New("string")),
							asyncapi3test.WithContentType("application/json"),
						),
					),
				))

				_, _, err := s.Topic("foo").Partitions[0].Write(
					kafka.RecordBatch{Records: []*kafka.Record{
						{
							Value: kafka.NewBytes([]byte("123")),
						},
					}},
				)
				require.EqualError(t, err, "validation error: invalid message: found 1 error:\ninvalid type, expected string but got number\nschema path #/type")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.fn(t)
		})
	}
}
