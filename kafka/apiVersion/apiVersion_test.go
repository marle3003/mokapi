package apiVersion_test

import (
	"bytes"
	"encoding/binary"
	"mokapi/kafka"
	"mokapi/kafka/apiVersion"
	"mokapi/kafka/kafkatest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.ApiVersions]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(3), reg.MaxVersion)
}

func TestNewApiKeyResponse(t *testing.T) {
	res := apiVersion.NewApiKeyResponse(kafka.ApiVersions, kafka.ApiTypes[kafka.ApiVersions])
	require.Equal(t, kafka.ApiVersions, res.ApiKey)
	require.Equal(t, int16(0), res.MinVersion)
	require.Equal(t, int16(3), res.MaxVersion)
}

func TestRequest(t *testing.T) {
	kafkatest.TestRequest(t, 2, &apiVersion.Request{})

	kafkatest.TestRequest(t, 3, &apiVersion.Request{
		ClientSwName:    "foo",
		ClientSwVersion: "1.1",
	})

	b := kafkatest.WriteRequest(t, 3, 123, "me", &apiVersion.Request{
		ClientSwName:    "foo",
		ClientSwVersion: "1.1",
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(22))                // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.ApiVersions)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(3))                 // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))               // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))                 // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))             // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))                  // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Key length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // Key
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Key length
	_ = binary.Write(expected, binary.BigEndian, []byte("1.1")) // Key
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	require.Equal(t, expected.Bytes(), b)
}

func TestResponse(t *testing.T) {
	kafkatest.TestResponse(t, 2, &apiVersion.Response{
		ErrorCode: 0,
		ApiKeys: []apiVersion.ApiKeyResponse{
			{
				ApiKey:     kafka.Produce,
				MinVersion: 0,
				MaxVersion: 3,
			},
		},
		ThrottleTimeMs: 123,
	})

	kafkatest.TestResponse(t, 3, &apiVersion.Response{
		ErrorCode: 0,
		ApiKeys: []apiVersion.ApiKeyResponse{
			{
				ApiKey:     kafka.Produce,
				MinVersion: 0,
				MaxVersion: 3,
			},
		},
		ThrottleTimeMs: 123,
	})

	b := kafkatest.WriteResponse(t, 3, 123, &apiVersion.Response{
		ErrorCode: 0,
		ApiKeys: []apiVersion.ApiKeyResponse{
			{
				ApiKey:     kafka.Produce,
				MinVersion: 0,
				MaxVersion: 3,
			},
		},
		ThrottleTimeMs: 123,
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(19))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	// no tag in header
	// message
	_ = binary.Write(expected, binary.BigEndian, int16(0))             // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int8(2))              // Array length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.Produce)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(0))             // MinVersion
	_ = binary.Write(expected, binary.BigEndian, int16(3))             // MaxVersion
	_ = binary.Write(expected, binary.BigEndian, int8(0))              // tag buffer
	_ = binary.Write(expected, binary.BigEndian, int32(123))           // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int8(0))              // tag buffer
	require.Equal(t, expected.Bytes(), b)
}
