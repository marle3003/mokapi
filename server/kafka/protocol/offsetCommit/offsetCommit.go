package offsetCommit

import (
	"math"
	"mokapi/server/kafka/protocol"
)

func init() {
	protocol.Register(
		protocol.ApiReg{
			ApiKey:     protocol.OffsetCommit,
			MinVersion: 0,
			MaxVersion: 2},
		&Request{},
		&Response{},
		math.MaxInt16,
		math.MaxInt16,
	)
}

type Request struct {
	GroupId       string  `kafka:""`
	GenerationId  int32   `kafka:"min=1"`
	MemberId      string  `kafka:"min=1"`
	RetentionTime int64   `kafka:"min=2"`
	Topics        []Topic `kafka:""`
}

type Topic struct {
	Name       string      `kafka:""`
	Partitions []Partition `kafka:""`
}

type Partition struct {
	Index     int32  `kafka:""`
	Offset    int64  `kafka:""`
	Timestamp int64  `kafka:"min=1,max=1"`
	Metadata  string `kafka:"nullable"`
}

type Response struct {
	Topics []ResponseTopic `kafka:""`
}

type ResponseTopic struct {
	Name       string              `kafka:""`
	Partitions []ResponsePartition `kafka:""`
}

type ResponsePartition struct {
	Index     int32 `kafka:""`
	ErrorCode int16 `kafka:""`
}
