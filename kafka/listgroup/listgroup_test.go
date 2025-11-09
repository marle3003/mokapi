package listgroup_test

import (
	"bytes"
	"encoding/binary"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/listgroup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.ListGroup]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(4), reg.MaxVersion)
}

func TestRequest(t *testing.T) {
	kafkatest.TestRequest(t, 2, &listgroup.Request{})

	kafkatest.TestRequest(t, 4, &listgroup.Request{
		StatesFilter: []string{
			"foo",
		},
	})

	b := kafkatest.WriteRequest(t, 4, 123, "me", &listgroup.Request{
		StatesFilter: []string{
			"foo",
		},
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(19))              // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.ListGroup)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(4))               // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))             // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))               // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))           // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))                // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // StatesFilters length
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // StatesFilter length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // Key
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	require.Equal(t, expected.Bytes(), b)
}

func TestResponse(t *testing.T) {
	kafkatest.TestResponse(t, 2, &listgroup.Response{
		ThrottleTimeMs: 0,
		ErrorCode:      0,
		Groups: []listgroup.Group{
			{
				GroupId:      "g-1",
				ProtocolType: "proto",
			},
		},
	})

	kafkatest.TestResponse(t, 4, &listgroup.Response{
		ThrottleTimeMs: 0,
		ErrorCode:      0,
		Groups: []listgroup.Group{
			{
				GroupId:      "g-1",
				ProtocolType: "proto",
				GroupState:   "state",
			},
		},
	})

	b := kafkatest.WriteResponse(t, 4, 123, &listgroup.Response{
		ThrottleTimeMs: 123,
		ErrorCode:      0,
		Groups: []listgroup.Group{
			{
				GroupId:      "g-1",
				ProtocolType: "proto",
				GroupState:   "state",
			},
		},
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(30))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(123))      // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int16(0))        // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int8(2))         // Groups length
	_ = binary.Write(expected, binary.BigEndian, int8(4))         // GroupId length
	_ = binary.Write(expected, binary.BigEndian, []byte("g-1"))   // GroupId
	_ = binary.Write(expected, binary.BigEndian, int8(6))         // ProtocolType length
	_ = binary.Write(expected, binary.BigEndian, []byte("proto")) // ProtocolType
	_ = binary.Write(expected, binary.BigEndian, int8(6))         // GroupState length
	_ = binary.Write(expected, binary.BigEndian, []byte("state")) // GroupState
	_ = binary.Write(expected, binary.BigEndian, int8(0))         // Groups tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))         // tag buffer
	require.Equal(t, expected.Bytes(), b)
}
