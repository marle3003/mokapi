package kafka_test

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"mokapi/kafka/createTopics"
	"testing"
)

func TestCreateTopics_Request(t *testing.T) {
	testcases := []struct {
		name  string
		input []byte
		test  func(t *testing.T, r *kafka.Request, unread int, err error)
	}{
		{
			name: "v4",
			input: []byte{
				0x0, 0x0, 0x0, 0xd1, // length: 42
				0x0, 0x13, // api key (create topics)
				0x0, 0x4, // version
				0x0, 0x0, 0x0, 0x3, // correlation ID: 3
				0x0, 0x7, 0x72, 0x64, 0x6b, 0x61, 0x66, 0x6b, 0x61, // client ID: rdkafka
				0x0, 0x0, 0x0, 0x1, // topics length
				0x0, 0x3, 0x66, 0x6f, 0x6f, // topic: foo
				0x0, 0x0, 0x0, 0x2, // num partition
				0x0, 0x1, // replication factor
				0x0, 0x0, 0x0, 0x0, // replication assignments length
				0x0, 0x0, 0x0, 0x0, // configs length
				0x0, 0x0, 0xea, 0x60, // timeout 60000
				0x0, // validate only false
			},
			test: func(t *testing.T, r *kafka.Request, unread int, err error) {
				require.NoError(t, err)
				require.Equal(t, 0, unread, "should read all bytes")
				m := r.Message.(*createTopics.Request)
				topic := m.Topics[0]
				require.Equal(t, "foo", topic.Name)
				require.Equal(t, int32(2), topic.NumPartitions)
				require.Equal(t, int16(1), topic.ReplicationFactor)
				require.Len(t, topic.Assignments, 0)
				require.Len(t, topic.Configs, 0)

				require.Equal(t, int32(60000), m.TimeoutMs)
				require.Equal(t, false, m.ValidateOnly)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r := &kafka.Request{}
			reader := bytes.NewReader(tc.input)
			err := r.Read(reader)
			unread := reader.Len()
			tc.test(t, r, unread, err)
		})
	}
}
