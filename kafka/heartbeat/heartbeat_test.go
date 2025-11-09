package heartbeat_test

import (
	"bytes"
	"encoding/binary"
	"mokapi/kafka"
	"mokapi/kafka/heartbeat"
	"mokapi/kafka/kafkatest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.Heartbeat]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(4), reg.MaxVersion)
}

func TestRequest(t *testing.T) {
	kafkatest.TestRequest(t, 3, &heartbeat.Request{
		GroupId:         "foo",
		GenerationId:    1,
		MemberId:        "m1",
		GroupInstanceId: "g1",
	})

	kafkatest.TestRequest(t, 4, &heartbeat.Request{
		GroupId:         "foo",
		GenerationId:    1,
		MemberId:        "m1",
		GroupInstanceId: "g1",
	})

	b := kafkatest.WriteRequest(t, 4, 123, "me", &heartbeat.Request{
		GroupId:         "foo",
		GenerationId:    1,
		MemberId:        "m1",
		GroupInstanceId: "g1",
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(28))              // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.Heartbeat)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(4))               // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))             // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))               // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))           // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))                // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // GroupId length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // GroupId
	_ = binary.Write(expected, binary.BigEndian, int32(1))      // GenerationId
	_ = binary.Write(expected, binary.BigEndian, int8(3))       // MemberId length
	_ = binary.Write(expected, binary.BigEndian, []byte("m1"))  // MemberId
	_ = binary.Write(expected, binary.BigEndian, int8(3))       // GroupInstanceId length
	_ = binary.Write(expected, binary.BigEndian, []byte("g1"))  // GroupInstanceId
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	require.Equal(t, expected.Bytes(), b)
}

func TestResponse(t *testing.T) {
	kafkatest.TestResponse(t, 3, &heartbeat.Response{
		ThrottleTimeMs: 123,
		ErrorCode:      0,
	})

	kafkatest.TestResponse(t, 4, &heartbeat.Response{
		ThrottleTimeMs: 123,
		ErrorCode:      0,
	})

	b := kafkatest.WriteResponse(t, 4, 123, &heartbeat.Response{
		ThrottleTimeMs: 123,
		ErrorCode:      0,
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(12))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int16(0))   // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	require.Equal(t, expected.Bytes(), b)
}
