package findCoordinator_test

import (
	"bytes"
	"encoding/binary"
	"mokapi/kafka"
	"mokapi/kafka/findCoordinator"
	"mokapi/kafka/kafkatest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.FindCoordinator]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(3), reg.MaxVersion)
}

func TestRequest(t *testing.T) {
	kafkatest.TestRequest(t, 2, &findCoordinator.Request{
		Key:     "foo-group",
		KeyType: findCoordinator.KeyTypeGroup,
	})

	kafkatest.TestRequest(t, 3, &findCoordinator.Request{
		Key:     "foo-group",
		KeyType: findCoordinator.KeyTypeGroup,
	})

	b := kafkatest.WriteRequest(t, 3, 123, "me", &findCoordinator.Request{
		Key:     "foo-group",
		KeyType: findCoordinator.KeyTypeGroup,
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(25))                    // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.FindCoordinator)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(3))                     // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))                   // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))                     // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))                 // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))                      // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int8(10))                           // Key length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo-group"))                // Key
	_ = binary.Write(expected, binary.BigEndian, int8(findCoordinator.KeyTypeGroup)) // KeyType
	_ = binary.Write(expected, binary.BigEndian, int8(0))                            // tag buffer
	require.Equal(t, expected.Bytes(), b)
}

func TestResponse(t *testing.T) {
	kafkatest.TestResponse(t, 2, &findCoordinator.Response{
		ThrottleTimeMs: 123,
		ErrorCode:      0,
		ErrorMessage:   "",
		NodeId:         1,
		Port:           1234,
	})

	kafkatest.TestResponse(t, 3, &findCoordinator.Response{
		ThrottleTimeMs: 123,
		ErrorCode:      0,
		ErrorMessage:   "",
		NodeId:         1,
		Host:           "foo",
		Port:           1234,
	})

	b := kafkatest.WriteResponse(t, 3, 123, &findCoordinator.Response{
		ThrottleTimeMs: 123,
		ErrorCode:      0,
		ErrorMessage:   "",
		NodeId:         1,
		Host:           "foo",
		Port:           1234,
	})

	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(25))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(123))    // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int16(0))      // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // ErrorMessage length
	_ = binary.Write(expected, binary.BigEndian, int32(1))      // NodeId
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Host length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // Host
	_ = binary.Write(expected, binary.BigEndian, int32(1234))   // Port
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	require.Equal(t, expected.Bytes(), b)
}
