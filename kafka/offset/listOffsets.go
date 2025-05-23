package offset

import (
	"mokapi/kafka"
)

func init() {
	kafka.Register(
		kafka.ApiReg{
			ApiKey:     kafka.Offset,
			MinVersion: 0,
			MaxVersion: 8},
		&Request{},
		&Response{},
		6,
		6,
	)
}

type Request struct {
	ReplicaId      int32            `kafka:""`
	IsolationLevel int8             `kafka:"min=2"`
	Topics         []RequestTopic   `kafka:"compact=6"`
	TagFields      map[int64]string `kafka:"type=TAG_BUFFER,min=6"`
}

type RequestTopic struct {
	Name       string             `kafka:"compact=6"`
	Partitions []RequestPartition `kafka:"compact=6"`
	TagFields  map[int64]string   `kafka:"type=TAG_BUFFER,min=6"`
}

type RequestPartition struct {
	Index         int32            `kafka:""`
	LeaderEpoch   int32            `kafka:"min=4"`
	Timestamp     int64            `kafka:""`
	MaxNumOffsets int32            `kafka:"max=0"`
	TagFields     map[int64]string `kafka:"type=TAG_BUFFER,min=6"`
}

type Response struct {
	ThrottleTimeMs int32            `kafka:"min=2"`
	Topics         []ResponseTopic  `kafka:"compact=6"`
	TagFields      map[int64]string `kafka:"type=TAG_BUFFER,min=6"`
}

type ResponseTopic struct {
	Name       string              `kafka:"compact=6"`
	Partitions []ResponsePartition `kafka:"compact=6"`
	TagFields  map[int64]string    `kafka:"type=TAG_BUFFER,min=6"`
}

type ResponsePartition struct {
	Index           int32            `kafka:""`
	ErrorCode       kafka.ErrorCode  `kafka:""`
	Timestamp       int64            `kafka:"min=1"`
	Offset          int64            `kafka:"min=1"`
	LeaderEpoch     int32            `kafka:"min=4"`
	OldStyleOffsets []int64          `kafka:"max=0"`
	TagFields       map[int64]string `kafka:"type=TAG_BUFFER,min=6"`
}
