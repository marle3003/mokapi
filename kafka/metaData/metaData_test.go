package metaData_test

import (
	"bytes"
	"encoding/binary"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/metaData"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.Metadata]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(9), reg.MaxVersion)
}

func TestRequest(t *testing.T) {
	kafkatest.TestRequest(t, 8, &metaData.Request{
		Topics: []metaData.TopicName{
			{Name: "foo"},
		},
		AllowAutoTopicCreation:             true,
		IncludeClusterAuthorizedOperations: false,
		IncludeTopicAuthorizedOperations:   false,
	})

	kafkatest.TestRequest(t, 9, &metaData.Request{
		Topics: []metaData.TopicName{
			{Name: "foo"},
		},
		AllowAutoTopicCreation:             true,
		IncludeClusterAuthorizedOperations: false,
		IncludeTopicAuthorizedOperations:   false,
	})

	b := kafkatest.WriteRequest(t, 9, 123, "me", &metaData.Request{
		Topics: []metaData.TopicName{
			{Name: "foo"},
		},
		AllowAutoTopicCreation:             true,
		IncludeClusterAuthorizedOperations: false,
		IncludeTopicAuthorizedOperations:   false,
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(23))             // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.Metadata)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(9))              // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))            // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))              // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))          // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))               // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // Name
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(1))       // AllowAutoTopicCreation
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // IncludeClusterAuthorizedOperations
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // IncludeTopicAuthorizedOperations
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	require.Equal(t, expected.Bytes(), b)
}

func TestResponse(t *testing.T) {
	kafkatest.TestResponse(t, 8, &metaData.Response{
		ThrottleTimeMs: 123,
		Brokers: []metaData.ResponseBroker{
			{
				NodeId: 1,
				Host:   "localhost",
				Port:   9092,
				Rack:   "",
			},
		},
		ClusterId:    "foo",
		ControllerId: 1,
		Topics: []metaData.ResponseTopic{
			{
				ErrorCode:  0,
				Name:       "bar",
				IsInternal: false,
				Partitions: []metaData.ResponsePartition{
					{
						ErrorCode:       0,
						PartitionIndex:  1,
						LeaderId:        0,
						LeaderEpoch:     0,
						ReplicaNodes:    []int32{0, 1},
						IsrNodes:        []int32{},
						OfflineReplicas: []int32{},
					},
				},
				TopicAuthorizedOperations: 0,
			},
		},
		ClusterAuthorizedOperations: 0,
	})

	kafkatest.TestResponse(t, 9, &metaData.Response{
		ThrottleTimeMs: 123,
		Brokers: []metaData.ResponseBroker{
			{
				NodeId: 1,
				Host:   "localhost",
				Port:   9092,
				Rack:   "",
			},
		},
		ClusterId:    "foo",
		ControllerId: 1,
		Topics: []metaData.ResponseTopic{
			{
				ErrorCode:  0,
				Name:       "bar",
				IsInternal: false,
				Partitions: []metaData.ResponsePartition{
					{
						ErrorCode:       0,
						PartitionIndex:  1,
						LeaderId:        0,
						LeaderEpoch:     0,
						ReplicaNodes:    []int32{0, 1},
						IsrNodes:        []int32{},
						OfflineReplicas: []int32{},
					},
				},
				TopicAuthorizedOperations: 0,
			},
		},
		ClusterAuthorizedOperations: 0,
	})

	b := kafkatest.WriteResponse(t, 9, 123, &metaData.Response{
		ThrottleTimeMs: 123,
		Brokers: []metaData.ResponseBroker{
			{
				NodeId: 1,
				Host:   "localhost",
				Port:   9092,
				Rack:   "",
			},
		},
		ClusterId:    "foo",
		ControllerId: 1,
		Topics: []metaData.ResponseTopic{
			{
				ErrorCode:  0,
				Name:       "bar",
				IsInternal: false,
				Partitions: []metaData.ResponsePartition{
					{
						ErrorCode:       0,
						PartitionIndex:  1,
						LeaderId:        0,
						LeaderEpoch:     0,
						ReplicaNodes:    []int32{0, 1},
						IsrNodes:        []int32{},
						OfflineReplicas: []int32{},
					},
				},
				TopicAuthorizedOperations: 0,
			},
		},
		ClusterAuthorizedOperations: 0,
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(83))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(123))          // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int8(2))             // Brokers length
	_ = binary.Write(expected, binary.BigEndian, int32(1))            // NodeId
	_ = binary.Write(expected, binary.BigEndian, int8(10))            // Host length
	_ = binary.Write(expected, binary.BigEndian, []byte("localhost")) // Host
	_ = binary.Write(expected, binary.BigEndian, int32(9092))         // Port
	_ = binary.Write(expected, binary.BigEndian, int8(0))             // Rack length
	_ = binary.Write(expected, binary.BigEndian, int8(0))             // Brokers tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(4))             // ClusterId length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo"))       // ClusterId
	_ = binary.Write(expected, binary.BigEndian, int32(1))            // ControllerId
	_ = binary.Write(expected, binary.BigEndian, int8(2))             // Topics length
	_ = binary.Write(expected, binary.BigEndian, int16(0))            // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int8(4))             // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("bar"))       // Name
	_ = binary.Write(expected, binary.BigEndian, int8(0))             // IsInternal
	_ = binary.Write(expected, binary.BigEndian, int8(2))             // Partitions length
	_ = binary.Write(expected, binary.BigEndian, int16(0))            // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int32(1))            // PartitionIndex
	_ = binary.Write(expected, binary.BigEndian, int32(0))            // LeaderId
	_ = binary.Write(expected, binary.BigEndian, int32(0))            // LeaderEpoch
	_ = binary.Write(expected, binary.BigEndian, int8(3))             // ReplicaNodes length
	_ = binary.Write(expected, binary.BigEndian, int32(0))            // ReplicaNode 0
	_ = binary.Write(expected, binary.BigEndian, int32(1))            // ReplicaNode 1
	_ = binary.Write(expected, binary.BigEndian, int8(1))             // IsrNodes length
	_ = binary.Write(expected, binary.BigEndian, int8(1))             // OfflineReplicas length
	_ = binary.Write(expected, binary.BigEndian, int8(0))             // Partitions tag buffer
	_ = binary.Write(expected, binary.BigEndian, int32(0))            // TopicAuthorizedOperations
	_ = binary.Write(expected, binary.BigEndian, int8(0))             // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int32(0))            // ClusterAuthorizedOperations
	_ = binary.Write(expected, binary.BigEndian, int8(0))             // tag buffer
	require.Equal(t, expected.Bytes(), b)
}
