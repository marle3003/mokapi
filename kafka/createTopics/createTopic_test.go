package createTopics_test

import (
	"bytes"
	"encoding/binary"
	"mokapi/kafka"
	"mokapi/kafka/createTopics"
	"mokapi/kafka/kafkatest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.CreateTopics]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(7), reg.MaxVersion)
}

func TestRequest(t *testing.T) {
	kafkatest.TestRequest(t, 4, &createTopics.Request{
		Topics: []createTopics.Topic{
			{
				Name:              "foo",
				NumPartitions:     2,
				ReplicationFactor: 1,
				Assignments: []createTopics.Assignment{
					{
						Index:     0,
						BrokerIds: []int32{0, 1},
					},
				},
				Configs: []createTopics.Config{
					{
						Name:  "foo",
						Value: "bar",
					},
				},
			},
		},
		TimeoutMs:    123,
		ValidateOnly: false,
	})

	kafkatest.TestRequest(t, 5, &createTopics.Request{
		Topics: []createTopics.Topic{
			{
				Name:              "foo",
				NumPartitions:     2,
				ReplicationFactor: 1,
				Assignments: []createTopics.Assignment{
					{
						Index:     0,
						BrokerIds: []int32{0, 1},
					},
				},
				Configs: []createTopics.Config{
					{
						Name:  "foo",
						Value: "bar",
					},
				},
			},
		},
		TimeoutMs:    123,
		ValidateOnly: false,
	})

	b := kafkatest.WriteRequest(t, 5, 123, "me", &createTopics.Request{
		Topics: []createTopics.Topic{
			{
				Name:              "foo",
				NumPartitions:     2,
				ReplicationFactor: 1,
				Assignments: []createTopics.Assignment{
					{
						Index:     0,
						BrokerIds: []int32{0, 1},
					},
				},
				Configs: []createTopics.Config{
					{
						Name:  "foo",
						Value: "bar",
					},
				},
			},
		},
		TimeoutMs:    123,
		ValidateOnly: false,
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(56))                 // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.CreateTopics)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(5))                  // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))                // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))                  // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))              // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))                   // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Topic name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // Topic name
	_ = binary.Write(expected, binary.BigEndian, int32(2))      // NumPartitions
	_ = binary.Write(expected, binary.BigEndian, int16(1))      // ReplicationFactor
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Assignments length
	_ = binary.Write(expected, binary.BigEndian, int32(0))      // Index
	_ = binary.Write(expected, binary.BigEndian, int8(3))       // BrokerIds length
	_ = binary.Write(expected, binary.BigEndian, int32(0))      // BrokerId 0
	_ = binary.Write(expected, binary.BigEndian, int32(1))      // BrokerId 1
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Configs length
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // Name
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Value length
	_ = binary.Write(expected, binary.BigEndian, []byte("bar")) // Value
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int32(123))    // TimeoutMs
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // ValidateOnly
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	require.Equal(t, expected.Bytes(), b)
}

func TestResponse(t *testing.T) {
	kafkatest.TestResponse(t, 4, &createTopics.Response{
		ThrottleTimeMs: 123,
		Topics: []createTopics.TopicResponse{
			{
				Name:         "foo",
				ErrorCode:    0,
				ErrorMessage: "",
			},
		},
	})

	kafkatest.TestResponse(t, 5, &createTopics.Response{
		ThrottleTimeMs: 123,
		Topics: []createTopics.TopicResponse{
			{
				Name:              "foo",
				ErrorCode:         0,
				ErrorMessage:      "",
				NumPartitions:     2,
				ReplicationFactor: 1,
				Configs: []createTopics.ConfigResponse{
					{
						Name:         "foo",
						Value:        "bar",
						ReadOnly:     false,
						ConfigSource: 0,
						IsSensitive:  false,
					},
				},
			},
		},
	})

	b := kafkatest.WriteResponse(t, 5, 123, &createTopics.Response{
		ThrottleTimeMs: 123,
		Topics: []createTopics.TopicResponse{
			{
				Name:              "foo",
				ErrorCode:         0,
				ErrorMessage:      "",
				NumPartitions:     2,
				ReplicationFactor: 1,
				Configs: []createTopics.ConfigResponse{
					{
						Name:         "foo",
						Value:        "bar",
						ReadOnly:     false,
						ConfigSource: 0,
						IsSensitive:  false,
					},
				},
			},
		},
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(38))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(123))    // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Topic name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // Topic name
	_ = binary.Write(expected, binary.BigEndian, int16(0))      // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // ErrorMessage length
	_ = binary.Write(expected, binary.BigEndian, int32(2))      // NumPartitions
	_ = binary.Write(expected, binary.BigEndian, int16(1))      // ReplicationFactor
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Configs length
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // Name
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Value length
	_ = binary.Write(expected, binary.BigEndian, []byte("bar")) // Value
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // ReadOnly
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // ConfigSource
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // IsSensitive
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // Configs tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	require.Equal(t, expected.Bytes(), b)
}
