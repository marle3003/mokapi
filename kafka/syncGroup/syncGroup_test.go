package syncGroup_test

import (
	"bytes"
	"encoding/binary"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/syncGroup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.SyncGroup]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(5), reg.MaxVersion)
}

func TestRequest(t *testing.T) {
	kafkatest.TestRequest(t, 3, &syncGroup.Request{
		GroupId:         "foo",
		GenerationId:    1,
		MemberId:        "m1",
		GroupInstanceId: "g1",
		GroupAssignments: []syncGroup.GroupAssignment{
			{
				MemberId:   "m2",
				Assignment: []byte("assign"),
			},
		},
	})

	kafkatest.TestRequest(t, 5, &syncGroup.Request{
		GroupId:         "foo",
		GenerationId:    1,
		MemberId:        "m1",
		GroupInstanceId: "g1",
		ProtocolType:    "proto",
		ProtocolName:    "p1",
		GroupAssignments: []syncGroup.GroupAssignment{
			{
				MemberId:   "m2",
				Assignment: []byte("assign"),
			},
		},
	})

	b := kafkatest.WriteRequest(t, 5, 123, "me", &syncGroup.Request{
		GroupId:         "foo",
		GenerationId:    1,
		MemberId:        "m1",
		GroupInstanceId: "g1",
		ProtocolType:    "proto",
		ProtocolName:    "p1",
		GroupAssignments: []syncGroup.GroupAssignment{
			{
				MemberId:   "m2",
				Assignment: []byte("assign"),
			},
		},
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(49))              // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.SyncGroup)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(5))               // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))             // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))               // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))           // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))                // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int8(4))          // GroupId length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo"))    // GroupId
	_ = binary.Write(expected, binary.BigEndian, int32(1))         // GenerationId
	_ = binary.Write(expected, binary.BigEndian, int8(3))          // MemberId length
	_ = binary.Write(expected, binary.BigEndian, []byte("m1"))     // MemberId
	_ = binary.Write(expected, binary.BigEndian, int8(3))          // GroupInstanceId length
	_ = binary.Write(expected, binary.BigEndian, []byte("g1"))     // GroupInstanceId
	_ = binary.Write(expected, binary.BigEndian, int8(6))          // ProtocolType length
	_ = binary.Write(expected, binary.BigEndian, []byte("proto"))  // ProtocolType
	_ = binary.Write(expected, binary.BigEndian, int8(3))          // ProtocolName length
	_ = binary.Write(expected, binary.BigEndian, []byte("p1"))     // ProtocolName
	_ = binary.Write(expected, binary.BigEndian, int8(2))          // GroupAssignments length
	_ = binary.Write(expected, binary.BigEndian, int8(3))          // MemberId length
	_ = binary.Write(expected, binary.BigEndian, []byte("m2"))     // MemberId
	_ = binary.Write(expected, binary.BigEndian, int8(7))          // Assignment length
	_ = binary.Write(expected, binary.BigEndian, []byte("assign")) // Assignment
	_ = binary.Write(expected, binary.BigEndian, int8(0))          // GroupAssignments tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))          // tag buffer
	require.Equal(t, expected.Bytes(), b)
}

func TestResponse(t *testing.T) {
	kafkatest.TestResponse(t, 3, &syncGroup.Response{
		ThrottleTimeMs: 123,
		ErrorCode:      0,
		Assignment:     []byte("assign"),
	})

	kafkatest.TestResponse(t, 5, &syncGroup.Response{
		ThrottleTimeMs: 123,
		ErrorCode:      0,
		ProtocolType:   "proto",
		ProtocolName:   "p1",
		Assignment:     []byte("assign"),
	})

	b := kafkatest.WriteResponse(t, 5, 123, &syncGroup.Response{
		ThrottleTimeMs: 123,
		ErrorCode:      0,
		ProtocolType:   "proto",
		ProtocolName:   "p1",
		Assignment:     []byte("assign"),
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(28))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(123))       // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int16(0))         // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int8(6))          // ProtocolType length
	_ = binary.Write(expected, binary.BigEndian, []byte("proto"))  // ProtocolType
	_ = binary.Write(expected, binary.BigEndian, int8(3))          // ProtocolType length
	_ = binary.Write(expected, binary.BigEndian, []byte("p1"))     // ProtocolType
	_ = binary.Write(expected, binary.BigEndian, int8(7))          // Assignment length
	_ = binary.Write(expected, binary.BigEndian, []byte("assign")) // Assignment
	_ = binary.Write(expected, binary.BigEndian, int8(0))          // tag buffer
	require.Equal(t, expected.Bytes(), b)
}
