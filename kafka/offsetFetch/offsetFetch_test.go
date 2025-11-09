package offsetFetch_test

import (
	"bytes"
	"encoding/binary"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/offsetFetch"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.OffsetFetch]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(9), reg.MaxVersion)
}

func TestRequest(t *testing.T) {
	kafkatest.TestRequest(t, 5, &offsetFetch.Request{
		GroupId: "foo",
		Topics: []offsetFetch.RequestTopic{
			{
				Name:             "foo",
				PartitionIndexes: []int32{0, 1},
			},
		},
		RequireStable: false,
	})

	kafkatest.TestRequest(t, 7, &offsetFetch.Request{
		GroupId: "foo",
		Topics: []offsetFetch.RequestTopic{
			{
				Name:             "foo",
				PartitionIndexes: []int32{0, 1},
			},
		},
		RequireStable: false,
	})

	b := kafkatest.WriteRequest(t, 7, 123, "me", &offsetFetch.Request{
		GroupId: "foo",
		Topics: []offsetFetch.RequestTopic{
			{
				Name:             "bar",
				PartitionIndexes: []int32{0, 1},
			},
		},
		RequireStable: false,
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(34))                // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.OffsetFetch)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(7))                 // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))               // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))                 // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))             // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))                  // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // GroupId length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // GroupId
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("bar")) // Name
	_ = binary.Write(expected, binary.BigEndian, int8(3))       // Partitions length
	_ = binary.Write(expected, binary.BigEndian, int32(0))      // Partition 0
	_ = binary.Write(expected, binary.BigEndian, int32(1))      // Partition 1
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // RequireStable
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	require.Equal(t, expected.Bytes(), b)

	b = kafkatest.WriteRequest(t, 9, 123, "me", &offsetFetch.Request{
		Groups: []offsetFetch.RequestGroup{
			{
				GroupId:     "foo",
				MemberId:    "m1",
				MemberEpoch: 0,
				Topics: []offsetFetch.RequestTopic{
					{
						Name:             "bar",
						PartitionIndexes: []int32{0, 1},
					},
				},
				TagFields: nil,
			},
		},
		RequireStable: false,
	})
	expected = new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(43))                // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.OffsetFetch)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(9))                 // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))               // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))                 // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))             // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))                  // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Groups length
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // GroupId length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // GroupId
	_ = binary.Write(expected, binary.BigEndian, int8(3))       // MemberId length
	_ = binary.Write(expected, binary.BigEndian, []byte("m1"))  // MemberId
	_ = binary.Write(expected, binary.BigEndian, int32(0))      // MemberEpoch
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("bar")) // Name
	_ = binary.Write(expected, binary.BigEndian, int8(3))       // Partitions length
	_ = binary.Write(expected, binary.BigEndian, int32(0))      // Partition 0
	_ = binary.Write(expected, binary.BigEndian, int32(1))      // Partition 1
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // Groups tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // RequireStable
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	require.Equal(t, expected.Bytes(), b)
}

func TestResponse(t *testing.T) {
	kafkatest.TestResponse(t, 5, &offsetFetch.Response{
		ThrottleTimeMs: 123,
		Topics: []offsetFetch.ResponseTopic{
			{
				Name: "",
				Partitions: []offsetFetch.Partition{
					{
						Index:                1,
						CommittedOffset:      12,
						CommittedLeaderEpoch: 0,
						Metadata:             "metadata",
						ErrorCode:            0,
					},
				},
			},
		},
		ErrorCode: 0,
	})

	kafkatest.TestResponse(t, 7, &offsetFetch.Response{
		ThrottleTimeMs: 123,
		Topics: []offsetFetch.ResponseTopic{
			{
				Name: "foo",
				Partitions: []offsetFetch.Partition{
					{
						Index:                1,
						CommittedOffset:      12,
						CommittedLeaderEpoch: 0,
						Metadata:             "metadata",
						ErrorCode:            0,
					},
				},
			},
		},
		ErrorCode: 0,
	})

	b := kafkatest.WriteResponse(t, 7, 123, &offsetFetch.Response{
		ThrottleTimeMs: 123,
		Topics: []offsetFetch.ResponseTopic{
			{
				Name: "foo",
				Partitions: []offsetFetch.Partition{
					{
						Index:                1,
						CommittedOffset:      12,
						CommittedLeaderEpoch: 0,
						Metadata:             "metadata",
						ErrorCode:            0,
					},
				},
			},
		},
		ErrorCode: 0,
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(47))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(123))         // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int8(2))            // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))            // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo"))      // Name
	_ = binary.Write(expected, binary.BigEndian, int8(2))            // Partitions length
	_ = binary.Write(expected, binary.BigEndian, int32(1))           // Partition Index
	_ = binary.Write(expected, binary.BigEndian, int64(12))          // CommittedOffset
	_ = binary.Write(expected, binary.BigEndian, int32(0))           // CommittedLeaderEpoch
	_ = binary.Write(expected, binary.BigEndian, int8(9))            // Metadata length
	_ = binary.Write(expected, binary.BigEndian, []byte("metadata")) // Metadata
	_ = binary.Write(expected, binary.BigEndian, int16(0))           // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // Partitions tag buffer
	_ = binary.Write(expected, binary.BigEndian, int16(0))           // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // tag buffer
	require.Equal(t, expected.Bytes(), b)

	b = kafkatest.WriteResponse(t, 9, 123, &offsetFetch.Response{
		ThrottleTimeMs: 123,
		Groups: []offsetFetch.ResponseGroup{
			{
				GroupId: "foo",
				Topics: []offsetFetch.ResponseTopic{
					{
						Name: "foo",
						Partitions: []offsetFetch.Partition{
							{
								Index:                1,
								CommittedOffset:      12,
								CommittedLeaderEpoch: 0,
								Metadata:             "metadata",
								ErrorCode:            0,
							},
						},
					},
				},
			},
		},
		ErrorCode: 0,
	})
	expected = new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(53))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(123))         // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int8(2))            // Groups length
	_ = binary.Write(expected, binary.BigEndian, int8(4))            // GroupId length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo"))      // GroupId
	_ = binary.Write(expected, binary.BigEndian, int8(2))            // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))            // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo"))      // Name
	_ = binary.Write(expected, binary.BigEndian, int8(2))            // Partitions length
	_ = binary.Write(expected, binary.BigEndian, int32(1))           // Partition Index
	_ = binary.Write(expected, binary.BigEndian, int64(12))          // CommittedOffset
	_ = binary.Write(expected, binary.BigEndian, int32(0))           // CommittedLeaderEpoch
	_ = binary.Write(expected, binary.BigEndian, int8(9))            // Metadata length
	_ = binary.Write(expected, binary.BigEndian, []byte("metadata")) // Metadata
	_ = binary.Write(expected, binary.BigEndian, int16(0))           // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // Partitions tag buffer
	_ = binary.Write(expected, binary.BigEndian, int16(0))           // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // Groups tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // tag buffer
	require.Equal(t, expected.Bytes(), b)
}
