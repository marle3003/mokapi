package offsetCommit_test

import (
	"bytes"
	"encoding/binary"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/offsetCommit"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.OffsetCommit]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(9), reg.MaxVersion)
}

func TestRequest(t *testing.T) {
	kafkatest.TestRequest(t, 7, &offsetCommit.Request{
		GroupId:         "foo",
		GenerationId:    1,
		MemberId:        "m1",
		GroupInstanceId: "g1",
		Topics: []offsetCommit.Topic{
			{
				Name: "bar",
				Partitions: []offsetCommit.Partition{
					{
						Index:       1,
						Offset:      12,
						LeaderEpoch: 0,
						Metadata:    "metadata",
					},
				},
			},
		},
	})

	kafkatest.TestRequest(t, 9, &offsetCommit.Request{
		GroupId:         "foo",
		GenerationId:    1,
		MemberId:        "m1",
		GroupInstanceId: "g1",
		Topics: []offsetCommit.Topic{
			{
				Name: "bar",
				Partitions: []offsetCommit.Partition{
					{
						Index:       1,
						Offset:      12,
						LeaderEpoch: 0,
						Metadata:    "metadata",
					},
				},
			},
		},
	})

	b := kafkatest.WriteRequest(t, 9, 123, "me", &offsetCommit.Request{
		GroupId:         "foo",
		GenerationId:    1,
		MemberId:        "m1",
		GroupInstanceId: "g1",
		Topics: []offsetCommit.Topic{
			{
				Name: "bar",
				Partitions: []offsetCommit.Partition{
					{
						Index:       1,
						Offset:      12,
						LeaderEpoch: 0,
						Metadata:    "metadata",
					},
				},
			},
		},
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(61))                 // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.OffsetCommit)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(9))                  // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))                // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))                  // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))              // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))                   // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int8(4))            // GroupId length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo"))      // GroupId
	_ = binary.Write(expected, binary.BigEndian, int32(1))           // GenerationId
	_ = binary.Write(expected, binary.BigEndian, int8(3))            // MemberId length
	_ = binary.Write(expected, binary.BigEndian, []byte("m1"))       // MemberId
	_ = binary.Write(expected, binary.BigEndian, int8(3))            // GroupInstanceId length
	_ = binary.Write(expected, binary.BigEndian, []byte("g1"))       // GroupInstanceId
	_ = binary.Write(expected, binary.BigEndian, int8(2))            // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))            // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("bar"))      // Name
	_ = binary.Write(expected, binary.BigEndian, int8(2))            // Partitions length
	_ = binary.Write(expected, binary.BigEndian, int32(1))           // Partition Index
	_ = binary.Write(expected, binary.BigEndian, int64(12))          // Offset
	_ = binary.Write(expected, binary.BigEndian, int32(0))           // LeaderEpoch
	_ = binary.Write(expected, binary.BigEndian, int8(9))            // Metadata length
	_ = binary.Write(expected, binary.BigEndian, []byte("metadata")) // Metadata
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // Partitions tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // tag buffer
	require.Equal(t, expected.Bytes(), b)
}

func TestResponse(t *testing.T) {
	kafkatest.TestResponse(t, 7, &offsetCommit.Response{
		ThrottleTimeMs: 123,
		Topics: []offsetCommit.ResponseTopic{
			{
				Name: "foo",
				Partitions: []offsetCommit.ResponsePartition{
					{
						Index:     1,
						ErrorCode: 0,
					},
				},
			},
		},
	})

	kafkatest.TestResponse(t, 9, &offsetCommit.Response{
		ThrottleTimeMs: 123,
		Topics: []offsetCommit.ResponseTopic{
			{
				Name: "foo",
				Partitions: []offsetCommit.ResponsePartition{
					{
						Index:     1,
						ErrorCode: 0,
					},
				},
			},
		},
	})

	b := kafkatest.WriteResponse(t, 9, 123, &offsetCommit.Response{
		ThrottleTimeMs: 123,
		Topics: []offsetCommit.ResponseTopic{
			{
				Name: "foo",
				Partitions: []offsetCommit.ResponsePartition{
					{
						Index:     1,
						ErrorCode: 0,
					},
				},
			},
		},
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(24))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(123))    // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // Name
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Partitions length
	_ = binary.Write(expected, binary.BigEndian, int32(1))      // Partitions Index
	_ = binary.Write(expected, binary.BigEndian, int16(0))      // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // Partitions tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	require.Equal(t, expected.Bytes(), b)
}
