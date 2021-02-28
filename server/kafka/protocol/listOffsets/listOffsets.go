package listOffsets

import "mokapi/server/kafka/protocol"

func init() {
	protocol.Register(
		protocol.ApiReg{
			ApiKey:     protocol.ListOffsets,
			MinVersion: 0,
			MaxVersion: 9},
		&Request{},
		&Response{},
		9,
		9,
	)
}

type Request struct {
	ReplicaId      int32          `kafka:""`
	IsolationLevel int8           `kafka:"min=2"`
	Topics         []RequestTopic `kafka:""`
}

type RequestTopic struct {
	Name       string             `kafka:""`
	Partitions []RequestPartition `kafka:""`
}

type RequestPartition struct {
	Index         int32 `kafka:""`
	LeaderEpoch   int32 `kafka:"min=4"`
	Timestamp     int64 `kafka:""`
	MaxNumOffsets int32 `kafka:"max=0"`
}

type Response struct {
	ThrottleTimeMs int32           `kafka:"min=2"`
	Topics         []ResponseTopic `kafka:""`
}

type ResponseTopic struct {
	Name       string              `kafka:""`
	Partitions []ResponsePartition `kafka:""`
}

type ResponsePartition struct {
	Index           int32 `kafka:""`
	ErrorCode       int16 `kafka:""`
	Timestamp       int64 `kafka:"min=1"`
	Offset          int64 `kafka:"min=1"`
	LeaderEpoch     int32 `kafka:"min=4"`
	OldStyleOffsets int64 `kafka:"max=0"`
}
