package offset_test

import (
	"bytes"
	"encoding/binary"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/offset"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.Offset]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(8), reg.MaxVersion)
}

func TestRequest(t *testing.T) {
	kafkatest.TestRequest(t, 5, &offset.Request{
		ReplicaId:      1,
		IsolationLevel: 0,
		Topics: []offset.RequestTopic{
			{
				Name: "foo",
				Partitions: []offset.RequestPartition{
					{
						Index:       1,
						LeaderEpoch: 0,
						Timestamp:   1657010762684,
					},
				},
			},
		},
	})

	kafkatest.TestRequest(t, 8, &offset.Request{
		ReplicaId:      1,
		IsolationLevel: 0,
		Topics: []offset.RequestTopic{
			{
				Name: "foo",
				Partitions: []offset.RequestPartition{
					{
						Index:       1,
						LeaderEpoch: 0,
						Timestamp:   1657010762684,
					},
				},
			},
		},
	})

	b := kafkatest.WriteRequest(t, 8, 123, "me", &offset.Request{
		ReplicaId:      1,
		IsolationLevel: 0,
		Topics: []offset.RequestTopic{
			{
				Name: "foo",
				Partitions: []offset.RequestPartition{
					{
						Index:       1,
						LeaderEpoch: 0,
						Timestamp:   1657010762684,
					},
				},
			},
		},
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(43))           // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.Offset)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(8))            // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))          // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))            // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))        // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))             // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(1))             // ReplicaId
	_ = binary.Write(expected, binary.BigEndian, int8(0))              // IsolationLevel
	_ = binary.Write(expected, binary.BigEndian, int8(2))              // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))              // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo"))        // Name
	_ = binary.Write(expected, binary.BigEndian, int8(2))              // Partitions length
	_ = binary.Write(expected, binary.BigEndian, int32(1))             // Index
	_ = binary.Write(expected, binary.BigEndian, int32(0))             // LeaderEpoch
	_ = binary.Write(expected, binary.BigEndian, int64(1657010762684)) // Timestamp
	_ = binary.Write(expected, binary.BigEndian, int8(0))              // Partitions tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))              // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))              // tag buffer

	require.Equal(t, expected.Bytes(), b)
}

func TestResponse(t *testing.T) {
	kafkatest.TestResponse(t, 5, &offset.Response{
		ThrottleTimeMs: 0,
		Topics: []offset.ResponseTopic{
			{
				Name: "foo",
				Partitions: []offset.ResponsePartition{
					{
						Index:       1,
						ErrorCode:   0,
						Timestamp:   1657010762684,
						Offset:      0,
						LeaderEpoch: 0,
					},
				},
			},
		},
	})

	kafkatest.TestResponse(t, 8, &offset.Response{
		ThrottleTimeMs: 0,
		Topics: []offset.ResponseTopic{
			{
				Name: "foo",
				Partitions: []offset.ResponsePartition{
					{
						Index:       1,
						ErrorCode:   0,
						Timestamp:   1657010762684,
						Offset:      0,
						LeaderEpoch: 0,
					},
				},
			},
		},
	})

	b := kafkatest.WriteResponse(t, 8, 123, &offset.Response{
		ThrottleTimeMs: 123,
		Topics: []offset.ResponseTopic{
			{
				Name: "foo",
				Partitions: []offset.ResponsePartition{
					{
						Index:       1,
						ErrorCode:   0,
						Timestamp:   1657010762684,
						Offset:      0,
						LeaderEpoch: 0,
					},
				},
			},
		},
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(44))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(123))           // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int8(2))              // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))              // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo"))        // Name
	_ = binary.Write(expected, binary.BigEndian, int8(2))              // Partitions length
	_ = binary.Write(expected, binary.BigEndian, int32(1))             // Index
	_ = binary.Write(expected, binary.BigEndian, int16(0))             // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int64(1657010762684)) // Timestamp
	_ = binary.Write(expected, binary.BigEndian, int64(0))             // Offset
	_ = binary.Write(expected, binary.BigEndian, int32(0))             // LeaderEpoch
	_ = binary.Write(expected, binary.BigEndian, int8(0))              // Partition tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))              // Topic tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))              // tag buffer
	require.Equal(t, expected.Bytes(), b)
}
