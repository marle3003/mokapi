package metaData

import "mokapi/server/kafka/protocol"

func init() {
	protocol.Register(
		protocol.ApiReg{
			ApiKey:     protocol.Metadata,
			MinVersion: 1,
			MaxVersion: 9},
		&Request{},
		&Response{},
	)
}

type Request struct {
	Topics                             []TopicName `kafka:"min=9"`
	AllowAutoTopicCreation             bool        `kafka:"min=4"`
	IncludeClusterAuthorizedOperations bool        `kafka:"min=8"`
	IncludeTopicAuthorizedOperations   bool        `kafka:"min=8"`
}

type TopicName struct {
	Name      string           `kafka:"compact=9"`
	TagFields map[int64]string `kafka:"type=TAG_BUFFER,min=9"`
}

type Response struct {
	ThrottleTimeMs              int32            `kafka:"min=3"`
	Brokers                     []ResponseBroker `kafka:""`
	ClusterId                   string           `kafka:"min=2,compact=9,nullable"`
	ControllerId                int32            `kafka:"min=1"`
	Topics                      []ResponseTopic  `kafka:""`
	ClusterAuthorizedOperations int32            `kafka:"min=8"`
}

type ResponseBroker struct {
	NodeId    int32            `kafka:""`
	Host      string           `kafka:"compact=9"`
	Port      int32            `kafka:""`
	Rack      string           `kafka:"min=1,compact=9,nullable"`
	TagFields map[int64]string `kafka:"type=TAG_BUFFER,min=9"`
}

type ResponseTopic struct {
	ErrorCode                 int16               `kafka:""`
	Name                      string              `kafka:"compact=9"`
	IsInternal                bool                `kafka:"min=1"`
	Partitions                []ResponsePartition `kafka:""`
	TopicAuthorizedOperations int32               `kafka:"min=8"`
	TagFields                 map[int64]string    `kafka:"type=TAG_BUFFER,min=9"`
}

type ResponsePartition struct {
	ErrorCode       int16            `kafka:""`
	PartitionIndex  int32            `kafka:""`
	LeaderId        int32            `kafka:""`
	LeaderEpoch     int32            `kafka:"min=7"`
	ReplicaNodes    []int32          `kafka:""`
	IsrNodes        []int32          `kafka:""`
	OfflineReplicas []int32          `kafka:"min=5"`
	TagFields       map[int64]string `kafka:"type=TAG_BUFFER,min=9"`
}
