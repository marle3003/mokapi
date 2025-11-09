package offsetCommit

import (
	"mokapi/kafka"
)

func init() {
	kafka.Register(
		kafka.ApiReg{
			ApiKey:     kafka.OffsetCommit,
			MinVersion: 0,
			MaxVersion: 9},
		&Request{},
		&Response{},
		8,
		8,
	)
}

type Request struct {
	GroupId         string           `kafka:"compact=8"`
	GenerationId    int32            `kafka:"min=1"`
	MemberId        string           `kafka:"min=1,compact=8"`
	GroupInstanceId string           `kafka:"min=7,compact=8,nullable"`
	RetentionTime   int64            `kafka:"min=2,max=4"`
	Topics          []Topic          `kafka:"compact=8"`
	TagFields       map[int64]string `kafka:"type=TAG_BUFFER,min=8"`
}

type Topic struct {
	Name       string           `kafka:"compact=8"`
	Partitions []Partition      `kafka:"compact=8"`
	TagFields  map[int64]string `kafka:"type=TAG_BUFFER,min=8"`
}

type Partition struct {
	Index       int32            `kafka:""`
	Offset      int64            `kafka:""`
	LeaderEpoch int32            `kafka:"min=6"`
	Timestamp   int64            `kafka:"min=1,max=1"`
	Metadata    string           `kafka:"nullable,compact=8"`
	TagFields   map[int64]string `kafka:"type=TAG_BUFFER,min=8"`
}

type Response struct {
	ThrottleTimeMs int32            `kafka:"min=3"`
	Topics         []ResponseTopic  `kafka:"compact=8"`
	TagFields      map[int64]string `kafka:"type=TAG_BUFFER,min=8"`
}

type ResponseTopic struct {
	Name       string              `kafka:"compact=8"`
	Partitions []ResponsePartition `kafka:"compact=8"`
	TagFields  map[int64]string    `kafka:"type=TAG_BUFFER,min=8"`
}

type ResponsePartition struct {
	Index     int32            `kafka:""`
	ErrorCode kafka.ErrorCode  `kafka:""`
	TagFields map[int64]string `kafka:"type=TAG_BUFFER,min=8"`
}
