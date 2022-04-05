package metaData

import (
	"mokapi/kafka"
)

func init() {
	kafka.Register(
		kafka.ApiReg{
			ApiKey:     kafka.Metadata,
			MinVersion: 0,
			MaxVersion: 9},
		&Request{},
		&Response{},
		9,
		9,
	)
}

type Request struct {
	Topics                             []TopicName      `kafka:"compact=9"`
	AllowAutoTopicCreation             bool             `kafka:"min=4"`
	IncludeClusterAuthorizedOperations bool             `kafka:"min=8"`
	IncludeTopicAuthorizedOperations   bool             `kafka:"min=8"`
	TagFields                          map[int64]string `kafka:"type=TAG_BUFFER,min=9"`
}

type TopicName struct {
	Name      string           `kafka:"compact=9"`
	TagFields map[int64]string `kafka:"type=TAG_BUFFER,min=9"`
}

type Response struct {
	ThrottleTimeMs              int32            `kafka:"min=3"`
	Brokers                     []ResponseBroker `kafka:"compact=9"`
	ClusterId                   string           `kafka:"min=2,compact=9,nullable"`
	ControllerId                int32            `kafka:"min=1"`
	Topics                      []ResponseTopic  `kafka:"compact=9"`
	ClusterAuthorizedOperations int32            `kafka:"min=8"`
	TagFields                   map[int64]string `kafka:"type=TAG_BUFFER,min=9"`
}

type ResponseBroker struct {
	NodeId    int32            `kafka:""`
	Host      string           `kafka:"compact=9"`
	Port      int32            `kafka:""`
	Rack      string           `kafka:"min=1,compact=9,nullable"`
	TagFields map[int64]string `kafka:"type=TAG_BUFFER,min=9"`
}

type ResponseTopic struct {
	ErrorCode                 kafka.ErrorCode     `kafka:""`
	Name                      string              `kafka:"compact=9"`
	IsInternal                bool                `kafka:"min=1"`
	Partitions                []ResponsePartition `kafka:"compact=9"`
	TopicAuthorizedOperations int32               `kafka:"min=8"`
	TagFields                 map[int64]string    `kafka:"type=TAG_BUFFER,min=9"`
}

type ResponsePartition struct {
	ErrorCode       int16            `kafka:""`
	PartitionIndex  int32            `kafka:""`
	LeaderId        int32            `kafka:""`
	LeaderEpoch     int32            `kafka:"min=7"`
	ReplicaNodes    []int32          `kafka:"compact=9"`
	IsrNodes        []int32          `kafka:"compact=9"`
	OfflineReplicas []int32          `kafka:"min=5,compact=9"`
	TagFields       map[int64]string `kafka:"type=TAG_BUFFER,min=9"`
}
