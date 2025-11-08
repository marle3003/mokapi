package joinGroup_test

import (
	"bytes"
	"encoding/binary"
	"mokapi/kafka"
	"mokapi/kafka/joinGroup"
	"mokapi/kafka/kafkatest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.JoinGroup]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(7), reg.MaxVersion)
}

func TestRequest(t *testing.T) {
	kafkatest.TestRequest(t, 5, &joinGroup.Request{
		GroupId:            "foo",
		SessionTimeoutMs:   0,
		RebalanceTimeoutMs: 0,
		MemberId:           "m1",
		GroupInstanceId:    "g1",
		ProtocolType:       "proto",
		Protocols: []joinGroup.Protocol{
			{
				Name:     "p1",
				MetaData: []byte("metadata"),
			},
		},
	})

	kafkatest.TestRequest(t, 6, &joinGroup.Request{
		GroupId:            "foo",
		SessionTimeoutMs:   0,
		RebalanceTimeoutMs: 0,
		MemberId:           "m1",
		GroupInstanceId:    "g1",
		ProtocolType:       "proto",
		Protocols: []joinGroup.Protocol{
			{
				Name:     "p1",
				MetaData: []byte("metadata"),
			},
		},
	})

	b := kafkatest.WriteRequest(t, 6, 123, "me", &joinGroup.Request{
		GroupId:            "foo",
		SessionTimeoutMs:   0,
		RebalanceTimeoutMs: 0,
		MemberId:           "m1",
		GroupInstanceId:    "g1",
		ProtocolType:       "proto",
		Protocols: []joinGroup.Protocol{
			{
				Name:     "p1",
				MetaData: []byte("metadata"),
			},
		},
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(52))              // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.JoinGroup)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(6))               // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))             // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))               // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))           // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))                // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int8(4))            // GroupId length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo"))      // GroupId
	_ = binary.Write(expected, binary.BigEndian, int32(0))           // SessionTimeoutMs
	_ = binary.Write(expected, binary.BigEndian, int32(0))           // RebalanceTimeoutMs
	_ = binary.Write(expected, binary.BigEndian, int8(3))            // MemberId length
	_ = binary.Write(expected, binary.BigEndian, []byte("m1"))       // MemberId
	_ = binary.Write(expected, binary.BigEndian, int8(3))            // GroupInstanceId length
	_ = binary.Write(expected, binary.BigEndian, []byte("g1"))       // GroupInstanceId
	_ = binary.Write(expected, binary.BigEndian, int8(6))            // ProtocolType length
	_ = binary.Write(expected, binary.BigEndian, []byte("proto"))    // ProtocolType
	_ = binary.Write(expected, binary.BigEndian, int8(2))            // Protocols length
	_ = binary.Write(expected, binary.BigEndian, int8(3))            // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("p1"))       // Name
	_ = binary.Write(expected, binary.BigEndian, int8(9))            // MetaData length
	_ = binary.Write(expected, binary.BigEndian, []byte("metadata")) // MetaData
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // Protocols tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // tag buffer
	require.Equal(t, expected.Bytes(), b)
}

func TestResponse(t *testing.T) {
	kafkatest.TestResponse(t, 5, &joinGroup.Response{
		ThrottleTimeMs: 123,
		ErrorCode:      0,
		GenerationId:   1,
		ProtocolName:   "p1",
		Leader:         "m2",
		MemberId:       "m1",
		Members: []joinGroup.Member{
			{
				MemberId:        "m1",
				GroupInstanceId: "g1",
				MetaData:        []byte("metadata"),
			},
		},
	})
	kafkatest.TestResponse(t, 7, &joinGroup.Response{
		ThrottleTimeMs: 123,
		ErrorCode:      0,
		GenerationId:   1,
		ProtocolType:   "proto",
		ProtocolName:   "p1",
		Leader:         "m2",
		MemberId:       "m1",
		Members: []joinGroup.Member{
			{
				MemberId:        "m1",
				GroupInstanceId: "g1",
				MetaData:        []byte("metadata"),
			},
		},
	})

	b := kafkatest.WriteResponse(t, 7, 123, &joinGroup.Response{
		ThrottleTimeMs: 123,
		ErrorCode:      0,
		GenerationId:   1,
		ProtocolType:   "proto",
		ProtocolName:   "p1",
		Leader:         "m2",
		MemberId:       "m1",
		Members: []joinGroup.Member{
			{
				MemberId:        "m1",
				GroupInstanceId: "g1",
				MetaData:        []byte("metadata"),
			},
		},
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(48))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(123))         // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int16(0))           // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int32(1))           // GenerationId
	_ = binary.Write(expected, binary.BigEndian, int8(6))            // ProtocolType length
	_ = binary.Write(expected, binary.BigEndian, []byte("proto"))    // ProtocolType
	_ = binary.Write(expected, binary.BigEndian, int8(3))            // ProtocolName length
	_ = binary.Write(expected, binary.BigEndian, []byte("p1"))       // ProtocolName
	_ = binary.Write(expected, binary.BigEndian, int8(3))            // Leader length
	_ = binary.Write(expected, binary.BigEndian, []byte("m2"))       // Leader
	_ = binary.Write(expected, binary.BigEndian, int8(3))            // MemberId length
	_ = binary.Write(expected, binary.BigEndian, []byte("m1"))       // MemberId
	_ = binary.Write(expected, binary.BigEndian, int8(2))            // Members length
	_ = binary.Write(expected, binary.BigEndian, int8(3))            // MemberId length
	_ = binary.Write(expected, binary.BigEndian, []byte("m1"))       // MemberId
	_ = binary.Write(expected, binary.BigEndian, int8(3))            // GroupInstanceId length
	_ = binary.Write(expected, binary.BigEndian, []byte("g1"))       // GroupInstanceId
	_ = binary.Write(expected, binary.BigEndian, int8(9))            // MetaData length
	_ = binary.Write(expected, binary.BigEndian, []byte("metadata")) // MetaData
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // Members tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // tag buffer
	require.Equal(t, expected.Bytes(), b)
}
